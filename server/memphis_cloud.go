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
	"io"
	"math"
	"memphis/analytics"
	"memphis/conf"
	"memphis/db"
	"memphis/models"
	"memphis/utils"
	"net/http"
	"regexp"
	"runtime"
	"strings"
	"time"

	dockerClient "github.com/docker/docker/client"
	"github.com/gin-contrib/cors"

	"github.com/docker/docker/api/types"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type BillingHandler struct{ S *Server }
type TenantHandler struct{ S *Server }
type LoginSchema struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

func InitializeBillingRoutes(router *gin.RouterGroup, h *Handlers) {
}

func InitializeTenantsRoutes(router *gin.RouterGroup, h *Handlers) {
}

func AddUsrMgmtCloudRoutes(userMgmtRoutes *gin.RouterGroup, userMgmtHandler UserMgmtHandler) {
}

func getStationStorageType(storageType string) string {
	return strings.ToLower(storageType)
}

func GetStationMaxAge(retentionType string, retentionValue int) time.Duration {
	if retentionType == "message_age_sec" && retentionValue > 0 {
		return time.Duration(retentionValue) * time.Second
	}
	return time.Duration(0)
}

func CreateRootUserOnFirstSystemLoad() error {
	password := configuration.ROOT_PASSWORD
	hashedPwd, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.MinCost)
	if err != nil {
		return err
	}
	hashedPwdString := string(hashedPwd)

	created, err := db.UpsertUserUpdatePassword(ROOT_USERNAME, "root", hashedPwdString, "", false, 1, globalAccountName)
	if err != nil {
		return err
	}

	if created && configuration.ANALYTICS == "true" {
		time.AfterFunc(5*time.Second, func() {
			var deviceIdValue string
			installationType := "stand-alone-k8s"
			if serv.JetStreamIsClustered() {
				installationType = "cluster"
				k8sClusterTimestamp, err := getK8sClusterTimestamp()
				if err == nil {
					deviceIdValue = k8sClusterTimestamp
				} else {
					serv.Errorf("Generate host unique id failed: %s", err.Error())
				}
			} else if configuration.DOCKER_ENV == "true" {
				installationType = "stand-alone-docker"
				dockerMacAddress, err := getDockerMacAddress()
				if err == nil {
					deviceIdValue = dockerMacAddress
				} else {
					serv.Errorf("Generate host unique id failed: %s", err.Error())
				}
			}

			param := analytics.EventParam{
				Name:  "installation-type",
				Value: installationType,
			}
			analyticsParams := []analytics.EventParam{param}
			analyticsParams = append(analyticsParams, analytics.EventParam{Name: "device-id", Value: deviceIdValue})
			analytics.SendEventWithParams("", analyticsParams, "installation")

			if configuration.EXPORTER {
				analytics.SendEventWithParams("", analyticsParams, "enable-exporter")
			}
		})
	}

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
			comp := models.SysComponent{
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
			}
			comp.Status = checkPodStatus(comp.CPU.Percentage, comp.Memory.Percentage, comp.Storage.Percentage)
			components = append(components, models.SystemComponents{
				Name:        "memphis",
				Components:  getComponentsStructByOneComp(comp),
				Status:      comp.Status,
				Ports:       []int{mh.S.opts.UiPort, mh.S.opts.Port, mh.S.opts.Websocket.Port, mh.S.opts.HTTPPort},
				DesiredPods: 1,
				ActualPods:  1,
				Hosts:       hosts,
			})
			healthy := false
			restGwComp := defaultSystemComp("memphis-rest-gateway", healthy)
			resp, err := http.Get(fmt.Sprintf("http://localhost:%v/monitoring/getResourcesUtilization", mh.S.opts.RestGwPort))
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
				restGwComp = models.SysComponent{
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
				}
				restGwComp.Status = checkPodStatus(restGwComp.CPU.Percentage, restGwComp.Memory.Percentage, restGwComp.Storage.Percentage)
			}
			actualRestGw := 1
			if !healthy {
				actualRestGw = 0
			}
			components = append(components, models.SystemComponents{
				Name:        "memphis-rest-gateway",
				Components:  getComponentsStructByOneComp(restGwComp),
				Status:      restGwComp.Status,
				Ports:       []int{mh.S.opts.RestGwPort},
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

			body, err := io.ReadAll(containerStats.Body)
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
			if strings.Contains(containerName, "metadata") {
				dbStorageSize, totalSize, err := getDbStorageSize()
				if err != nil {
					return components, metricsEnabled, err
				}
				storageStat = models.CompStats{
					Total:      shortenFloat(totalSize),
					Current:    shortenFloat(dbStorageSize),
					Percentage: int(math.Ceil(float64(dbStorageSize) / float64(totalSize))),
				}
				containerName = strings.TrimPrefix(containerName, "memphis-")
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
			comp := models.SysComponent{
				Name:    containerName,
				CPU:     cpuStat,
				Memory:  memoryStat,
				Storage: storageStat,
				Healthy: true,
			}
			comp.Status = checkPodStatus(comp.CPU.Percentage, comp.Memory.Percentage, comp.Storage.Percentage)
			components = append(components, models.SystemComponents{
				Name:        strings.TrimSuffix(containerName, "-1"),
				Components:  getComponentsStructByOneComp(comp),
				Status:      comp.Status,
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
		comp := models.SysComponent{
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
		}
		comp.Status = checkPodStatus(comp.CPU.Percentage, comp.Memory.Percentage, comp.Storage.Percentage)
		components = append(components, models.SystemComponents{
			Name:        "memphis",
			Components:  getComponentsStructByOneComp(comp),
			Status:      comp.Status,
			Ports:       []int{mh.S.opts.UiPort, mh.S.opts.Port, mh.S.opts.Websocket.Port, mh.S.opts.HTTPPort},
			DesiredPods: 1,
			ActualPods:  1,
			Hosts:       hosts,
		})
		resp, err := http.Get(fmt.Sprintf("http://localhost:%v/monitoring/getResourcesUtilization", mh.S.opts.RestGwPort))
		healthy := false
		restGwComp := defaultSystemComp("memphis-rest-gateway", healthy)
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
			restGwComp := models.SysComponent{
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
			}
			restGwComp.Status = checkPodStatus(restGwComp.CPU.Percentage, restGwComp.Memory.Percentage, restGwComp.Storage.Percentage)
		}
		actualRestGw := 1
		if !healthy {
			actualRestGw = 0
		}
		components = append(components, models.SystemComponents{
			Name:        "memphis-rest-gateway",
			Components:  getComponentsStructByOneComp(restGwComp),
			Status:      restGwComp.Status,
			Ports:       []int{mh.S.opts.RestGwPort},
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

			body, err := io.ReadAll(containerStats.Body)
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
			if strings.Contains(containerName, "metadata") && !strings.Contains(containerName, "coordinator") {
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
			comp := models.SysComponent{
				Name:    containerName,
				CPU:     cpuStat,
				Memory:  memoryStat,
				Storage: storageStat,
				Healthy: true,
			}
			comp.Status = checkPodStatus(comp.CPU.Percentage, comp.Memory.Percentage, comp.Storage.Percentage)
			components = append(components, models.SystemComponents{
				Name:        containerName,
				Components:  getComponentsStructByOneComp(comp),
				Status:      comp.Status,
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
		deploymentsClient := clientset.AppsV1().Deployments(mh.S.opts.K8sNamespace)
		deploymentsList, err := deploymentsClient.List(context.TODO(), metav1.ListOptions{})
		if err != nil {
			return components, metricsEnabled, err
		}

		pods, err := clientset.CoreV1().Pods(mh.S.opts.K8sNamespace).List(context.TODO(), metav1.ListOptions{})
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
			podMetrics, err := metricsclientset.MetricsV1beta1().PodMetricses(mh.S.opts.K8sNamespace).Get(context.TODO(), pod.Name, metav1.GetOptions{})
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
			pvcClient := clientset.CoreV1().PersistentVolumeClaims(mh.S.opts.K8sNamespace)
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
				if strings.Contains(container.Name, "memphis") || strings.Contains(container.Name, "postgresql") {
					for _, mount := range pod.Spec.Containers[0].VolumeMounts {
						if strings.Contains(mount.Name, "memphis") || strings.Contains(mount.Name, "data") { // data is for postgres mount name
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
				if strings.Contains(strings.ToLower(pod.Name), "metadata") {
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
				storageUsage, err = getContainerStorageUsage(config, mountpath, containerForExec, pod.Name, mh.S.opts.K8sNamespace)
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
			comp.Status = checkPodStatus(comp.CPU.Percentage, comp.Memory.Percentage, comp.Storage.Percentage)
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
					status = healthyStatus
				} else {
					status = unhealthyStatus
				}
			}
			if d.Name == "memphis-rest-gateway" {
				if mh.S.opts.RestGwHost != "" {
					hosts = []string{mh.S.opts.RestGwHost}
				}
			} else if d.Name == "memphis" {
				if mh.S.opts.BrokerHost == "" {
					hosts = []string{}
				} else {
					hosts = []string{mh.S.opts.BrokerHost}
				}
				if mh.S.opts.UiHost != "" {
					hosts = append(hosts, mh.S.opts.UiHost)
				}
			} else if strings.Contains(d.Name, "metadata") {
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

		statefulsetsClient := clientset.AppsV1().StatefulSets(mh.S.opts.K8sNamespace)
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
					status = healthyStatus
				} else {
					status = unhealthyStatus
				}
			}
			if s.Name == "memphis-rest-gateway" {
				if mh.S.opts.RestGwHost != "" {
					hosts = []string{mh.S.opts.RestGwHost}
				}
			} else if s.Name == "memphis" {
				if mh.S.opts.BrokerHost == "" {
					hosts = []string{}
				} else {
					hosts = []string{mh.S.opts.BrokerHost}
				}
				if mh.S.opts.UiHost != "" {
					hosts = append(hosts, mh.S.opts.UiHost)
				}
			} else if strings.Contains(s.Name, "metadata") {
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

	logSource := request.LogSource
	if filterSubjectSuffix != _EMPTY_ {
		if request.LogSource != "empty" && request.LogType != "external" {
			filterSubject = fmt.Sprintf("%s.%s.%s", syslogsStreamName, logSource, filterSubjectSuffix)
		} else if request.LogSource != "empty" && request.LogType == "external" {
			filterSubject = fmt.Sprintf("%s.%s.%s.%s", syslogsStreamName, logSource, "extern", ">")
		} else {
			filterSubject = fmt.Sprintf("%s.%s.%s", syslogsStreamName, "*", filterSubjectSuffix)
		}
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

func memphisWSGetSystemLogs(h *Handlers, logLevel, logSource string) (models.SystemLogsResponse, error) {
	const amount = 100
	const timeout = 3 * time.Second
	filterSubjectSuffix := ""
	switch logLevel {
	case "err":
		filterSubjectSuffix = syslogsErrSubject
	case "warn":
		filterSubjectSuffix = syslogsWarnSubject
	case "info":
		filterSubjectSuffix = syslogsInfoSubject
	default:
		filterSubjectSuffix = syslogsExternalSubject
	}

	filterSubject := "$memphis_syslogs.*." + filterSubjectSuffix

	if filterSubjectSuffix != _EMPTY_ {
		if logSource != "empty" && logLevel != "external" {
			filterSubject = fmt.Sprintf("%s.%s.%s", syslogsStreamName, logSource, filterSubjectSuffix)
		} else if logSource != "empty" && logLevel == "external" {
			filterSubject = fmt.Sprintf("%s.%s.%s.%s", syslogsStreamName, logSource, "extern", ">")
		} else {
			filterSubject = fmt.Sprintf("%s.%s.%s", syslogsStreamName, "*", filterSubjectSuffix)
		}
	}
	return h.Monitoring.S.GetSystemLogs(amount, timeout, true, 0, filterSubject, false)
}

func (s *Server) InitializeEventCounter() error {
	return nil
}

func (s *Server) InitializeFirestore() error {
	return nil
}

func (s *Server) UploadTenantUsageToDB() error {
	return nil
}

func IncrementEventCounter(tenantName string, counterType string, amount int64) {}

func (ch ConfigurationsHandler) EditClusterConfig(c *gin.Context) {
	var body models.EditClusterConfigSchema
	ok := utils.Validate(c, &body, false, nil)
	if !ok {
		return
	}
	if ch.S.opts.DlsRetentionHours != body.DlsRetention {
		err := changeDlsRetention(body.DlsRetention)
		if err != nil {
			serv.Errorf("EditConfigurations: " + err.Error())
			c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
			return
		}
	}
	if ch.S.opts.LogsRetentionDays != body.LogsRetention {
		err := changeLogsRetention(body.LogsRetention)
		if err != nil {
			serv.Errorf("EditConfigurations: " + err.Error())
			c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
			return
		}
	}
	if ch.S.opts.TieredStorageUploadIntervalSec != body.TSTimeSec {
		if body.TSTimeSec > 3600 || body.TSTimeSec < 5 {
			serv.Errorf("EditConfigurations: Tiered storage time can't be less than 5 seconds or more than 60 minutes")
			c.AbortWithStatusJSON(SHOWABLE_ERROR_STATUS_CODE, gin.H{"message": "Tiered storage time can't be less than 5 seconds or more than 60 minutes"})
		} else {
			err := changeTSTime(body.TSTimeSec)
			if err != nil {
				serv.Errorf("EditConfigurations: " + err.Error())
				c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
				return
			}
		}
	}

	brokerHost := strings.ToLower(body.BrokerHost)
	if ch.S.opts.BrokerHost != brokerHost {
		err := EditClusterCompHost("broker_host", brokerHost)
		if err != nil {
			serv.Errorf("EditConfigurations: " + err.Error())
			c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
			return
		}
	}

	uiHost := strings.ToLower(body.UiHost)
	if ch.S.opts.UiHost != uiHost {
		err := EditClusterCompHost("ui_host", uiHost)
		if err != nil {
			serv.Errorf("EditConfigurations: " + err.Error())
			c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
			return
		}
	}

	restGWHost := strings.ToLower(body.RestGWHost)
	if ch.S.opts.RestGwHost != restGWHost {
		err := EditClusterCompHost("rest_gw_host", restGWHost)
		if err != nil {
			serv.Errorf("EditConfigurations: " + err.Error())
			c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
			return
		}
	}

	if ch.S.opts.MaxPayload != int32(body.MaxMsgSizeMb) {
		err := changeMaxMsgSize(body.MaxMsgSizeMb)
		if err != nil {
			serv.Errorf("EditConfigurations: " + err.Error())
			c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
			return
		}
	}

	// send signal to reload config
	err := serv.sendInternalAccountMsgWithReply(serv.GlobalAccount(), CONFIGURATIONS_RELOAD_SIGNAL_SUBJ, _EMPTY_, nil, _EMPTY_, true)
	if err != nil {
		serv.Errorf("EditConfigurations: " + err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}

	shouldSendAnalytics, _ := shouldSendAnalytics()
	if shouldSendAnalytics {
		user, _ := getUserDetailsFromMiddleware(c)
		analytics.SendEvent(user.Username, "user-update-cluster-config")
	}

	c.IndentedJSON(200, gin.H{
		"dls_retention":           ch.S.opts.DlsRetentionHours,
		"logs_retention":          ch.S.opts.LogsRetentionDays,
		"broker_host":             ch.S.opts.BrokerHost,
		"ui_host":                 ch.S.opts.UiHost,
		"rest_gw_host":            ch.S.opts.RestGwHost,
		"tiered_storage_time_sec": ch.S.opts.TieredStorageUploadIntervalSec,
		"max_msg_size_mb":         ch.S.opts.MaxPayload / 1024 / 1024,
	})
}

func (ch ConfigurationsHandler) GetClusterConfig(c *gin.Context) {
	shouldSendAnalytics, _ := shouldSendAnalytics()
	if shouldSendAnalytics {
		user, _ := getUserDetailsFromMiddleware(c)
		analytics.SendEvent(user.Username, "user-enter-cluster-config-page")
	}
	c.IndentedJSON(200, gin.H{
		"dls_retention":           ch.S.opts.DlsRetentionHours,
		"logs_retention":          ch.S.opts.LogsRetentionDays,
		"broker_host":             ch.S.opts.BrokerHost,
		"ui_host":                 ch.S.opts.UiHost,
		"rest_gw_host":            ch.S.opts.RestGwHost,
		"tiered_storage_time_sec": ch.S.opts.TieredStorageUploadIntervalSec,
		"max_msg_size_mb":         ch.S.opts.MaxPayload / 1024 / 1024,
	})
}

func SetCors(router *gin.Engine) {
	router.Use(cors.New(cors.Config{
		AllowOriginFunc: func(origin string) bool {
			return true
		},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Length", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		AllowWildcard:    true,
		AllowWebSockets:  true,
		AllowFiles:       true,
	}))
}

func validateTenantName(tenantName string) error {
	return nil
}

func (th TenantHandler) CreateTenant(c *gin.Context) {
	c.IndentedJSON(404, gin.H{})
}

func (umh UserMgmtHandler) Login(c *gin.Context) {
	var body LoginSchema
	ok := utils.Validate(c, &body, false, nil)
	if !ok {
		return
	}

	username := strings.ToLower(body.Username)
	authenticated, user, err := authenticateUser(username, body.Password)
	if err != nil {
		serv.Errorf("Login : User " + body.Username + ": " + err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}
	if !authenticated || user.UserType == "application" {
		c.AbortWithStatusJSON(401, gin.H{"message": "Unauthorized"})
		return
	}

	token, refreshToken, err := CreateTokens(user)
	if err != nil {
		serv.Errorf("Login: User " + body.Username + ": " + err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}

	if !user.AlreadyLoggedIn {
		err = db.UpdateUserAlreadyLoggedIn(user.ID)
		if err != nil {
			serv.Errorf("Login: User " + body.Username + ": " + err.Error())
			c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
			return
		}
	}

	shouldSendAnalytics, _ := shouldSendAnalytics()
	if shouldSendAnalytics {
		analytics.SendEvent(user.Username, "user-login")
	}

	env := "K8S"
	if configuration.DOCKER_ENV != "" || configuration.LOCAL_CLUSTER_ENV {
		env = "docker"
	}
	exist, tenant, err := db.GetTenantByName(user.TenantName)
	if err != nil {
		serv.Errorf("Login: User " + body.Username + ": " + err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}
	if !exist {
		serv.Warnf("Login: User " + body.Username + ": tenant " + user.TenantName + " does not exist")
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}

	domain := ""
	secure := false
	c.SetCookie("jwt-refresh-token", refreshToken, REFRESH_JWT_EXPIRES_IN_MINUTES*60*1000, "/", domain, secure, true)
	c.IndentedJSON(200, gin.H{
		"jwt":                     token,
		"expires_in":              JWT_EXPIRES_IN_MINUTES * 60 * 1000,
		"user_id":                 user.ID,
		"username":                user.Username,
		"user_type":               user.UserType,
		"created_at":              user.CreatedAt,
		"already_logged_in":       user.AlreadyLoggedIn,
		"avatar_id":               user.AvatarId,
		"send_analytics":          shouldSendAnalytics,
		"env":                     env,
		"full_name":               user.FullName,
		"skip_get_started":        user.SkipGetStarted,
		"broker_host":             serv.opts.BrokerHost,
		"rest_gw_host":            serv.opts.RestGwHost,
		"ui_host":                 serv.opts.UiHost,
		"tiered_storage_time_sec": serv.opts.TieredStorageUploadIntervalSec,
		"ws_port":                 serv.opts.Websocket.Port,
		"http_port":               serv.opts.UiPort,
		"clients_port":            serv.opts.Port,
		"rest_gw_port":            serv.opts.RestGwPort,
		"user_pass_based_auth":    configuration.USER_PASS_BASED_AUTH,
		"connection_token":        configuration.CONNECTION_TOKEN,
		"account_id":              tenant.ID,
	})
}

func (umh UserMgmtHandler) AddUser(c *gin.Context) {
	var body models.AddUserSchema
	ok := utils.Validate(c, &body, false, nil)
	if !ok {
		return
	}

	user, err := getUserDetailsFromMiddleware(c)
	if err != nil {
		serv.Errorf("AddUser: " + err.Error())
		c.AbortWithStatusJSON(401, gin.H{"message": "Unauthorized"})
		return
	}

	var subscription, pending bool
	team := strings.ToLower(body.Team)
	position := strings.ToLower(body.Position)
	fullName := strings.ToLower(body.FullName)
	owner := user.Username
	description := strings.ToLower(body.Description)

	if user.TenantName != conf.GlobalAccountName {
		user.TenantName = strings.ToLower(user.TenantName)
	}

	username := strings.ToLower(body.Username)
	usernameError := validateUsername(username)
	if usernameError != nil {
		serv.Warnf("AddUser: " + usernameError.Error())
		c.AbortWithStatusJSON(SHOWABLE_ERROR_STATUS_CODE, gin.H{"message": usernameError.Error()})
		return
	}
	exist, _, err := db.GetUserByUsername(username, user.TenantName)
	if err != nil {
		serv.Errorf("AddUser: User " + body.Username + ": " + err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}
	if exist {
		errMsg := "A user with the name " + body.Username + " already exists"
		serv.Warnf("CreateUser: " + errMsg)
		c.AbortWithStatusJSON(SHOWABLE_ERROR_STATUS_CODE, gin.H{"message": errMsg})
		return
	}

	userType := strings.ToLower(body.UserType)
	userTypeError := validateUserType(userType)
	if userTypeError != nil {
		serv.Warnf("AddUser: " + userTypeError.Error())
		c.AbortWithStatusJSON(SHOWABLE_ERROR_STATUS_CODE, gin.H{"message": userTypeError.Error()})
		return
	}

	var password string
	var avatarId int
	if userType == "management" {
		if body.Password == "" {
			serv.Warnf("AddUser: Password was not provided for user " + username)
			c.AbortWithStatusJSON(SHOWABLE_ERROR_STATUS_CODE, gin.H{"message": "Password was not provided"})
			return
		}

		hashedPwd, err := bcrypt.GenerateFromPassword([]byte(body.Password), bcrypt.MinCost)
		if err != nil {
			serv.Errorf("AddUser: User " + body.Username + ": " + err.Error())
			c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
			return
		}
		password = string(hashedPwd)

		avatarId = 1
		if body.AvatarId > 0 {
			avatarId = body.AvatarId
		}
	}

	var brokerConnectionCreds string
	if userType == "application" {
		fullName = ""
		subscription = false
		pending = false
		if configuration.USER_PASS_BASED_AUTH {
			if body.Password == "" {
				serv.Warnf("AddUser: Password was not provided for user " + username)
				c.AbortWithStatusJSON(SHOWABLE_ERROR_STATUS_CODE, gin.H{"message": "Password was not provided"})
				return
			}
			password, err = EncryptAES([]byte(body.Password))
			if err != nil {
				serv.Errorf("AddUser: User " + body.Username + ": " + err.Error())
				c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
				return
			}
			avatarId = 1
			if body.AvatarId > 0 {
				avatarId = body.AvatarId
			}
		} else {
			brokerConnectionCreds = configuration.CONNECTION_TOKEN
		}
	}
	newUser, err := db.CreateUser(username, userType, password, fullName, subscription, avatarId, user.TenantName, pending, team, position, owner, description)
	if err != nil {
		if strings.Contains(err.Error(), "already exist") {
			serv.Warnf("CreateUserManagement: " + err.Error())
			c.AbortWithStatusJSON(SHOWABLE_ERROR_STATUS_CODE, gin.H{"message": err.Error()})
			return
		}
		serv.Errorf("AddUser: User " + body.Username + ": " + err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}

	shouldSendAnalytics, _ := shouldSendAnalytics()
	if shouldSendAnalytics {
		analytics.SendEvent(user.Username, "user-add-user")
	}

	if userType == "application" && configuration.USER_PASS_BASED_AUTH {
		// send signal to reload config
		err = serv.sendInternalAccountMsgWithReply(serv.GlobalAccount(), CONFIGURATIONS_RELOAD_SIGNAL_SUBJ, _EMPTY_, nil, _EMPTY_, true)
		if err != nil {
			serv.Errorf("AddUser: User " + body.Username + ": " + err.Error())
			c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
			return
		}
	}

	serv.Noticef("User " + username + " has been created")
	c.IndentedJSON(200, gin.H{
		"id":                      newUser.ID,
		"username":                username,
		"user_type":               userType,
		"created_at":              newUser.CreatedAt,
		"already_logged_in":       false,
		"avatar_id":               body.AvatarId,
		"broker_connection_creds": brokerConnectionCreds,
		"position":                newUser.Position,
		"team":                    newUser.Team,
		"pending":                 newUser.Pending,
		"owner":                   newUser.Owner,
		"description":             newUser.Description,
	})
}

func (umh UserMgmtHandler) RemoveUser(c *gin.Context) {
	var body models.RemoveUserSchema
	ok := utils.Validate(c, &body, false, nil)
	if !ok {
		return
	}

	username := strings.ToLower(body.Username)
	user, err := getUserDetailsFromMiddleware(c)
	if err != nil {
		serv.Errorf("RemoveUser: User " + body.Username + ": " + err.Error())
		c.AbortWithStatusJSON(401, gin.H{"message": "Unauthorized"})
		return
	}
	if user.Username == username {
		serv.Warnf("RemoveUser: You can not remove your own user")
		c.AbortWithStatusJSON(SHOWABLE_ERROR_STATUS_CODE, gin.H{"message": "You can not remove your own user"})
		return
	}

	exist, userToRemove, err := db.GetUserByUsername(username, user.TenantName)
	if err != nil {
		serv.Errorf("RemoveUser: User " + body.Username + ": " + err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}
	if !exist {
		serv.Warnf("RemoveUser: User does not exist")
		c.AbortWithStatusJSON(SHOWABLE_ERROR_STATUS_CODE, gin.H{"message": "User does not exist"})
		return
	}
	if userToRemove.UserType == "root" {
		serv.Warnf("RemoveUser: You can not remove the root user")
		c.AbortWithStatusJSON(SHOWABLE_ERROR_STATUS_CODE, gin.H{"message": "You can not remove the root user"})
		return
	}

	err = updateDeletedUserResources(userToRemove)
	if err != nil {
		serv.Errorf("RemoveUser: User " + body.Username + ": " + err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": err.Error()})
		return
	}

	err = db.DeleteUser(username, userToRemove.TenantName)
	if err != nil {
		serv.Errorf("RemoveUser: User " + body.Username + ": " + err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}

	if userToRemove.UserType == "application" && configuration.USER_PASS_BASED_AUTH {
		// send signal to reload config
		err = serv.sendInternalAccountMsgWithReply(serv.GlobalAccount(), CONFIGURATIONS_RELOAD_SIGNAL_SUBJ, _EMPTY_, nil, _EMPTY_, true)
		if err != nil {
			serv.Errorf("RemoveUser: User " + body.Username + ": " + err.Error())
			c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
			return
		}
	}

	shouldSendAnalytics, _ := shouldSendAnalytics()
	if shouldSendAnalytics {
		analytics.SendEvent(user.Username, "user-remove-user")
	}

	serv.Noticef("User " + username + " has been deleted by user " + user.Username)
	c.IndentedJSON(200, gin.H{})
}

func validateUsername(username string) error {
	re := regexp.MustCompile("^[a-z0-9_.-]*$")

	validName := re.MatchString(username)
	if !validName || len(username) == 0 {
		return errors.New("username has to include only letters/numbers/./_/- ")
	}
	return nil
}

func (umh UserMgmtHandler) RemoveMyUser(c *gin.Context) {
	user, err := getUserDetailsFromMiddleware(c)
	if err != nil {
		serv.Errorf("RemoveMyUser: " + err.Error())
		c.AbortWithStatusJSON(401, gin.H{"message": "Unauthorized"})
		return
	}

	if user.UserType != "root" {
		serv.Warnf("RemoveMyUser: Only root user can remove his tenant")
		c.AbortWithStatusJSON(SHOWABLE_ERROR_STATUS_CODE, gin.H{"message": "Only root user can remove his tenant"})
		return
	}

	username := strings.ToLower(user.Username)
	tenantName := user.TenantName
	if user.TenantName != conf.GlobalAccountName {
		user.TenantName = strings.ToLower(user.TenantName)
	}
	err = removeTenantResources(tenantName)
	if err != nil {
		serv.Errorf("RemoveMyUser: User " + username + ": " + err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": err.Error()})
		return
	}

	err = db.DeleteUser(user.Username, user.TenantName)
	if err != nil {
		serv.Errorf("RemoveMyUser: User " + user.Username + ": " + err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}

	shouldSendAnalytics, _ := shouldSendAnalytics()
	if shouldSendAnalytics {
		analytics.SendEvent(user.Username, "user-remove-himself")
	}

	serv.Noticef("Tenant " + user.TenantName + " has been deleted")
	c.IndentedJSON(200, gin.H{})
}

func (s *Server) ConnectToFirebaseFunction() {
}
