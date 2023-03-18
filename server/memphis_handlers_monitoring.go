// Copyright 2022-2023 The Memphis.dev Authors
// Licensed under the Memphis Business Source License 1.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// Changed License: [Apache License, Version 2.0 (https://www.apache.org/licenses/LICENSE-2.0), as published by the Apache Foundation.
//
// https://github.com/memphisdev/memphis/blob/master/LICENSE
//
// Additional Use Grant: You may make use of the Licensed Work (i) only as part of your own product or service, provided it is not a message broker or a message queue product or service; and (ii) provided that you do not use, provide, distribute, or make available the Licensed Work as a Service.
// A "Service" is a commercial offering, product, hosted, or managed service, that allows third parties (other than your own employees and contractors acting on your behalf) to access and/or use the Licensed Work or a substantial set of the features or functionality of the Licensed Work to third parties as a software-as-a-service, platform-as-a-service, infrastructure-as-a-service or other similar services that compete with Licensor products or services.
package server

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"math"
	"memphis/analytics"
	"memphis/conf"
	"memphis/db"
	"memphis/models"
	"memphis/utils"
	"net/http"
	"os"
	"os/exec"
	"regexp"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/docker/docker/api/types"
	dockerClient "github.com/docker/docker/client"
	"github.com/gin-gonic/gin"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/remotecommand"
	metricsv "k8s.io/metrics/pkg/client/clientset/versioned"
)

type MonitoringHandler struct{ S *Server }

var clientset *kubernetes.Clientset
var metricsclientset *metricsv.Clientset
var config *rest.Config
var noMetricsInstalledLog bool
var noMetricsPermissionLog bool

func clientSetClusterConfig() error {
	var err error
	// in cluster config
	config, err = rest.InClusterConfig()
	if err != nil {
		return err
	}
	clientset, err = kubernetes.NewForConfig(config)

	if err != nil {
		return err
	}
	if metricsclientset == nil {
		metricsclientset, err = metricsv.NewForConfig(config)
		if err != nil {
			return err
		}
	}

	noMetricsInstalledLog = false
	noMetricsPermissionLog = false

	return nil
}

func (mh MonitoringHandler) GetSystemComponents() ([]models.SystemComponents, bool, error) {
	components := []models.SystemComponents{}
	allComponents := []models.SysComponent{}
	portsMap := map[string][]int{}
	hosts := []string{}
	metricsEnabled := true
	defaultStat := models.CompStats{
		Total:      0,
		Current:    0,
		Percentage: 0,
	}
	if configuration.DOCKER_ENV == "true" { // docker env
		metricsEnabled = true
		hosts = []string{"localhost"}
		if configuration.DEV_ENV == "true" {
			maxCpu := float64(runtime.GOMAXPROCS(0))
			v, err := serv.Varz(nil)
			if err != nil {
				return components, metricsEnabled, err
			}
			var storageComp models.CompStats
			memUsage := float64(0)
			os := runtime.GOOS
			storage_size := float64(0)
			isWindows := false
			switch os {
			case "windows":
				isWindows = true
				storageComp = defaultStat // TODO: add support for windows
			default:
				storage_size, err = getUnixStorageSize()
				if err != nil {
					return components, metricsEnabled, err
				}
				storageComp = models.CompStats{
					Total:      shortenFloat(storage_size),
					Current:    shortenFloat(float64(v.JetStream.Stats.Store)),
					Percentage: int(math.Ceil((float64(v.JetStream.Stats.Store) / storage_size) * 100)),
				}
				memUsage, err = getUnixMemoryUsage()
				if err != nil {
					return components, metricsEnabled, err
				}
			}
			memPerc := (memUsage / float64(v.JetStream.Config.MaxMemory)) * 100
			cpuComps := []models.SysComponent{{
				Name: "memphis-0",
				CPU: models.CompStats{
					Total:      shortenFloat(maxCpu),
					Current:    shortenFloat((v.CPU / 100) * maxCpu),
					Percentage: int(math.Ceil(v.CPU)),
				},
				Memory: models.CompStats{
					Total:      shortenFloat(float64(v.JetStream.Config.MaxMemory)),
					Current:    shortenFloat(memUsage),
					Percentage: int(math.Ceil(memPerc)),
				},
				Storage: storageComp,
				Healthy: true,
			}}
			components = append(components, models.SystemComponents{
				Name:        "memphis",
				Components:  cpuComps,
				Status:      checkCompStatus(cpuComps),
				Ports:       []int{9000, 6666, 7770, 8222},
				DesiredPods: 1,
				ActualPods:  1,
				Hosts:       hosts,
			})
			resp, err := http.Get("http://localhost:4444/monitoring/getResourcesUtilization")
			healthy := false
			restGwComps := []models.SysComponent{defaultSystemComp("memphis-rest-gateway", healthy)}
			if err == nil {
				healthy = true
				var restGwMonitorInfo models.RestGwMonitoringResponse
				defer resp.Body.Close()
				err = json.NewDecoder(resp.Body).Decode(&restGwMonitorInfo)
				if err != nil {
					return components, metricsEnabled, err
				}
				if !isWindows {
					storageComp = models.CompStats{
						Total:      shortenFloat(storage_size),
						Current:    shortenFloat((restGwMonitorInfo.Storage / 100) * storage_size),
						Percentage: int(math.Ceil(float64(restGwMonitorInfo.Storage))),
					}
				}
				restGwComps = []models.SysComponent{{
					Name: "memphis-rest-gateway",
					CPU: models.CompStats{
						Total:      shortenFloat(maxCpu),
						Current:    shortenFloat((restGwMonitorInfo.CPU / 100) * maxCpu),
						Percentage: int(math.Ceil(restGwMonitorInfo.CPU)),
					},
					Memory: models.CompStats{
						Total:      shortenFloat(float64(v.JetStream.Config.MaxMemory)),
						Current:    shortenFloat((restGwMonitorInfo.Memory / 100) * float64(v.JetStream.Config.MaxMemory)),
						Percentage: int(math.Ceil(float64(restGwMonitorInfo.Memory))),
					},
					Storage: storageComp,
					Healthy: healthy,
				}}
			}
			actualRestGw := 1
			if !healthy {
				actualRestGw = 0
			}
			components = append(components, models.SystemComponents{
				Name:        "memphis-rest-gateway",
				Components:  restGwComps,
				Status:      checkCompStatus(restGwComps),
				Ports:       []int{4444},
				DesiredPods: 1,
				ActualPods:  actualRestGw,
				Hosts:       hosts,
			})
		}

		ctx := context.Background()
		dockerCli, err := dockerClient.NewClientWithOpts(dockerClient.FromEnv)
		if err != nil {
			return components, metricsEnabled, err
		}
		containers, err := dockerCli.ContainerList(ctx, types.ContainerListOptions{})
		if err != nil {
			return components, metricsEnabled, err
		}

		for _, container := range containers {
			containerName := container.Names[0]
			if !strings.Contains(containerName, "memphis") {
				continue
			}
			containerName = strings.TrimPrefix(containerName, "/")
			if container.State != "running" {
				comp := defaultSystemComp(containerName, false)
				allComponents = append(allComponents, comp)
				continue
			}
			containerStats, err := dockerCli.ContainerStats(ctx, container.ID, false)
			if err != nil {
				return components, metricsEnabled, err
			}
			defer containerStats.Body.Close()

			body, err := ioutil.ReadAll(containerStats.Body) // TODO replace ioutil
			if err != nil {
				return components, metricsEnabled, err
			}
			var dockerStats types.Stats
			err = json.Unmarshal(body, &dockerStats)
			if err != nil {
				return components, metricsEnabled, err
			}
			cpuLimit := float64(runtime.GOMAXPROCS(0))
			cpuPercentage := math.Ceil((float64(dockerStats.CPUStats.CPUUsage.TotalUsage) / float64(dockerStats.CPUStats.SystemUsage)) * 100)
			totalCpuUsage := (cpuPercentage / 100) * cpuLimit
			totalMemoryUsage := float64(dockerStats.MemoryStats.Usage)
			memoryLimit := float64(dockerStats.MemoryStats.Limit)
			memoryPercentage := math.Ceil((float64(totalMemoryUsage) / float64(memoryLimit)) * 100)
			storage_size, err := getUnixStorageSize()
			if err != nil {
				return components, metricsEnabled, err
			}
			cpuStat := models.CompStats{
				Total:      shortenFloat(cpuLimit),
				Current:    shortenFloat(totalCpuUsage),
				Percentage: int(cpuPercentage),
			}
			memoryStat := models.CompStats{
				Total:      shortenFloat(memoryLimit),
				Current:    shortenFloat(totalMemoryUsage),
				Percentage: int(memoryPercentage),
			}
			storageStat := defaultStat
			dockerPorts := []int{}
			if strings.Contains(containerName, "mongo") {
				dbStorageSize, totalSize, err := getDbStorageSize()
				if err != nil {
					return components, metricsEnabled, err
				}
				storageStat = models.CompStats{
					Total:      shortenFloat(totalSize),
					Current:    shortenFloat(dbStorageSize),
					Percentage: int(math.Ceil(float64(dbStorageSize) / float64(totalSize))),
				}

			} else if strings.Contains(containerName, "cluster") {
				v, err := serv.Varz(nil)
				if err != nil {
					return components, metricsEnabled, err
				}
				storageStat = models.CompStats{
					Total:      shortenFloat(storage_size),
					Current:    shortenFloat(float64(v.JetStream.Stats.Store)),
					Percentage: int(math.Ceil(float64(v.JetStream.Stats.Store) / storage_size)),
				}
			}
			for _, port := range container.Ports {
				if int(port.PublicPort) != 0 {
					dockerPorts = append(dockerPorts, int(port.PublicPort))
				}
			}
			comps := []models.SysComponent{{
				Name:    containerName,
				CPU:     cpuStat,
				Memory:  memoryStat,
				Storage: storageStat,
				Healthy: true,
			}}
			components = append(components, models.SystemComponents{
				Name:        containerName,
				Components:  comps,
				Status:      checkCompStatus(comps),
				Ports:       dockerPorts,
				DesiredPods: 1,
				ActualPods:  1,
				Hosts:       hosts,
			})
		}
	} else if configuration.LOCAL_CLUSTER_ENV { // TODO not fully supported - currently shows the current broker stats only
		metricsEnabled = true
		hosts = []string{"localhost"}
		maxCpu := float64(runtime.GOMAXPROCS(0))
		v, err := serv.Varz(nil)
		if err != nil {
			return components, metricsEnabled, err
		}
		var storageComp models.CompStats
		memUsage := float64(0)
		os := runtime.GOOS
		storage_size := float64(0)
		isWindows := false
		switch os {
		case "windows":
			isWindows = true
			storageComp = defaultStat // TODO: add support for windows
		default:
			storage_size, err = getUnixStorageSize()
			if err != nil {
				return components, metricsEnabled, err
			}
			storageComp = models.CompStats{
				Total:      shortenFloat(storage_size),
				Current:    shortenFloat(float64(v.JetStream.Stats.Store)),
				Percentage: int(math.Ceil((float64(v.JetStream.Stats.Store) / storage_size) * 100)),
			}
			memUsage, err = getUnixMemoryUsage()
			if err != nil {
				return components, metricsEnabled, err
			}
		}
		memPerc := (memUsage / float64(v.JetStream.Config.MaxMemory)) * 100
		cpuComps := []models.SysComponent{{
			Name: "memphis-0",
			CPU: models.CompStats{
				Total:      shortenFloat(maxCpu),
				Current:    shortenFloat((v.CPU / 100) * maxCpu),
				Percentage: int(math.Ceil(v.CPU)),
			},
			Memory: models.CompStats{
				Total:      shortenFloat(float64(v.JetStream.Config.MaxMemory)),
				Current:    shortenFloat(memUsage),
				Percentage: int(math.Ceil(memPerc)),
			},
			Storage: storageComp,
			Healthy: true,
		}}
		components = append(components, models.SystemComponents{
			Name:        "memphis",
			Components:  cpuComps,
			Status:      checkCompStatus(cpuComps),
			Ports:       []int{9000, 6666, 7770, 8222},
			DesiredPods: 1,
			ActualPods:  1,
			Hosts:       hosts,
		})
		resp, err := http.Get("http://localhost:4444/monitoring/getResourcesUtilization")
		healthy := false
		restGwComps := []models.SysComponent{defaultSystemComp("memphis-rest-gateway", healthy)}
		if err == nil {
			healthy = true
			var restGwMonitorInfo models.RestGwMonitoringResponse
			defer resp.Body.Close()
			err = json.NewDecoder(resp.Body).Decode(&restGwMonitorInfo)
			if err != nil {
				return components, metricsEnabled, err
			}
			if !isWindows {
				storageComp = models.CompStats{
					Total:      shortenFloat(storage_size),
					Current:    shortenFloat((restGwMonitorInfo.Storage / 100) * storage_size),
					Percentage: int(math.Ceil(float64(restGwMonitorInfo.Storage))),
				}
			}
			restGwComps = []models.SysComponent{{
				Name: "memphis-rest-gateway",
				CPU: models.CompStats{
					Total:      shortenFloat(maxCpu),
					Current:    shortenFloat((restGwMonitorInfo.CPU / 100) * maxCpu),
					Percentage: int(math.Ceil(restGwMonitorInfo.CPU)),
				},
				Memory: models.CompStats{
					Total:      shortenFloat(float64(v.JetStream.Config.MaxMemory)),
					Current:    shortenFloat((restGwMonitorInfo.Memory / 100) * float64(v.JetStream.Config.MaxMemory)),
					Percentage: int(math.Ceil(float64(restGwMonitorInfo.Memory))),
				},
				Storage: storageComp,
				Healthy: healthy,
			}}
		}
		actualRestGw := 1
		if !healthy {
			actualRestGw = 0
		}
		components = append(components, models.SystemComponents{
			Name:        "memphis-rest-gateway",
			Components:  restGwComps,
			Status:      checkCompStatus(restGwComps),
			Ports:       []int{4444},
			DesiredPods: 1,
			ActualPods:  actualRestGw,
			Hosts:       hosts,
		})

		ctx := context.Background()
		dockerCli, err := dockerClient.NewClientWithOpts(dockerClient.FromEnv)
		if err != nil {
			return components, metricsEnabled, err
		}
		containers, err := dockerCli.ContainerList(ctx, types.ContainerListOptions{})
		if err != nil {
			return components, metricsEnabled, err
		}

		for _, container := range containers {
			containerName := container.Names[0]
			if !strings.Contains(containerName, "memphis") {
				continue
			}
			containerName = strings.TrimPrefix(containerName, "/")
			if container.State != "running" {
				comp := defaultSystemComp(containerName, false)
				allComponents = append(allComponents, comp)
				continue
			}
			containerStats, err := dockerCli.ContainerStats(ctx, container.ID, false)
			if err != nil {
				return components, metricsEnabled, err
			}
			defer containerStats.Body.Close()

			body, err := ioutil.ReadAll(containerStats.Body) // TODO replace ioutil
			if err != nil {
				return components, metricsEnabled, err
			}
			var dockerStats types.Stats
			err = json.Unmarshal(body, &dockerStats)
			if err != nil {
				return components, metricsEnabled, err
			}
			cpuLimit := float64(runtime.GOMAXPROCS(0))
			cpuPercentage := math.Ceil((float64(dockerStats.CPUStats.CPUUsage.TotalUsage) / float64(dockerStats.CPUStats.SystemUsage)) * 100)
			totalCpuUsage := (cpuPercentage / 100) * cpuLimit
			totalMemoryUsage := float64(dockerStats.MemoryStats.Usage)
			memoryLimit := float64(dockerStats.MemoryStats.Limit)
			memoryPercentage := math.Ceil((float64(totalMemoryUsage) / float64(memoryLimit)) * 100)
			storage_size, err := getUnixStorageSize()
			if err != nil {
				return components, metricsEnabled, err
			}
			cpuStat := models.CompStats{
				Total:      shortenFloat(cpuLimit),
				Current:    shortenFloat(totalCpuUsage),
				Percentage: int(cpuPercentage),
			}
			memoryStat := models.CompStats{
				Total:      shortenFloat(memoryLimit),
				Current:    shortenFloat(totalMemoryUsage),
				Percentage: int(memoryPercentage),
			}
			storageStat := defaultStat
			dockerPorts := []int{}
			if strings.Contains(containerName, "mongo") {
				dbStorageSize, totalSize, err := getDbStorageSize()
				if err != nil {
					return components, metricsEnabled, err
				}
				storageStat = models.CompStats{
					Total:      shortenFloat(totalSize),
					Current:    shortenFloat(dbStorageSize),
					Percentage: int(math.Ceil(float64(dbStorageSize) / float64(totalSize))),
				}

			} else if strings.Contains(containerName, "cluster") {
				v, err := serv.Varz(nil)
				if err != nil {
					return components, metricsEnabled, err
				}
				storageStat = models.CompStats{
					Total:      shortenFloat(storage_size),
					Current:    shortenFloat(float64(v.JetStream.Stats.Store)),
					Percentage: int(math.Ceil(float64(v.JetStream.Stats.Store) / storage_size)),
				}
			}
			for _, port := range container.Ports {
				if int(port.PublicPort) != 0 {
					dockerPorts = append(dockerPorts, int(port.PublicPort))
				}
			}
			comps := []models.SysComponent{{
				Name:    containerName,
				CPU:     cpuStat,
				Memory:  memoryStat,
				Storage: storageStat,
				Healthy: true,
			}}
			components = append(components, models.SystemComponents{
				Name:        containerName,
				Components:  comps,
				Status:      checkCompStatus(comps),
				Ports:       dockerPorts,
				DesiredPods: 1,
				ActualPods:  1,
				Hosts:       hosts,
			})
		}
	} else { // k8s env
		if clientset == nil {
			err := clientSetClusterConfig()
			if err != nil {
				return components, metricsEnabled, err
			}
		}
		deploymentsClient := clientset.AppsV1().Deployments(configuration.K8S_NAMESPACE)
		deploymentsList, err := deploymentsClient.List(context.TODO(), metav1.ListOptions{})
		if err != nil {
			return components, metricsEnabled, err
		}

		pods, err := clientset.CoreV1().Pods(configuration.K8S_NAMESPACE).List(context.TODO(), metav1.ListOptions{})
		if err != nil {
			return components, metricsEnabled, err
		}
		minikubeCheck := false
		isMinikube := false
		for _, pod := range pods.Items {
			if pod.Status.Phase != v1.PodRunning {
				allComponents = append(allComponents, defaultSystemComp(pod.Name, false))
				continue
			}
			var ports []int
			podMetrics, err := metricsclientset.MetricsV1beta1().PodMetricses(configuration.K8S_NAMESPACE).Get(context.TODO(), pod.Name, metav1.GetOptions{})
			if err != nil {
				if strings.Contains(err.Error(), "could not find the requested resource") {
					metricsEnabled = false
					allComponents = append(allComponents, defaultSystemComp(pod.Name, true))
					if !noMetricsInstalledLog {
						serv.Warnf("GetSystemComponents: k8s metrics not installed: " + err.Error())
						noMetricsInstalledLog = true
					}
					continue
				} else if strings.Contains(err.Error(), "is forbidden") {
					metricsEnabled = false
					allComponents = append(allComponents, defaultSystemComp(pod.Name, true))
					if !noMetricsPermissionLog {
						serv.Warnf("GetSystemComponents: No permissions for k8s metrics: " + err.Error())
						noMetricsPermissionLog = true
					}
					continue
				}
				return components, metricsEnabled, err
			}
			node, err := clientset.CoreV1().Nodes().Get(context.TODO(), pod.Spec.NodeName, metav1.GetOptions{})
			if err != nil {
				return components, metricsEnabled, err
			}
			if !minikubeCheck {
				isMinikube = checkIsMinikube(node.Labels)
			}
			pvcClient := clientset.CoreV1().PersistentVolumeClaims(configuration.K8S_NAMESPACE)
			pvcList, err := pvcClient.List(context.TODO(), metav1.ListOptions{})
			if err != nil {
				return components, metricsEnabled, err
			}
			cpuLimit := pod.Spec.Containers[0].Resources.Limits.Cpu().AsApproximateFloat64()
			if cpuLimit == float64(0) {
				cpuLimit = node.Status.Capacity.Cpu().AsApproximateFloat64()
			}
			memLimit := pod.Spec.Containers[0].Resources.Limits.Memory().AsApproximateFloat64()
			if memLimit == float64(0) {
				memLimit = node.Status.Capacity.Memory().AsApproximateFloat64()
			}
			storageLimit := float64(0)
			if len(pvcList.Items) == 1 {
				size := pvcList.Items[0].Status.Capacity[v1.ResourceStorage]
				floatSize := size.AsApproximateFloat64()
				if floatSize != float64(0) {
					storageLimit = floatSize
				}
			} else {
				for _, pvc := range pvcList.Items {
					if strings.Contains(pvc.Name, pod.Name) {
						size := pvc.Status.Capacity[v1.ResourceStorage]
						floatSize := size.AsApproximateFloat64()
						if floatSize != float64(0) {
							storageLimit = floatSize
						}
						break
					}
				}
			}
			mountpath := ""
			containerForExec := ""
			for _, container := range pod.Spec.Containers {
				for _, port := range container.Ports {
					if int(port.ContainerPort) != 0 {
						ports = append(ports, int(port.ContainerPort))
					}
				}
				brokerMatch, err := regexp.MatchString(`^memphis-\d*[0-9]\d*$`, container.Name)
				if err != nil {
					return components, metricsEnabled, err
				}
				if brokerMatch || strings.Contains(container.Name, "memphis-rest-gateway") || strings.Contains(container.Name, "mongo") {
					for _, mount := range pod.Spec.Containers[0].VolumeMounts {
						if strings.Contains(mount.Name, "memphis") {
							mountpath = mount.MountPath
							break
						}
					}
					containerForExec = container.Name
				}
			}

			cpuUsage := float64(0)
			memUsage := float64(0)
			for _, container := range podMetrics.Containers {
				cpuUsage += container.Usage.Cpu().AsApproximateFloat64()
				memUsage += container.Usage.Memory().AsApproximateFloat64()
			}

			storageUsage := float64(0)
			if isMinikube {
				if strings.Contains(strings.ToLower(pod.Name), "mongo") {
					storageUsage, _, err = getDbStorageSize()
					if err != nil {
						return components, metricsEnabled, err
					}
				} else if strings.Contains(strings.ToLower(pod.Name), "cluster") {
					v, err := serv.Varz(nil)
					if err != nil {
						return components, metricsEnabled, err
					}
					storageUsage = shortenFloat(float64(v.JetStream.Stats.Store))
				}
			} else if containerForExec != "" && mountpath != "" {
				storageUsage, err = getContainerStorageUsage(config, mountpath, containerForExec, pod.Name)
				if err != nil {
					return components, metricsEnabled, err
				}
			}
			storagePercentage := 0
			if storageUsage > float64(0) && storageLimit > float64(0) {
				storagePercentage = int(math.Ceil((storageUsage / storageLimit) * 100))
			}

			comp := models.SysComponent{
				Name: pod.Name,
				CPU: models.CompStats{
					Total:      shortenFloat(cpuLimit),
					Current:    shortenFloat(cpuUsage),
					Percentage: int(math.Ceil((float64(cpuUsage) / float64(cpuLimit)) * 100)),
				},
				Memory: models.CompStats{
					Total:      shortenFloat(memLimit),
					Current:    shortenFloat(memUsage),
					Percentage: int(math.Ceil((float64(memUsage) / float64(memLimit)) * 100)),
				},
				Storage: models.CompStats{
					Total:      shortenFloat(storageLimit),
					Current:    shortenFloat(storageUsage),
					Percentage: storagePercentage,
				},
				Healthy: true,
			}
			allComponents = append(allComponents, comp)
			portsMap[pod.Name] = ports
		}

		for _, d := range deploymentsList.Items {
			desired := int(*d.Spec.Replicas)
			actual := int(d.Status.ReadyReplicas)
			relevantComponents := getRelevantComponents(d.Name, allComponents, desired)
			var relevantPorts []int
			var status string
			if metricsEnabled {
				relevantPorts = getRelevantPorts(d.Name, portsMap)
				status = checkCompStatus(relevantComponents)
			} else {
				for _, container := range d.Spec.Template.Spec.Containers {
					for _, port := range container.Ports {
						if int(port.ContainerPort) != 0 {
							relevantPorts = append(relevantPorts, int(port.ContainerPort))
						}
					}
				}
				if desired == actual {
					status = "healthy"
				} else {
					status = "unhealthy"
				}
			}
			brokerMatch, err := regexp.MatchString(`^memphis-\d*[0-9]\d*$`, d.Name)
			if err != nil {
				return components, metricsEnabled, err
			}
			if brokerMatch {
				if BROKER_HOST == "" {
					hosts = []string{}
				} else {
					hosts = []string{BROKER_HOST}
				}
				if UI_HOST != "" {
					hosts = append(hosts, UI_HOST)
				}
			} else if strings.Contains(d.Name, "memphis-rest-gateway") {
				if REST_GW_HOST != "" {
					hosts = []string{REST_GW_HOST}
				}
			} else if strings.Contains(d.Name, "mongo") {
				hosts = []string{}
			}
			components = append(components, models.SystemComponents{
				Name:        d.Name,
				Components:  relevantComponents,
				Status:      status,
				Ports:       relevantPorts,
				DesiredPods: desired,
				ActualPods:  actual,
				Hosts:       hosts,
			})
		}

		statefulsetsClient := clientset.AppsV1().StatefulSets(configuration.K8S_NAMESPACE)
		statefulsetsList, err := statefulsetsClient.List(context.TODO(), metav1.ListOptions{})
		if err != nil {
			return components, metricsEnabled, err
		}
		for _, s := range statefulsetsList.Items {
			desired := int(*s.Spec.Replicas)
			actual := int(s.Status.ReadyReplicas)
			relevantComponents := getRelevantComponents(s.Name, allComponents, desired)
			var relevantPorts []int
			var status string
			if metricsEnabled {
				relevantPorts = getRelevantPorts(s.Name, portsMap)
				status = checkCompStatus(relevantComponents)
			} else {
				for _, container := range s.Spec.Template.Spec.Containers {
					for _, port := range container.Ports {
						if int(port.ContainerPort) != 0 {
							relevantPorts = append(relevantPorts, int(port.ContainerPort))
						}
					}
				}
				if desired == actual {
					status = "healthy"
				} else {
					status = "unhealthy"
				}
			}
			brokerMatch, err := regexp.MatchString(`^memphis-\d*[0-9]\d*$`, s.Name)
			if err != nil {
				return components, metricsEnabled, err
			}
			if brokerMatch {
				if BROKER_HOST == "" {
					hosts = []string{}
				} else {
					hosts = []string{BROKER_HOST}
				}
				if UI_HOST != "" {
					hosts = append(hosts, UI_HOST)
				}
			} else if strings.Contains(s.Name, "memphis-rest-gateway") {
				if REST_GW_HOST != "" {
					hosts = []string{REST_GW_HOST}
				}
			} else if strings.Contains(s.Name, "mongo") {
				hosts = []string{}
			}
			components = append(components, models.SystemComponents{
				Name:        s.Name,
				Components:  relevantComponents,
				Status:      status,
				Ports:       relevantPorts,
				DesiredPods: desired,
				ActualPods:  actual,
				Hosts:       hosts,
			})
		}
	}
	return components, metricsEnabled, nil
}

func (mh MonitoringHandler) GetClusterInfo(c *gin.Context) {
	fileContent, err := ioutil.ReadFile("version.conf")
	if err != nil {
		serv.Errorf("GetClusterInfo: " + err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}
	c.IndentedJSON(200, gin.H{"version": string(fileContent)})
}

func (mh MonitoringHandler) GetBrokersThroughputs() ([]models.BrokerThroughputResponse, error) {
	uid := serv.memphis.nuid.Next()
	durableName := "$memphis_fetch_throughput_consumer_" + uid
	var msgs []StoredMsg
	var throughputs []models.BrokerThroughputResponse
	streamInfo, err := serv.memphisStreamInfo(throughputStreamNameV1)
	if err != nil {
		return throughputs, err
	}

	amount := streamInfo.State.Msgs
	startSeq := uint64(1)
	if streamInfo.State.FirstSeq > 0 {
		startSeq = streamInfo.State.FirstSeq
	}

	cc := ConsumerConfig{
		OptStartSeq:   startSeq,
		DeliverPolicy: DeliverByStartSequence,
		AckPolicy:     AckExplicit,
		Durable:       durableName,
	}

	err = serv.memphisAddConsumer(throughputStreamNameV1, &cc)
	if err != nil {
		return throughputs, err
	}

	responseChan := make(chan StoredMsg)
	subject := fmt.Sprintf(JSApiRequestNextT, throughputStreamNameV1, durableName)
	reply := durableName + "_reply"
	req := []byte(strconv.FormatUint(amount, 10))

	sub, err := serv.subscribeOnGlobalAcc(reply, reply+"_sid", func(_ *client, subject, reply string, msg []byte) {
		go func(respCh chan StoredMsg, subject, reply string, msg []byte) {
			// ack
			serv.sendInternalAccountMsg(serv.GlobalAccount(), reply, []byte(_EMPTY_))
			rawTs := tokenAt(reply, 8)
			seq, _, _ := ackReplyInfo(reply)

			intTs, err := strconv.Atoi(rawTs)
			if err != nil {
				serv.Errorf("GetBrokersThroughputs: " + err.Error())
			}

			respCh <- StoredMsg{
				Subject:  subject,
				Sequence: uint64(seq),
				Data:     msg,
				Time:     time.Unix(0, int64(intTs)),
			}
		}(responseChan, subject, reply, copyBytes(msg))
	})
	if err != nil {
		return throughputs, err
	}

	serv.sendInternalAccountMsgWithReply(serv.GlobalAccount(), subject, reply, nil, req, true)
	timeout := 300 * time.Millisecond
	timer := time.NewTimer(timeout)
	for i := uint64(0); i < amount; i++ {
		select {
		case <-timer.C:
			goto cleanup
		case msg := <-responseChan:
			msgs = append(msgs, msg)
		}
	}

cleanup:
	timer.Stop()
	serv.unsubscribeOnGlobalAcc(sub)
	err = serv.memphisRemoveConsumer(throughputStreamNameV1, durableName)
	if err != nil {
		return throughputs, err
	}

	sort.Slice(msgs, func(i, j int) bool { // old to new
		return msgs[i].Time.Before(msgs[j].Time)
	})

	m := make(map[string]models.BrokerThroughputResponse)
	for _, msg := range msgs {
		var brokerThroughput models.BrokerThroughput
		err = json.Unmarshal(msg.Data, &brokerThroughput)
		if err != nil {
			return throughputs, err
		}

		if _, ok := m[brokerThroughput.Name]; !ok {
			m[brokerThroughput.Name] = models.BrokerThroughputResponse{
				Name: brokerThroughput.Name,
			}
		}

		mapEntry := m[brokerThroughput.Name]
		mapEntry.Read = append(m[brokerThroughput.Name].Read, models.ThroughputReadResponse{
			Timestamp: msg.Time,
			Read:      brokerThroughput.Read,
		})
		mapEntry.Write = append(m[brokerThroughput.Name].Write, models.ThroughputWriteResponse{
			Timestamp: msg.Time,
			Write:     brokerThroughput.Write,
		})
		m[brokerThroughput.Name] = mapEntry
	}

	throughputs = make([]models.BrokerThroughputResponse, 0, len(m))
	totalRead := make([]models.ThroughputReadResponse, ws_updates_interval_sec)
	totalWrite := make([]models.ThroughputWriteResponse, ws_updates_interval_sec)
	for _, t := range m {
		throughputs = append(throughputs, t)
		for i, r := range t.Read {
			totalRead[i].Timestamp = r.Timestamp
			totalRead[i].Read += r.Read
		}
		for i, w := range t.Write {
			totalWrite[i].Timestamp = w.Timestamp
			totalWrite[i].Write += w.Write
		}
	}
	throughputs = append([]models.BrokerThroughputResponse{{
		Name:  "total",
		Read:  totalRead,
		Write: totalWrite,
	}}, throughputs...)

	return throughputs, nil
}

func (mh MonitoringHandler) GetMainOverviewData(c *gin.Context) {
	stationsHandler := StationsHandler{S: mh.S}
	stations, totalMessages, totalDlsMsgs, err := stationsHandler.GetAllStationsDetails()
	if err != nil {
		serv.Errorf("GetMainOverviewData: GetAllStationsDetails: " + err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}
	systemComponents, metricsEnabled, err := mh.GetSystemComponents()
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "cannot connect to the docker daemon") {
			serv.Warnf("GetMainOverviewData: GetSystemComponents: " + err.Error())
			c.AbortWithStatusJSON(configuration.SHOWABLE_ERROR_STATUS_CODE, gin.H{"message": "Failed getting system components data: " + err.Error()})
		} else {
			serv.Errorf("GetMainOverviewData: GetSystemComponents: " + err.Error())
			c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		}
		return
	}
	k8sEnv := true
	if configuration.DOCKER_ENV == "true" || configuration.LOCAL_CLUSTER_ENV {
		k8sEnv = false
	}
	brokersThroughputs, err := mh.GetBrokersThroughputs()
	if err != nil {
		serv.Errorf("GetMainOverviewData: GetBrokersThroughputs: " + err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}
	response := models.MainOverviewData{
		TotalStations:     len(stations),
		TotalMessages:     totalMessages,
		TotalDlsMessages:  totalDlsMsgs,
		SystemComponents:  systemComponents,
		Stations:          stations,
		K8sEnv:            k8sEnv,
		BrokersThroughput: brokersThroughputs,
		MetricsEnabled:    metricsEnabled,
	}

	shouldSendAnalytics, _ := shouldSendAnalytics()
	if shouldSendAnalytics {
		user, _ := getUserDetailsFromMiddleware(c)
		analytics.SendEvent(user.Username, "user-enter-main-overview")
	}

	c.IndentedJSON(200, response)
}

func getFakeProdsAndConsForPreview() ([]map[string]interface{}, []map[string]interface{}, []map[string]interface{}, []map[string]interface{}) {
	connectedProducers := make([]map[string]interface{}, 0)
	connectedProducers = append(connectedProducers, map[string]interface{}{
		"id":              "63b68df439e19dd69996f3d8",
		"name":            "prod.20",
		"type":            "application",
		"connection_id":   "f95f24fbcf959dfb941e6ff3",
		"created_by_user": "root",
		"creation_date":   "2023-01-05T08:44:36.999Z",
		"station_name":    "idanasulin6",
		"is_active":       true,
		"is_deleted":      false,
		"client_address":  "127.0.0.1:61430",
	})
	connectedProducers = append(connectedProducers, map[string]interface{}{
		"id":              "63b68df439e19dd69996f3d6",
		"name":            "prod.19",
		"type":            "application",
		"connection_id":   "f95f24fbcf959dfb941e6ff3",
		"created_by_user": "root",
		"creation_date":   "2023-01-05T08:44:36.99Z",
		"station_name":    "idanasulin6",
		"is_active":       true,
		"is_deleted":      false,
		"client_address":  "127.0.0.1:61430",
	})
	connectedProducers = append(connectedProducers, map[string]interface{}{
		"id":              "63b68df439e19dd69996f3d4",
		"name":            "prod.18",
		"type":            "application",
		"connection_id":   "f95f24fbcf959dfb941e6ff3",
		"created_by_user": "root",
		"creation_date":   "2023-01-05T08:44:36.982Z",
		"station_name":    "idanasulin6",
		"is_active":       true,
		"is_deleted":      false,
		"client_address":  "127.0.0.1:61430",
	})
	connectedProducers = append(connectedProducers, map[string]interface{}{
		"id":              "63b68df439e19dd69996f3d2",
		"name":            "prod.17",
		"type":            "application",
		"connection_id":   "f95f24fbcf959dfb941e6ff3",
		"created_by_user": "root",
		"creation_date":   "2023-01-05T08:44:36.969Z",
		"station_name":    "idanasulin6",
		"is_active":       true,
		"is_deleted":      false,
		"client_address":  "127.0.0.1:61430",
	})

	disconnectedProducers := make([]map[string]interface{}, 0)
	disconnectedProducers = append(connectedProducers, map[string]interface{}{
		"id":              "63b68df439e19dd69996f3d0",
		"name":            "prod.16",
		"type":            "application",
		"connection_id":   "f95f24fbcf959dfb941e6ff3",
		"created_by_user": "root",
		"creation_date":   "2023-01-05T08:44:36.959Z",
		"station_name":    "idanasulin6",
		"is_active":       false,
		"is_deleted":      false,
		"client_address":  "127.0.0.1:61430",
	})
	disconnectedProducers = append(connectedProducers, map[string]interface{}{
		"id":              "63b68df439e19dd69996f3ce",
		"name":            "prod.15",
		"type":            "application",
		"connection_id":   "f95f24fbcf959dfb941e6ff3",
		"created_by_user": "root",
		"creation_date":   "2023-01-05T08:44:36.951Z",
		"station_name":    "idanasulin6",
		"is_active":       false,
		"is_deleted":      false,
		"client_address":  "127.0.0.1:61430",
	})
	disconnectedProducers = append(connectedProducers, map[string]interface{}{
		"id":              "63b68df439e19dd69996f3cc",
		"name":            "prod.14",
		"type":            "application",
		"connection_id":   "f95f24fbcf959dfb941e6ff3",
		"created_by_user": "root",
		"creation_date":   "2023-01-05T08:44:36.941Z",
		"station_name":    "idanasulin6",
		"is_active":       false,
		"is_deleted":      false,
		"client_address":  "127.0.0.1:61430",
	})
	disconnectedProducers = append(connectedProducers, map[string]interface{}{
		"id":              "63b68df439e19dd69996f3ca",
		"name":            "prod.13",
		"type":            "application",
		"connection_id":   "f95f24fbcf959dfb941e6ff3",
		"created_by_user": "root",
		"creation_date":   "2023-01-05T08:44:36.93Z",
		"station_name":    "idanasulin6",
		"is_active":       false,
		"is_deleted":      false,
		"client_address":  "127.0.0.1:61430",
	})
	disconnectedProducers = append(connectedProducers, map[string]interface{}{
		"id":              "63b68df439e19dd69996f3c8",
		"name":            "prod.12",
		"type":            "application",
		"connection_id":   "f95f24fbcf959dfb941e6ff3",
		"created_by_user": "root",
		"creation_date":   "2023-01-05T08:44:36.92Z",
		"station_name":    "idanasulin6",
		"is_active":       false,
		"is_deleted":      false,
		"client_address":  "127.0.0.1:61430",
	})
	disconnectedProducers = append(connectedProducers, map[string]interface{}{
		"id":              "63b68df439e19dd69996f3c6",
		"name":            "prod.11",
		"type":            "application",
		"connection_id":   "f95f24fbcf959dfb941e6ff3",
		"created_by_user": "root",
		"creation_date":   "2023-01-05T08:44:36.911Z",
		"station_name":    "idanasulin6",
		"is_active":       false,
		"is_deleted":      false,
		"client_address":  "127.0.0.1:61430",
	})
	disconnectedProducers = append(connectedProducers, map[string]interface{}{
		"id":              "63b68df439e19dd69996f3c4",
		"name":            "prod.10",
		"type":            "application",
		"connection_id":   "f95f24fbcf959dfb941e6ff3",
		"created_by_user": "root",
		"creation_date":   "2023-01-05T08:44:36.902Z",
		"station_name":    "idanasulin6",
		"is_active":       false,
		"is_deleted":      false,
		"client_address":  "127.0.0.1:61430",
	})
	disconnectedProducers = append(connectedProducers, map[string]interface{}{
		"id":              "63b68df439e19dd69996f3c2",
		"name":            "prod.9",
		"type":            "application",
		"connection_id":   "f95f24fbcf959dfb941e6ff3",
		"created_by_user": "root",
		"creation_date":   "2023-01-05T08:44:36.892Z",
		"station_name":    "idanasulin6",
		"is_active":       false,
		"is_deleted":      false,
		"client_address":  "127.0.0.1:61430",
	})
	disconnectedProducers = append(connectedProducers, map[string]interface{}{
		"id":              "63b68df439e19dd69996f3c0",
		"name":            "prod.8",
		"type":            "application",
		"connection_id":   "f95f24fbcf959dfb941e6ff3",
		"created_by_user": "root",
		"creation_date":   "2023-01-05T08:44:36.882Z",
		"station_name":    "idanasulin6",
		"is_active":       false,
		"is_deleted":      false,
		"client_address":  "127.0.0.1:61430",
	})
	disconnectedProducers = append(connectedProducers, map[string]interface{}{
		"id":              "63b68df439e19dd69996f3be",
		"name":            "prod.7",
		"type":            "application",
		"connection_id":   "f95f24fbcf959dfb941e6ff3",
		"created_by_user": "root",
		"creation_date":   "2023-01-05T08:44:36.872Z",
		"station_name":    "idanasulin6",
		"is_active":       false,
		"is_deleted":      false,
		"client_address":  "127.0.0.1:61430",
	})
	disconnectedProducers = append(connectedProducers, map[string]interface{}{
		"id":              "63b68df439e19dd69996f3bc",
		"name":            "prod.6",
		"type":            "application",
		"connection_id":   "f95f24fbcf959dfb941e6ff3",
		"created_by_user": "root",
		"creation_date":   "2023-01-05T08:44:36.862Z",
		"station_name":    "idanasulin6",
		"is_active":       false,
		"is_deleted":      false,
		"client_address":  "127.0.0.1:61430",
	})

	connectedCgs := make([]map[string]interface{}, 0)
	connectedCgs = append(connectedCgs, map[string]interface{}{
		"name":                    "cg.20",
		"unprocessed_messages":    0,
		"poison_messages":         0,
		"is_active":               true,
		"is_deleted":              false,
		"in_process_messages":     0,
		"max_ack_time_ms":         30000,
		"max_msg_deliveries":      10,
		"connected_consumers":     []string{},
		"disconnected_consumers":  []string{},
		"deleted_consumers":       []string{},
		"last_status_change_date": "2023-01-05T08:44:37.165Z",
	})
	connectedCgs = append(connectedCgs, map[string]interface{}{
		"name":                    "cg.19",
		"unprocessed_messages":    0,
		"poison_messages":         0,
		"is_active":               true,
		"is_deleted":              false,
		"in_process_messages":     0,
		"max_ack_time_ms":         30000,
		"max_msg_deliveries":      10,
		"connected_consumers":     []string{},
		"disconnected_consumers":  []string{},
		"deleted_consumers":       []string{},
		"last_status_change_date": "2023-01-05T08:44:37.165Z",
	})
	connectedCgs = append(connectedCgs, map[string]interface{}{
		"name":                    "cg.18",
		"unprocessed_messages":    0,
		"poison_messages":         0,
		"is_active":               true,
		"is_deleted":              false,
		"in_process_messages":     0,
		"max_ack_time_ms":         30000,
		"max_msg_deliveries":      10,
		"connected_consumers":     []string{},
		"disconnected_consumers":  []string{},
		"deleted_consumers":       []string{},
		"last_status_change_date": "2023-01-05T08:44:37.165Z",
	})
	connectedCgs = append(connectedCgs, map[string]interface{}{
		"name":                    "cg.17",
		"unprocessed_messages":    0,
		"poison_messages":         0,
		"is_active":               true,
		"is_deleted":              false,
		"in_process_messages":     0,
		"max_ack_time_ms":         30000,
		"max_msg_deliveries":      10,
		"connected_consumers":     []string{},
		"disconnected_consumers":  []string{},
		"deleted_consumers":       []string{},
		"last_status_change_date": "2023-01-05T08:44:37.165Z",
	})
	connectedCgs = append(connectedCgs, map[string]interface{}{
		"name":                    "cg.16",
		"unprocessed_messages":    0,
		"poison_messages":         0,
		"is_active":               true,
		"is_deleted":              false,
		"in_process_messages":     0,
		"max_ack_time_ms":         30000,
		"max_msg_deliveries":      10,
		"connected_consumers":     []string{},
		"disconnected_consumers":  []string{},
		"deleted_consumers":       []string{},
		"last_status_change_date": "2023-01-05T08:44:37.165Z",
	})

	disconnectedCgs := make([]map[string]interface{}, 0)
	disconnectedCgs = append(disconnectedCgs, map[string]interface{}{
		"name":                    "cg.15",
		"unprocessed_messages":    0,
		"poison_messages":         0,
		"is_active":               false,
		"is_deleted":              false,
		"in_process_messages":     0,
		"max_ack_time_ms":         30000,
		"max_msg_deliveries":      10,
		"connected_consumers":     []string{},
		"disconnected_consumers":  []string{},
		"deleted_consumers":       []string{},
		"last_status_change_date": "2023-01-05T08:44:37.165Z",
	})
	disconnectedCgs = append(disconnectedCgs, map[string]interface{}{
		"name":                    "cg.14",
		"unprocessed_messages":    0,
		"poison_messages":         0,
		"is_active":               false,
		"is_deleted":              false,
		"in_process_messages":     0,
		"max_ack_time_ms":         30000,
		"max_msg_deliveries":      10,
		"connected_consumers":     []string{},
		"disconnected_consumers":  []string{},
		"deleted_consumers":       []string{},
		"last_status_change_date": "2023-01-05T08:44:37.165Z",
	})
	disconnectedCgs = append(disconnectedCgs, map[string]interface{}{
		"name":                    "cg.13",
		"unprocessed_messages":    0,
		"poison_messages":         0,
		"is_active":               false,
		"is_deleted":              false,
		"in_process_messages":     0,
		"max_ack_time_ms":         30000,
		"max_msg_deliveries":      10,
		"connected_consumers":     []string{},
		"disconnected_consumers":  []string{},
		"deleted_consumers":       []string{},
		"last_status_change_date": "2023-01-05T08:44:37.165Z",
	})
	disconnectedCgs = append(disconnectedCgs, map[string]interface{}{
		"name":                    "cg.12",
		"unprocessed_messages":    0,
		"poison_messages":         0,
		"is_active":               false,
		"is_deleted":              false,
		"in_process_messages":     0,
		"max_ack_time_ms":         30000,
		"max_msg_deliveries":      10,
		"connected_consumers":     []string{},
		"disconnected_consumers":  []string{},
		"deleted_consumers":       []string{},
		"last_status_change_date": "2023-01-05T08:44:37.165Z",
	})
	disconnectedCgs = append(disconnectedCgs, map[string]interface{}{
		"name":                    "cg.11",
		"unprocessed_messages":    0,
		"poison_messages":         0,
		"is_active":               false,
		"is_deleted":              false,
		"in_process_messages":     0,
		"max_ack_time_ms":         30000,
		"max_msg_deliveries":      10,
		"connected_consumers":     []string{},
		"disconnected_consumers":  []string{},
		"deleted_consumers":       []string{},
		"last_status_change_date": "2023-01-05T08:44:37.165Z",
	})
	disconnectedCgs = append(disconnectedCgs, map[string]interface{}{
		"name":                    "cg.10",
		"unprocessed_messages":    0,
		"poison_messages":         0,
		"is_active":               false,
		"is_deleted":              false,
		"in_process_messages":     0,
		"max_ack_time_ms":         30000,
		"max_msg_deliveries":      10,
		"connected_consumers":     []string{},
		"disconnected_consumers":  []string{},
		"deleted_consumers":       []string{},
		"last_status_change_date": "2023-01-05T08:44:37.165Z",
	})
	disconnectedCgs = append(disconnectedCgs, map[string]interface{}{
		"name":                    "cg.9",
		"unprocessed_messages":    0,
		"poison_messages":         0,
		"is_active":               false,
		"is_deleted":              false,
		"in_process_messages":     0,
		"max_ack_time_ms":         30000,
		"max_msg_deliveries":      10,
		"connected_consumers":     []string{},
		"disconnected_consumers":  []string{},
		"deleted_consumers":       []string{},
		"last_status_change_date": "2023-01-05T08:44:37.165Z",
	})

	return connectedProducers, disconnectedProducers, connectedCgs, disconnectedCgs
}

func (mh MonitoringHandler) GetStationOverviewData(c *gin.Context) {
	stationsHandler := StationsHandler{S: mh.S}
	producersHandler := ProducersHandler{S: mh.S}
	consumersHandler := ConsumersHandler{S: mh.S}
	auditLogsHandler := AuditLogsHandler{}
	poisonMsgsHandler := PoisonMessagesHandler{S: mh.S}
	tagsHandler := TagsHandler{S: mh.S}
	var body models.GetStationOverviewDataSchema
	ok := utils.Validate(c, &body, false, nil)
	if !ok {
		return
	}

	stationName, err := StationNameFromStr(body.StationName)
	if err != nil {
		serv.Errorf("GetStationOverviewData: At station " + body.StationName + ": " + err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}
	exist, station, err := db.GetStationByName(stationName.Ext())
	if err != nil {
		serv.Errorf("GetStationOverviewData: At station " + body.StationName + ": " + err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}
	if !exist {
		errMsg := "Station " + body.StationName + " does not exist"
		serv.Warnf("GetStationOverviewData: " + errMsg)
		c.AbortWithStatusJSON(404, gin.H{"message": errMsg})
		return
	}

	connectedProducers, disconnectedProducers, deletedProducers := make([]models.Producer, 0), make([]models.Producer, 0), make([]models.Producer, 0)
	if station.IsNative {
		connectedProducers, disconnectedProducers, deletedProducers, err = producersHandler.GetProducersByStation(station)
		if err != nil {
			serv.Errorf("GetStationOverviewData: At station " + body.StationName + ": " + err.Error())
			c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
			return
		}
	}

	auditLogs, err := auditLogsHandler.GetAuditLogsByStation(station)
	if err != nil {
		serv.Errorf("GetStationOverviewData: At station " + body.StationName + ": " + err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}
	totalMessages, err := stationsHandler.GetTotalMessages(station.Name)
	if err != nil {
		if IsNatsErr(err, JSStreamNotFoundErr) {
			serv.Warnf("GetStationOverviewData: Station " + body.StationName + " does not exist")
			c.AbortWithStatusJSON(404, gin.H{"message": "Station " + body.StationName + " does not exist"})
		} else {
			serv.Errorf("GetStationOverviewData: At station " + body.StationName + ": " + err.Error())
			c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		}
		return
	}
	avgMsgSize, err := stationsHandler.GetAvgMsgSize(station)
	if err != nil {
		if IsNatsErr(err, JSStreamNotFoundErr) {
			serv.Warnf("GetStationOverviewData: Station " + body.StationName + " does not exist")
			c.AbortWithStatusJSON(404, gin.H{"message": "Station " + body.StationName + " does not exist"})
		} else {
			serv.Errorf("GetStationOverviewData: At station " + body.StationName + ": " + err.Error())
			c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		}
		return
	}

	messagesToFetch := 1000
	messages, err := stationsHandler.GetMessages(station, messagesToFetch)
	if err != nil {
		if IsNatsErr(err, JSStreamNotFoundErr) {
			serv.Warnf("GetStationOverviewData: Station " + body.StationName + " does not exist")
			c.AbortWithStatusJSON(404, gin.H{"message": "Station " + body.StationName + " does not exist"})
		} else {
			serv.Errorf("GetStationOverviewData: At station " + body.StationName + ": " + err.Error())
			c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		}
		return
	}

	poisonMessages, schemaFailedMessages, totalDlsAmount, poisonCgMap, err := poisonMsgsHandler.GetDlsMsgsByStationLight(station)
	if err != nil {
		if IsNatsErr(err, JSStreamNotFoundErr) {
			serv.Warnf("GetStationOverviewData: Station " + body.StationName + " does not exist")
			c.AbortWithStatusJSON(404, gin.H{"message": "Station " + body.StationName + " does not exist"})
		} else {
			serv.Errorf("GetStationOverviewData: At station " + body.StationName + ": " + err.Error())
			c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		}
		return
	}

	connectedCgs, disconnectedCgs, deletedCgs := make([]models.Cg, 0), make([]models.Cg, 0), make([]models.Cg, 0)

	// Only native stations have CGs
	if station.IsNative {
		connectedCgs, disconnectedCgs, deletedCgs, err = consumersHandler.GetCgsByStation(stationName, station, poisonCgMap)
		if err != nil {
			serv.Errorf("GetStationOverviewData: At station " + body.StationName + ": " + err.Error())
			c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
			return
		}
	}

	tags, err := tagsHandler.GetTagsByEntityWithID("station", station.ID)
	if err != nil {
		serv.Errorf("GetStationOverviewData: At station " + body.StationName + ": " + err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}
	leader, followers, err := stationsHandler.GetLeaderAndFollowers(station)
	if err != nil {
		if IsNatsErr(err, JSStreamNotFoundErr) {
			serv.Warnf("GetStationOverviewData: Station " + body.StationName + " does not exist")
			c.AbortWithStatusJSON(404, gin.H{"message": "Station " + body.StationName + " does not exist"})
		} else {
			serv.Errorf("GetStationOverviewData: At station " + body.StationName + ": " + err.Error())
			c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		}
		return
	}

	_, ok = IntegrationsCache["s3"].(models.Integration)
	if !ok {
		station.TieredStorageEnabled = false
	}
	var response gin.H

	// Check when the schema object in station is not empty, not optional for non native stations
	if station.SchemaName != "" && station.SchemaVersionNumber != 0 {

		var schemaDetails models.StationOverviewSchemaDetails
		exist, schema, err := db.GetSchemaByName(station.SchemaName)
		if !exist {
			schemaDetails = models.StationOverviewSchemaDetails{}
		} else {
			if err != nil {
				serv.Errorf("GetStationOverviewData: At station " + body.StationName + ": " + err.Error())
				c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
				return
			}

			_, schemaVersion, err := db.GetSchemaVersionByNumberAndID(station.SchemaVersionNumber, schema.ID)
			if err != nil {
				serv.Errorf("GetStationOverviewData: At station " + body.StationName + ": " + err.Error())
				c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
				return
			}
			updatesAvailable := !schemaVersion.Active
			schemaDetails = models.StationOverviewSchemaDetails{
				SchemaName:       schema.Name,
				VersionNumber:    station.SchemaVersionNumber,
				UpdatesAvailable: updatesAvailable,
				SchemaType:       schema.Type,
			}
		}
		response = gin.H{
			"connected_producers":           connectedProducers,
			"disconnected_producers":        disconnectedProducers,
			"deleted_producers":             deletedProducers,
			"connected_cgs":                 connectedCgs,
			"disconnected_cgs":              disconnectedCgs,
			"deleted_cgs":                   deletedCgs,
			"total_messages":                totalMessages,
			"average_message_size":          avgMsgSize,
			"audit_logs":                    auditLogs,
			"messages":                      messages,
			"poison_messages":               poisonMessages,
			"schema_failed_messages":        schemaFailedMessages,
			"tags":                          tags,
			"leader":                        leader,
			"followers":                     followers,
			"schema":                        schemaDetails,
			"idempotency_window_in_ms":      station.IdempotencyWindow,
			"dls_configuration_poison":      station.DlsConfigurationPoison,
			"dls_configuration_schemaverse": station.DlsConfigurationSchemaverse,
			"total_dls_messages":            totalDlsAmount,
			"tiered_storage_enabled":        station.TieredStorageEnabled,
		}
	} else {
		var emptyResponse struct{}
		if !station.IsNative {
			cp, dp, cc, dc := getFakeProdsAndConsForPreview()
			response = gin.H{
				"connected_producers":           cp,
				"disconnected_producers":        dp,
				"deleted_producers":             deletedProducers,
				"connected_cgs":                 cc,
				"disconnected_cgs":              dc,
				"deleted_cgs":                   deletedCgs,
				"total_messages":                totalMessages,
				"average_message_size":          avgMsgSize,
				"audit_logs":                    auditLogs,
				"messages":                      messages,
				"poison_messages":               poisonMessages,
				"schema_failed_messages":        schemaFailedMessages,
				"tags":                          tags,
				"leader":                        leader,
				"followers":                     followers,
				"schema":                        emptyResponse,
				"idempotency_window_in_ms":      station.IdempotencyWindow,
				"dls_configuration_poison":      station.DlsConfigurationPoison,
				"dls_configuration_schemaverse": station.DlsConfigurationSchemaverse,
				"total_dls_messages":            totalDlsAmount,
				"tiered_storage_enabled":        station.TieredStorageEnabled,
			}
		} else {
			response = gin.H{
				"connected_producers":           connectedProducers,
				"disconnected_producers":        disconnectedProducers,
				"deleted_producers":             deletedProducers,
				"connected_cgs":                 connectedCgs,
				"disconnected_cgs":              disconnectedCgs,
				"deleted_cgs":                   deletedCgs,
				"total_messages":                totalMessages,
				"average_message_size":          avgMsgSize,
				"audit_logs":                    auditLogs,
				"messages":                      messages,
				"poison_messages":               poisonMessages,
				"schema_failed_messages":        schemaFailedMessages,
				"tags":                          tags,
				"leader":                        leader,
				"followers":                     followers,
				"schema":                        emptyResponse,
				"idempotency_window_in_ms":      station.IdempotencyWindow,
				"dls_configuration_poison":      station.DlsConfigurationPoison,
				"dls_configuration_schemaverse": station.DlsConfigurationSchemaverse,
				"total_dls_messages":            totalDlsAmount,
				"tiered_storage_enabled":        station.TieredStorageEnabled,
			}
		}
	}

	shouldSendAnalytics, _ := shouldSendAnalytics()
	if shouldSendAnalytics {
		user, _ := getUserDetailsFromMiddleware(c)
		analytics.SendEvent(user.Username, "user-enter-station-overview")
	}

	c.IndentedJSON(200, response)
}

func (mh MonitoringHandler) GetSystemLogs(c *gin.Context) {
	const amount = 100
	const timeout = 500 * time.Millisecond

	var request models.SystemLogsRequest
	ok := utils.Validate(c, &request, false, nil)
	if !ok {
		return
	}

	startSeq := uint64(request.StartIdx)
	getLast := false
	if request.StartIdx == -1 {
		getLast = true
	}

	filterSubject, filterSubjectSuffix := _EMPTY_, _EMPTY_
	switch request.LogType {
	case "err":
		filterSubjectSuffix = syslogsErrSubject
	case "warn":
		filterSubjectSuffix = syslogsWarnSubject
	case "info":
		filterSubjectSuffix = syslogsInfoSubject
	case "sys":
		filterSubjectSuffix = syslogsSysSubject
	case "external":
		filterSubjectSuffix = syslogsExternalSubject
	}

	if filterSubjectSuffix != _EMPTY_ {
		filterSubject = fmt.Sprintf("%s.%s.%s", syslogsStreamName, "*", filterSubjectSuffix)
	}
	response, err := mh.S.GetSystemLogs(amount, timeout, getLast, startSeq, filterSubject, false)
	if err != nil {
		serv.Errorf("GetSystemLogs: " + err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}

	shouldSendAnalytics, _ := shouldSendAnalytics()
	if shouldSendAnalytics {
		user, _ := getUserDetailsFromMiddleware(c)
		analytics.SendEvent(user.Username, "user-enter-syslogs-page")
	}

	c.IndentedJSON(200, response)
}

func (mh MonitoringHandler) DownloadSystemLogs(c *gin.Context) {
	const timeout = 20 * time.Second
	response, err := mh.S.GetSystemLogs(100, timeout, false, 0, _EMPTY_, true)
	if err != nil {
		serv.Errorf("DownloadSystemLogs: " + err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}

	b := new(bytes.Buffer)
	datawriter := bufio.NewWriter(b)

	for _, log := range response.Logs {
		_, _ = datawriter.WriteString(log.Source + ": " + log.Data + "\n")
	}

	datawriter.Flush()
	c.Writer.Write(b.Bytes())
}

func min(x, y uint64) uint64 {
	if x < y {
		return x
	}
	return y
}

func (s *Server) GetSystemLogs(amount uint64,
	timeout time.Duration,
	fromLast bool,
	lastKnownSeq uint64,
	filterSubject string,
	getAll bool) (models.SystemLogsResponse, error) {
	uid := s.memphis.nuid.Next()
	durableName := "$memphis_fetch_logs_consumer_" + uid
	var msgs []StoredMsg

	streamInfo, err := s.memphisStreamInfo(syslogsStreamName)
	if err != nil {
		return models.SystemLogsResponse{}, err
	}

	amount = min(streamInfo.State.Msgs, amount)
	startSeq := lastKnownSeq - amount + 1

	if getAll {
		startSeq = streamInfo.State.FirstSeq
		amount = streamInfo.State.Msgs
	} else if fromLast {
		startSeq = streamInfo.State.LastSeq - amount + 1

		//handle uint wrap around
		if amount >= streamInfo.State.LastSeq {
			startSeq = 1
		}
		lastKnownSeq = streamInfo.State.LastSeq

	} else if amount >= lastKnownSeq {
		startSeq = 1
		amount = lastKnownSeq
	}

	cc := ConsumerConfig{
		OptStartSeq:   startSeq,
		DeliverPolicy: DeliverByStartSequence,
		AckPolicy:     AckExplicit,
		Durable:       durableName,
	}

	if filterSubject != _EMPTY_ {
		cc.FilterSubject = filterSubject
	}

	err = s.memphisAddConsumer(syslogsStreamName, &cc)
	if err != nil {
		return models.SystemLogsResponse{}, err
	}

	responseChan := make(chan StoredMsg)
	subject := fmt.Sprintf(JSApiRequestNextT, syslogsStreamName, durableName)
	reply := durableName + "_reply"
	req := []byte(strconv.FormatUint(amount, 10))

	sub, err := s.subscribeOnGlobalAcc(reply, reply+"_sid", func(_ *client, subject, reply string, msg []byte) {
		go func(respCh chan StoredMsg, subject, reply string, msg []byte) {
			// ack
			s.sendInternalAccountMsg(s.GlobalAccount(), reply, []byte(_EMPTY_))
			rawTs := tokenAt(reply, 8)
			seq, _, _ := ackReplyInfo(reply)

			intTs, err := strconv.Atoi(rawTs)
			if err != nil {
				s.Errorf("GetSystemLogs: " + err.Error())
			}

			respCh <- StoredMsg{
				Subject:  subject,
				Sequence: uint64(seq),
				Data:     msg,
				Time:     time.Unix(0, int64(intTs)),
			}
		}(responseChan, subject, reply, copyBytes(msg))
	})
	if err != nil {
		return models.SystemLogsResponse{}, err
	}

	s.sendInternalAccountMsgWithReply(s.GlobalAccount(), subject, reply, nil, req, true)

	timer := time.NewTimer(timeout)
	for i := uint64(0); i < amount; i++ {
		select {
		case <-timer.C:
			goto cleanup
		case msg := <-responseChan:
			msgs = append(msgs, msg)
		}
	}

cleanup:
	timer.Stop()
	s.unsubscribeOnGlobalAcc(sub)
	err = s.memphisRemoveConsumer(syslogsStreamName, durableName)
	if err != nil {
		return models.SystemLogsResponse{}, err
	}

	var resMsgs []models.Log
	if uint64(len(msgs)) < amount && streamInfo.State.Msgs > amount && streamInfo.State.FirstSeq < startSeq {
		return s.GetSystemLogs(amount*2, timeout, false, lastKnownSeq, filterSubject, getAll)
	}
	for _, msg := range msgs {
		if err != nil {
			return models.SystemLogsResponse{}, err
		}

		splittedSubj := strings.Split(msg.Subject, tsep)
		var (
			logSource string
			logType   string
		)

		if len(splittedSubj) == 2 {
			// old version's logs
			logSource = "broker"
			logType = splittedSubj[1]
		} else if len(splittedSubj) == 3 {
			// old version's logs
			logSource, logType = splittedSubj[1], splittedSubj[2]
		} else {
			logSource, logType = splittedSubj[1], splittedSubj[3]
		}

		data := string(msg.Data)
		resMsgs = append(resMsgs, models.Log{
			MessageSeq: int(msg.Sequence),
			Type:       logType,
			Data:       data,
			Source:     logSource,
			TimeSent:   msg.Time,
		})
	}

	if getAll {
		sort.Slice(resMsgs, func(i, j int) bool {
			return resMsgs[i].MessageSeq < resMsgs[j].MessageSeq
		})
	} else {
		sort.Slice(resMsgs, func(i, j int) bool {
			return resMsgs[i].MessageSeq > resMsgs[j].MessageSeq
		})

		if len(resMsgs) > 100 {
			resMsgs = resMsgs[:100]
		}
	}

	return models.SystemLogsResponse{Logs: resMsgs}, nil
}

func checkCompStatus(components []models.SysComponent) string {
	status := "healthy"
	yellowCount := 0
	redCount := 0
	for _, component := range components {
		if !component.Healthy {
			redCount++
			continue
		}
		compRedCount := 0
		compYellowCount := 0
		if component.CPU.Percentage > 66 {
			compRedCount++
		} else if component.CPU.Percentage > 33 {
			compYellowCount++
		}
		if component.Memory.Percentage > 66 {
			compRedCount++
		} else if component.Memory.Percentage > 33 {
			compYellowCount++
		}
		if component.Storage.Percentage > 66 {
			compRedCount++
		} else if component.Storage.Percentage > 33 {
			compYellowCount++
		}
		if compRedCount >= 1 {
			redCount++
		} else if compYellowCount > 0 {
			yellowCount++
		}
	}
	redStatus := float64(redCount) / float64(len(components))
	if redStatus >= 0.66 {
		status = "unhealthy"
	} else if redStatus >= 0.33 || yellowCount > 0 {
		status = "risky"
	}

	return status
}

func getDbStorageSize() (float64, float64, error) {
	var configuration = conf.GetConfig()
	sbStats, err := serv.memphis.dbClient.Database(configuration.DB_NAME).RunCommand(context.TODO(), map[string]interface{}{
		"dbStats": 1,
	}).DecodeBytes()
	if err != nil {
		return 0, 0, err
	}

	dbStorageSize := sbStats.Lookup("dataSize").Double() + sbStats.Lookup("indexSize").Double()
	totalSize := sbStats.Lookup("fsTotalSize").Double()
	return dbStorageSize, totalSize, nil
}

func getUnixStorageSize() (float64, error) {
	out, err := exec.Command("df", "-h", "/").Output()
	if err != nil {
		return 0, err
	}
	var storage_size float64
	output := string(out[:])
	splitted_output := strings.Split(output, "\n")
	parsedline := strings.Fields(splitted_output[1])
	if len(parsedline) > 0 {
		stringSize := strings.Split(parsedline[1], "G")
		storage_size, err = strconv.ParseFloat(stringSize[0], 64)
		if err != nil {
			return 0, err
		}
	}
	return storage_size * 1024 * 1024 * 1024, nil
}

func getUnixMemoryUsage() (float64, error) {
	pid := os.Getpid()
	pidStr := strconv.Itoa(pid)
	out, err := exec.Command("ps", "-o", "vsz", "-p", pidStr).Output()
	if err != nil {
		return 0, err
	}
	memUsage := float64(0)
	output := string(out[:])
	splitted_output := strings.Split(output, "\n")
	parsedline := strings.Fields(splitted_output[1])
	if len(parsedline) > 0 {
		memUsage, err = strconv.ParseFloat(parsedline[0], 64)
		if err != nil {
			return 0, err
		}
	}
	return memUsage, nil
}

func defaultSystemComp(compName string, healthy bool) models.SysComponent {
	defaultStat := models.CompStats{
		Total:      0,
		Current:    0,
		Percentage: 0,
	}

	return models.SysComponent{
		Name:    compName,
		CPU:     defaultStat,
		Memory:  defaultStat,
		Storage: defaultStat,
		Healthy: healthy,
	}
}

func getRelevantComponents(name string, components []models.SysComponent, desired int) []models.SysComponent {
	res := []models.SysComponent{}
	for _, comp := range components {
		regexMatch, _ := regexp.MatchString(`^`+name+`-\d*[0-9]\d*$`, comp.Name)
		if regexMatch {
			res = append(res, comp)
		} else if name == "memphis-rest-gateway" {
			if strings.Contains(comp.Name, name) {
				res = append(res, comp)
			}
		}
	}
	missingComps := desired - len(res)
	if missingComps > 0 {
		for i := 0; i < missingComps; i++ {
			res = append(res, defaultSystemComp(name, false))
		}
	}
	return res
}

func getRelevantPorts(name string, portsMap map[string][]int) []int {
	res := []int{}
	mPorts := make(map[int]bool)
	for key, ports := range portsMap {
		if strings.Contains(key, name) {
			for _, port := range ports {
				if !mPorts[port] {
					mPorts[port] = true
					res = append(res, port)
				}
			}
		}
	}
	return res
}

func getContainerStorageUsage(config *rest.Config, mountPath string, container string, pod string) (float64, error) {
	usage := float64(0)
	req := clientset.CoreV1().RESTClient().Post().
		Resource("pods").
		Name(pod).
		Namespace(configuration.K8S_NAMESPACE).
		SubResource("exec")
	req.VersionedParams(&v1.PodExecOptions{
		Container: container,
		Command:   []string{"df", mountPath},
		Stdin:     false,
		Stdout:    true,
		Stderr:    true,
		TTY:       true,
	}, scheme.ParameterCodec)

	exec, err := remotecommand.NewSPDYExecutor(config, "POST", req.URL())
	if err != nil {
		return 0, err
	}

	var stdout, stderr bytes.Buffer
	err = exec.Stream(remotecommand.StreamOptions{
		Stdin:  nil,
		Stdout: &stdout,
		Stderr: &stderr,
	})
	if err != nil {
		return 0, err
	}
	splitted_output := strings.Split(stdout.String(), "\n")
	parsedline := strings.Fields(splitted_output[1])
	if stderr.String() != "" {
		return usage, errors.New(stderr.String())
	}
	if len(parsedline) > 1 {
		usage, err = strconv.ParseFloat(parsedline[2], 64)
		if err != nil {
			return usage, err
		}
		usage = usage * 1024
	}

	return usage, nil
}

func shortenFloat(f float64) float64 {
	// round up very small number
	if f < float64(0.01) && f > float64(0) {
		return float64(0.01)
	}
	// shorten float to 2 decimal places
	return math.Floor(f*100) / 100
}

func (mh MonitoringHandler) GetAvailableReplicas(c *gin.Context) {
	v, err := serv.Varz(nil)
	if err != nil {
		serv.Errorf("GetAvailableReplicas: " + err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}
	c.IndentedJSON(200, gin.H{
		"available_replicas": v.Routes + 1})
}

func checkIsMinikube(labels map[string]string) bool {
	for key := range labels {
		if strings.Contains(strings.ToLower(key), "minikube") {
			return true
		}
	}
	return false
}
