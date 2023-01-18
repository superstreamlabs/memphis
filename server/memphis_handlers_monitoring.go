// Copyright 2022-2023 The Memphis.dev Authors
// Licensed under the Memphis Business Source License 1.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// Changed License: [Apache License, Version 2.0 (https://www.apache.org/licenses/LICENSE-2.0), as published by the Apache Foundation.
//
// https://github.com/memphisdev/memphis-broker/blob/master/LICENSE
//
// Additional Use Grant: You may make use of the Licensed Work (i) only as part of your own product or service, provided it is not a message broker or a message queue product or service; and (ii) provided that you do not use, provide, distribute, or make available the Licensed Work as a Service.
// A "Service" is a commercial offering, product, hosted, or managed service, that allows third parties (other than your own employees and contractors acting on your behalf) to access and/or use the Licensed Work or a substantial set of the features or functionality of the Licensed Work to third parties as a software-as-a-service, platform-as-a-service, infrastructure-as-a-service or other similar services that compete with Licensor products or services.
package server

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math"
	"memphis-broker/analytics"
	"memphis-broker/conf"
	"memphis-broker/models"
	"memphis-broker/utils"
	"net/http"
	"os/exec"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/docker/docker/api/types"
	dockerClient "github.com/docker/docker/client"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	metricsv "k8s.io/metrics/pkg/client/clientset/versioned"
)

type MonitoringHandler struct{ S *Server }

var clientset *kubernetes.Clientset
var metricsclientset *metricsv.Clientset

func clientSetClusterConfig() error {
	var config *rest.Config
	var err error
	// in cluster config
	config, err = rest.InClusterConfig()
	if err != nil {
		serv.Errorf("clientSetClusterConfig: InClusterConfig: " + err.Error())
		return err
	}

	clientset, err = kubernetes.NewForConfig(config)
	if err != nil {
		serv.Errorf("clientSetClusterConfig: NewForConfig: " + err.Error())
		return err
	}
	if metricsclientset == nil {
		metricsclientset, err = metricsv.NewForConfig(config)
		if err != nil {
			serv.Errorf("clientSetClusterConfig: metricsclientset: " + err.Error())
			return err
		}
	}

	return nil
}

func (mh MonitoringHandler) GetSystemComponents() ([]models.SystemComponents, error) {
	var components []models.SystemComponents
	var dbComponents []models.SysComponent
	var dbPorts []int
	var brokerComponents []models.SysComponent
	var brokerPorts []int
	var proxyComponents []models.SysComponent
	var proxyPorts []int
	dbActual := 0
	dbDesired := 0
	brokerActual := 0
	brokerDesired := 0
	proxyActual := 0
	proxyDesired := 0
	dbPodIp := ""
	proxyPodIp := ""
	brokerPodIp := ""
	defaultStat := models.CompStats{
		Max:        0,
		Current:    0,
		Percentage: 0,
	}
	if configuration.DOCKER_ENV != "" { // docker env
		var rt runtime.MemStats
		runtime.ReadMemStats(&rt)
		if configuration.DEV_ENV == "true" {
			var storage_size float64
			os := runtime.GOOS
			switch os {
			case "windows":
				storage_size = 500 // TODO: add support
			default:
				size, err := getUnixStorageSize()
				if err != nil {
					return components, err
				}
				storage_size = size
			}
			maxCpu := runtime.GOMAXPROCS(0)
			v, err := serv.Varz(nil)
			if err != nil {
				return components, err
			}
			brokerComponents = append(brokerComponents, models.SysComponent{
				Name: "memphis-broker",
				CPU: models.CompStats{
					Max:        float64(maxCpu),
					Current:    float64(v.CPU/100) * float64(maxCpu),
					Percentage: math.Ceil(v.CPU),
				},
				Memory: models.CompStats{
					Max:        float64(v.JetStream.Config.MaxMemory),
					Current:    float64(v.JetStream.Stats.Memory),
					Percentage: math.Ceil(float64(v.JetStream.Stats.Memory) / float64(v.JetStream.Config.MaxMemory)),
				},
				Storage: models.CompStats{
					Max:        storage_size * 1024 * 1024 * 1024,
					Current:    float64(v.JetStream.Stats.Store),
					Percentage: math.Ceil(float64(v.JetStream.Stats.Store/1024/1024/1024) / storage_size),
				},
				Connected: true,
			})
			brokerPorts = []int{9000, 6666, 7770, 8222}
			httpProxy := "http://localhost:4444"
			resp, err := http.Get(httpProxy + "/dev/getSystemInfo")
			con := true
			if err != nil {
				con = false
				proxyComponents = append(proxyComponents, models.SysComponent{
					Name:      "memphis-http-proxy",
					CPU:       defaultStat,
					Memory:    defaultStat,
					Storage:   defaultStat,
					Connected: con,
				})
			} else {
				var proxyDevInfo models.DevSystemInfoResponse
				defer resp.Body.Close()
				err = json.NewDecoder(resp.Body).Decode(&proxyDevInfo)
				if err != nil {
					return components, err
				}
				proxyComponents = append(proxyComponents, models.SysComponent{
					Name: "memphis-http-proxy",
					CPU: models.CompStats{
						Max:        float64(maxCpu),
						Current:    float64(proxyDevInfo.CPU/100) * float64(maxCpu),
						Percentage: math.Ceil(proxyDevInfo.CPU),
					},
					Memory: models.CompStats{
						Max:        float64(v.JetStream.Config.MaxMemory),
						Current:    float64(proxyDevInfo.Memory/100) * float64(v.JetStream.Config.MaxMemory),
						Percentage: math.Ceil(float64(proxyDevInfo.Memory)),
					},
					Storage:   defaultStat,
					Connected: con,
				})
			}
			proxyPorts = []int{4444}
		}

		ctx := context.Background()
		dockerCli, err := dockerClient.NewClientWithOpts(dockerClient.FromEnv)
		if err != nil {
			return components, err
		}
		containers, err := dockerCli.ContainerList(ctx, types.ContainerListOptions{})
		if err != nil {
			return components, err
		}

		for _, container := range containers {
			continerName := container.Names[0]
			if container.State != "running" {
				comp := models.SysComponent{
					Name:      continerName,
					CPU:       defaultStat,
					Memory:    defaultStat,
					Storage:   defaultStat,
					Connected: false,
				}
				if strings.Contains(continerName, "mongo") {
					dbComponents = append(dbComponents, comp)
				} else if strings.Contains(continerName, "broker") {
					brokerComponents = append(brokerComponents, comp)
				} else if strings.Contains(continerName, "proxy") {
					proxyComponents = append(proxyComponents, comp)
				}
				continue
			}
			stats, err := dockerCli.ContainerStats(ctx, container.ID, false)
			if err != nil {
				return components, err
			}
			defer stats.Body.Close()

			body, err := ioutil.ReadAll(stats.Body)
			if err != nil {
				return components, err
			}
			var statsType types.Stats
			err = json.Unmarshal(body, &statsType)
			if err != nil {
				return components, err
			}
			cpuLimit := float64(runtime.GOMAXPROCS(0))
			cpuPercentage := math.Ceil((float64(statsType.CPUStats.CPUUsage.TotalUsage) / float64(statsType.CPUStats.SystemUsage)) * 100)
			totalCpuUsage := (float64(cpuPercentage) / 100) * cpuLimit
			totalMemoryUsage := float64(statsType.MemoryStats.Usage)
			memoryLimit := float64(statsType.MemoryStats.Limit)
			memoryPercentage := math.Ceil((totalMemoryUsage / memoryLimit) * 100)
			storage_size, err := getUnixStorageSize()
			if err != nil {
				return components, err
			}

			if strings.Contains(continerName, "mongo") {
				dbStorageSize, totalSize, err := getDbStorageSize()
				if err != nil {
					return components, err
				}
				dbComponents = append(dbComponents, models.SysComponent{
					Name: continerName,
					CPU: models.CompStats{
						Max:        cpuLimit,
						Current:    totalCpuUsage,
						Percentage: cpuPercentage,
					},
					Memory: models.CompStats{
						Max:        memoryLimit,
						Current:    totalMemoryUsage,
						Percentage: memoryPercentage,
					},
					Storage: models.CompStats{
						Max:        totalSize,
						Current:    dbStorageSize,
						Percentage: math.Ceil(dbStorageSize / totalSize),
					},
					Connected: true,
				})
				for _, port := range container.Ports {
					dbPorts = append(dbPorts, int(port.PublicPort))
				}
			} else if strings.Contains(continerName, "broker") {
				v, err := serv.Varz(nil)
				if err != nil {
					return components, err
				}
				brokerComponents = append(brokerComponents, models.SysComponent{
					Name: continerName,
					CPU: models.CompStats{
						Max:        cpuLimit,
						Current:    totalCpuUsage,
						Percentage: cpuPercentage,
					},
					Memory: models.CompStats{
						Max:        memoryLimit,
						Current:    totalMemoryUsage,
						Percentage: memoryPercentage,
					},
					Storage: models.CompStats{
						Max:        storage_size * 1024 * 1024 * 1024,
						Current:    float64(v.JetStream.Stats.Store),
						Percentage: math.Ceil(float64(v.JetStream.Stats.Store/1024/1024/1024) / storage_size),
					},
					Connected: true,
				})
				for _, port := range container.Ports {
					brokerPorts = append(brokerPorts, int(port.PublicPort))
				}
			} else if strings.Contains(continerName, "proxy") {
				for _, port := range container.Ports {
					proxyPorts = append(proxyPorts, int(port.PublicPort))
				}
				if err != nil {
					proxyComponents = append(proxyComponents, models.SysComponent{
						Name: continerName,
						CPU: models.CompStats{
							Max:        cpuLimit,
							Current:    totalCpuUsage,
							Percentage: cpuPercentage,
						},
						Memory: models.CompStats{
							Max:        memoryLimit,
							Current:    totalMemoryUsage,
							Percentage: memoryPercentage,
						},
						Storage:   defaultStat,
						Connected: false,
					})
					continue
				}
			}
		}

		if dbComponents[0].Connected {
			dbActual = 1
		}
		components = append(components, models.SystemComponents{
			Name:        dbComponents[0].Name,
			Components:  dbComponents,
			Status:      checkCompStatus(dbComponents),
			Ports:       removeDuplicatePorts(dbPorts),
			DesiredPods: 1,
			ActualPods:  dbActual,
			Address:     "http://localhost",
		})

		if brokerComponents[0].Connected {
			brokerActual = 1
		}
		components = append(components, models.SystemComponents{
			Name:        brokerComponents[0].Name,
			Components:  brokerComponents,
			Status:      checkCompStatus(brokerComponents),
			Ports:       removeDuplicatePorts(brokerPorts),
			DesiredPods: 1,
			ActualPods:  brokerActual,
			Address:     "http://localhost",
		})

		if proxyComponents[0].Connected {
			proxyActual = 1
		}
		components = append(components, models.SystemComponents{
			Name:        proxyComponents[0].Name,
			Components:  proxyComponents,
			Status:      checkCompStatus(proxyComponents),
			Ports:       removeDuplicatePorts(proxyPorts),
			DesiredPods: 1,
			ActualPods:  proxyActual,
			Address:     "http://localhost",
		})
	} else { // k8s env
		if clientset == nil {
			err := clientSetClusterConfig()
			if err != nil {
				return components, err
			}
		}
		deploymentsClient := clientset.AppsV1().Deployments(configuration.K8S_NAMESPACE)
		deploymentsList, err := deploymentsClient.List(context.TODO(), metav1.ListOptions{})
		if err != nil {
			return components, err
		}
		// podMetrics, err := metricsclientset.MetricsV1beta1().PodMetricses(configuration.K8S_NAMESPACE).List(context.TODO(), metav1.ListOptions{})

		// for _, m := range podMetrics.Items{

		// for _, container := range m.Containers{
		// cpu := container.Usage.Cpu()
		// memory := container.Usage.Memory()
		// storage := container.Usage.Storage()
		// storageEphermal := container.Usage.StorageEphemeral()

		// }
		// }

		pods, err := clientset.CoreV1().Pods(configuration.K8S_NAMESPACE).List(context.TODO(), metav1.ListOptions{})
		if err != nil {
			return components, err
		}

		for _, pod := range pods.Items {
			serv.Noticef("pod loop")
			fmt.Println(pod.Name)
			// for _, d := range deploymentsList.Items {
			var ports []int
			podMetrics, err := metricsclientset.MetricsV1beta1().PodMetricses(configuration.K8S_NAMESPACE).Get(context.TODO(), pod.Name, metav1.GetOptions{})
			if err != nil {
				serv.Errorf("podMetrics: " + err.Error())
				return components, err
			}
			// pod1, err := clientset.CoreV1().Pods(configuration.K8S_NAMESPACE).Get(context.TODO(), pod.Name, metav1.GetOptions{})
			// if err != nil {
			// 	return components, err
			// }
			// pod1.Spec.Containers[0].Resources.Limits.Cpu().Value()
			// re := regexp.MustCompile("[0-9]+")
			// number := re.FindAllString(pod1.Spec.Containers[0].Resources.Limits.Cpu().String(), -1)
			// numberFloat, _ := strconv.ParseFloat(number[0], 64)
			// cpuLimit := numberFloat
			// number = re.FindAllString(pod1.Spec.Containers[0].Resources.Limits.Memory().String(), -1)
			// numberFloat, _ = strconv.ParseFloat(number[0], 64)
			// memLimit := numberFloat
			// number = re.FindAllString(pod1.Spec.Containers[0].Resources.Limits.Storage().String(), -1)
			// numberFloat, _ = strconv.ParseFloat(number[0], 64)
			// storageLimit := numberFloat
			cpuLimit := pod.Spec.Containers[0].Resources.Limits.Cpu().AsApproximateFloat64()
			if cpuLimit == float64(0) {
				node, err := clientset.CoreV1().Nodes().Get(context.TODO(), pod.Spec.NodeName, metav1.GetOptions{})
				if err != nil {
					serv.Errorf("nodes: " + err.Error())
					return components, err
				}
				cpuLimit = node.Status.Capacity.Cpu().AsApproximateFloat64()
			}
			// cpuLimit := float64(0)
			memLimit := pod.Spec.Containers[0].Resources.Limits.Memory().AsApproximateFloat64()
			if memLimit == float64(0) {
				node, err := clientset.CoreV1().Nodes().Get(context.TODO(), pod.Spec.NodeName, metav1.GetOptions{})
				if err != nil {
					serv.Errorf("nodes: " + err.Error())
					return components, err
				}
				memLimit = node.Status.Capacity.Memory().AsApproximateFloat64()
			}
			// memLimit := float64(0)
			// storageLimit := pod.Spec.Containers[0].Resources.Limits.Storage().AsApproximateFloat64()
			// storageLimit := float64(0)
			cpuUsage := float64(0)
			memUsage := float64(0)
			// storageUsage := float64(0)
			for _, container := range podMetrics.Containers {
				// re := regexp.MustCompile("[0-9]+")
				// numbers := re.FindAllString(container.Usage.Cpu().String(), -1)
				// usage, _ := strconv.ParseFloat(numbers[0], 64)
				// cpuUsage += usage
				// cpuLimit += float64(pod.Spec.Containers[i].Resources.Limits.Cpu().Value())
				cpuUsage += float64(container.Usage.Cpu().Value())
				// numbers = re.FindAllString(container.Usage.Memory().String(), -1)
				// usage, _ = strconv.ParseFloat(numbers[0], 64)
				// memUsage += usage
				memUsage += float64(container.Usage.Memory().Value())

				// memLimit += float64(pod.Spec.Containers[i].Resources.Limits.Memory().Value())
				// numbers = re.FindAllString(container.Usage.Storage().String(), -1)
				// usage, _ = strconv.ParseFloat(numbers[0], 64)
				// storageUsage += usage
				// storageUsage += float64(container.Usage.Storage().Value())
				// storageLimit += float64(pod.Spec.Containers[i].Resources.Limits.Storage().Value())
			}
			for _, container := range pod.Spec.Containers {
				for _, port := range container.Ports {
					ports = append(ports, int(port.ContainerPort))
				}
			}
			comp := models.SysComponent{
				Name: pod.Name,
				CPU: models.CompStats{
					Max:        cpuLimit,
					Current:    cpuUsage,
					Percentage: math.Ceil((cpuUsage / cpuLimit) * 100),
				},
				Memory: models.CompStats{
					Max:        memLimit,
					Current:    memUsage,
					Percentage: math.Ceil((memUsage / memLimit) * 100),
				},
				// Storage: models.CompStats{
				// 	Max:        storageLimit,
				// 	Current:    storageUsage,
				// 	Percentage: math.Ceil((storageUsage / storageLimit) * 100),
				// },
				Connected: true,
			}
			serv.Noticef(pod.Name + " CPU: " + fmt.Sprintf("%f", math.Ceil((cpuUsage/cpuLimit)*100)) + "%/" + fmt.Sprintf("%f", cpuUsage) + " usage/" + fmt.Sprintf("%f", cpuLimit) + " limit")
			serv.Noticef(pod.Name + " Memory: " + fmt.Sprintf("%f", math.Ceil((memUsage/memLimit)*100)) + "%/" + fmt.Sprintf("%f", memUsage) + " usage/" + fmt.Sprintf("%f", memLimit) + " limit")
			// serv.Noticef(pod.Name + " Storage: " + fmt.Sprintf("%f", math.Ceil((storageUsage/storageLimit)*100)) + "%/" + fmt.Sprintf("%f", storageUsage) + " usage/" + fmt.Sprintf("%f", storageLimit) + " limit")
			if strings.Contains(pod.Name, "mongo") {
				dbComponents = append(dbComponents, comp)
				dbPorts = ports
				// dbDesired = int(*d.Spec.Replicas)
				// dbActual = int(d.Status.ReadyReplicas)
				dbPodIp = pod.Status.PodIP
			} else if strings.Contains(pod.Name, "broker") {
				brokerComponents = append(brokerComponents, comp)
				brokerPorts = ports
				// brokerDesired = int(*d.Spec.Replicas)
				// brokerActual = int(d.Status.ReadyReplicas)
				brokerPodIp = pod.Status.PodIP
			} else if strings.Contains(pod.Name, "proxy") {
				proxyComponents = append(proxyComponents, comp)
				proxyPorts = ports
				// proxyDesired = int(*d.Spec.Replicas)
				// proxyActual = int(d.Status.ReadyReplicas)
				proxyPodIp = pod.Status.PodIP
			}
		}

		for _, d := range deploymentsList.Items {
			if strings.Contains(d.Name, "mongo") {
				dbDesired = int(*d.Spec.Replicas)
				dbActual = int(d.Status.ReadyReplicas)
			} else if strings.Contains(d.Name, "broker") {
				brokerDesired = int(*d.Spec.Replicas)
				brokerActual = int(d.Status.ReadyReplicas)
			} else if strings.Contains(d.Name, "proxy") {
				proxyDesired = int(*d.Spec.Replicas)
				proxyActual = int(d.Status.ReadyReplicas)
			}
		}

		statefulsetsClient := clientset.AppsV1().StatefulSets(configuration.K8S_NAMESPACE)
		statefulsetsList, err := statefulsetsClient.List(context.TODO(), metav1.ListOptions{})
		if err != nil {
			return components, err
		}
		for _, s := range statefulsetsList.Items {
			if strings.Contains(s.Name, "mongo") {
				dbDesired = int(*s.Spec.Replicas)
				dbActual = int(s.Status.ReadyReplicas)
			} else if strings.Contains(s.Name, "broker") {
				brokerDesired = int(*s.Spec.Replicas)
				brokerActual = int(s.Status.ReadyReplicas)
			} else if strings.Contains(s.Name, "proxy") {
				proxyDesired = int(*s.Spec.Replicas)
				proxyActual = int(s.Status.ReadyReplicas)
			}

			// var ports []int
			// podMetrics, err := metricsclientset.MetricsV1beta1().PodMetricses(configuration.K8S_NAMESPACE).Get(context.TODO(), s.Name, metav1.GetOptions{})
			// if err != nil {
			// 	return components, err
			// }
			// serv.Noticef("statefulsetsList: " + podMetrics.)
			// pod, err := clientset.CoreV1().Pods(configuration.K8S_NAMESPACE).Get(context.TODO(), s.Name, metav1.GetOptions{})
			// if err != nil {
			// 	return components, err
			// }

			// cpuLimit := float64(pod.Spec.Containers[0].Resources.Limits.Cpu().Value())
			// memLimit := float64(pod.Spec.Containers[0].Resources.Limits.Memory().Value())
			// storageLimit := float64(pod.Spec.Containers[0].Resources.Limits.Storage().Value())
			// cpuUsage := float64(0)
			// memUsage := float64(0)
			// storageUsage := float64(0)
			// for _, container := range podMetrics.Containers {
			// 	cpuUsage += float64(container.Usage.Cpu().Value())
			// 	memUsage += float64(container.Usage.Memory().Value())
			// 	storageUsage += float64(container.Usage.Storage().Value())
			// }
			// for _, container := range s.Spec.Template.Spec.Containers {
			// 	for _, port := range container.Ports {
			// 		ports = append(ports, int(port.ContainerPort))
			// 	}
			// }
			// comp := models.SysComponent{
			// 	Name: pod.Name,
			// 	CPU: models.CompStats{
			// 		Max:        cpuLimit,
			// 		Current:    cpuUsage,
			// 		Percentage: math.Ceil(cpuUsage / cpuLimit),
			// 	},
			// 	Memory: models.CompStats{
			// 		Max:        memLimit,
			// 		Current:    memUsage,
			// 		Percentage: math.Ceil(memUsage / memLimit),
			// 	},
			// 	Storage: models.CompStats{
			// 		Max:        storageLimit,
			// 		Current:    storageUsage,
			// 		Percentage: math.Ceil(storageUsage / storageLimit),
			// 	},
			// 	Connected: true,
			// }
			// serv.Noticef(pod.Name + " CPU: " + fmt.Sprintf("%f", math.Ceil(cpuUsage/cpuLimit)) + "%/" + fmt.Sprintf("%f", cpuUsage) + " usage/" + fmt.Sprintf("%f", cpuLimit) + " limit")
			// serv.Noticef(pod.Name + " Memory: " + fmt.Sprintf("%f", math.Ceil(memUsage/memLimit)) + "%/" + fmt.Sprintf("%f", memUsage) + " usage/" + fmt.Sprintf("%f", memLimit) + " limit")
			// serv.Noticef(pod.Name + " Storage: " + fmt.Sprintf("%f", math.Ceil(storageUsage/storageLimit)) + "%/" + fmt.Sprintf("%f", storageUsage) + " usage/" + fmt.Sprintf("%f", storageLimit) + " limit")
			// if strings.Contains(pod.Name, "mongo") {
			// 	dbComponents = append(dbComponents, comp)
			// 	dbPorts = ports
			// 	dbDesired = int(*s.Spec.Replicas)
			// 	dbActual = int(s.Status.ReadyReplicas)
			// 	dbPodIp = pod.Status.PodIP
			// } else if strings.Contains(pod.Name, "broker") {
			// 	brokerComponents = append(brokerComponents, comp)
			// 	brokerPorts = ports
			// 	brokerDesired = int(*s.Spec.Replicas)
			// 	brokerActual = int(s.Status.ReadyReplicas)
			// 	brokerPodIp = pod.Status.PodIP
			// } else if strings.Contains(pod.Name, "proxy") {
			// 	proxyComponents = append(proxyComponents, comp)
			// 	proxyPorts = ports
			// 	proxyDesired = int(*s.Spec.Replicas)
			// 	proxyActual = int(s.Status.ReadyReplicas)
			// 	proxyPodIp = pod.Status.PodIP
			// }
		}
		if len(proxyComponents) > 0 {
			components = append(components, models.SystemComponents{
				Name:        proxyComponents[0].Name,
				Components:  proxyComponents,
				Status:      checkCompStatus(proxyComponents),
				Ports:       removeDuplicatePorts(proxyPorts),
				DesiredPods: proxyDesired,
				ActualPods:  proxyActual,
				Address:     proxyPodIp,
			})
		}
		if len(dbComponents) > 0 {
			components = append(components, models.SystemComponents{
				Name:        dbComponents[0].Name,
				Components:  dbComponents,
				Status:      checkCompStatus(dbComponents),
				Ports:       removeDuplicatePorts(dbPorts),
				DesiredPods: dbDesired,
				ActualPods:  dbActual,
				Address:     dbPodIp,
			})
		}
		if len(brokerComponents) > 0 {
			components = append(components, models.SystemComponents{
				Name:        brokerComponents[0].Name,
				Components:  brokerComponents,
				Status:      checkCompStatus(brokerComponents),
				Ports:       removeDuplicatePorts(brokerPorts),
				DesiredPods: brokerDesired,
				ActualPods:  brokerActual,
				Address:     brokerPodIp,
			})
		}
	}

	return components, nil
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

func (mh MonitoringHandler) GetMainOverviewData(c *gin.Context) {
	stationsHandler := StationsHandler{S: mh.S}
	stations, err := stationsHandler.GetAllStationsDetails()
	if err != nil {
		serv.Errorf("GetMainOverviewData: GetAllStationsDetails: " + err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}
	totalMessages, err := stationsHandler.GetTotalMessagesAcrossAllStations()
	if err != nil {
		serv.Errorf("GetMainOverviewData: GetTotalMessagesAcrossAllStations: " + err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}
	systemComponents, err := mh.GetSystemComponents()
	if err != nil {
		serv.Errorf("GetMainOverviewData: GetSystemComponents: " + err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}
	k8sEnv := true
	if configuration.DOCKER_ENV != "" {
		k8sEnv = false
	}
	response := models.MainOverviewData{
		TotalStations:    len(stations),
		TotalMessages:    totalMessages,
		SystemComponents: systemComponents,
		Stations:         stations,
		K8sEnv:           k8sEnv,
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
	schemasHandler := SchemasHandler{S: mh.S}
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
	exist, station, err := IsStationExist(stationName)
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

	connectedProducers, disconnectedProducers, deletedProducers := make([]models.ExtendedProducer, 0), make([]models.ExtendedProducer, 0), make([]models.ExtendedProducer, 0)
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
		serv.Errorf("GetStationOverviewData: At station " + body.StationName + ": " + err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}
	avgMsgSize, err := stationsHandler.GetAvgMsgSize(station)
	if err != nil {
		serv.Errorf("GetStationOverviewData: At station " + body.StationName + ": " + err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}

	messagesToFetch := 1000
	messages, err := stationsHandler.GetMessages(station, messagesToFetch)
	if err != nil {
		serv.Errorf("GetStationOverviewData: At station " + body.StationName + ": " + err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}

	poisonMessages, schemaFailedMessages, totalDlsAmount, poisonCgMap, err := poisonMsgsHandler.GetDlsMsgsByStationLight(station)
	if err != nil {
		serv.Errorf("GetStationOverviewData: At station " + body.StationName + ": " + err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
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

	tags, err := tagsHandler.GetTagsByStation(station.ID)
	if err != nil {
		serv.Errorf("GetStationOverviewData: At station " + body.StationName + ": " + err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}
	leader, followers, err := stationsHandler.GetLeaderAndFollowers(station)
	if err != nil {
		serv.Errorf("GetStationOverviewData: At station " + body.StationName + ": " + err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}

	var emptySchemaDetailsObj models.SchemaDetails
	var response gin.H

	// Check when the schema object in station is not empty, not optional for non native stations
	if station.Schema != emptySchemaDetailsObj {
		var schema models.Schema
		err = schemasCollection.FindOne(context.TODO(), bson.M{"name": station.Schema.SchemaName}).Decode(&schema)
		if err != nil {
			serv.Errorf("GetStationOverviewData: At station " + body.StationName + ": " + err.Error())
			c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
			return
		}

		schemaVersion, err := schemasHandler.GetSchemaVersion(station.Schema.VersionNumber, schema.ID)
		if err != nil {
			serv.Errorf("GetStationOverviewData: At station " + body.StationName + ": " + err.Error())
			c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
			return
		}
		updatesAvailable := !schemaVersion.Active
		schemaDetails := models.StationOverviewSchemaDetails{SchemaName: schema.Name, VersionNumber: station.Schema.VersionNumber, UpdatesAvailable: updatesAvailable}

		response = gin.H{
			"connected_producers":      connectedProducers,
			"disconnected_producers":   disconnectedProducers,
			"deleted_producers":        deletedProducers,
			"connected_cgs":            connectedCgs,
			"disconnected_cgs":         disconnectedCgs,
			"deleted_cgs":              deletedCgs,
			"total_messages":           totalMessages,
			"average_message_size":     avgMsgSize,
			"audit_logs":               auditLogs,
			"messages":                 messages,
			"poison_messages":          poisonMessages,
			"schema_failed_messages":   schemaFailedMessages,
			"tags":                     tags,
			"leader":                   leader,
			"followers":                followers,
			"schema":                   schemaDetails,
			"idempotency_window_in_ms": station.IdempotencyWindow,
			"dls_configuration":        station.DlsConfiguration,
			"total_dls_messages":       totalDlsAmount,
		}
	} else {
		var emptyResponse struct{}
		if !station.IsNative {
			cp, dp, cc, dc := getFakeProdsAndConsForPreview()
			response = gin.H{
				"connected_producers":      cp,
				"disconnected_producers":   dp,
				"deleted_producers":        deletedProducers,
				"connected_cgs":            cc,
				"disconnected_cgs":         dc,
				"deleted_cgs":              deletedCgs,
				"total_messages":           totalMessages,
				"average_message_size":     avgMsgSize,
				"audit_logs":               auditLogs,
				"messages":                 messages,
				"poison_messages":          poisonMessages,
				"schema_failed_messages":   schemaFailedMessages,
				"tags":                     tags,
				"leader":                   leader,
				"followers":                followers,
				"schema":                   emptyResponse,
				"idempotency_window_in_ms": station.IdempotencyWindow,
				"dls_configuration":        station.DlsConfiguration,
				"total_dls_messages":       totalDlsAmount,
			}
		} else {
			response = gin.H{
				"connected_producers":      connectedProducers,
				"disconnected_producers":   disconnectedProducers,
				"deleted_producers":        deletedProducers,
				"connected_cgs":            connectedCgs,
				"disconnected_cgs":         disconnectedCgs,
				"deleted_cgs":              deletedCgs,
				"total_messages":           totalMessages,
				"average_message_size":     avgMsgSize,
				"audit_logs":               auditLogs,
				"messages":                 messages,
				"poison_messages":          poisonMessages,
				"schema_failed_messages":   schemaFailedMessages,
				"tags":                     tags,
				"leader":                   leader,
				"followers":                followers,
				"schema":                   emptyResponse,
				"idempotency_window_in_ms": station.IdempotencyWindow,
				"dls_configuration":        station.DlsConfiguration,
				"total_dls_messages":       totalDlsAmount,
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

func removeDuplicatePorts(ports []int) []int {
	res := []int{}
	mPorts := make(map[int]bool)
	for _, port := range ports {
		if !mPorts[port] {
			mPorts[port] = true
			res = append(res, port)
		}
	}
	return res
}

func checkCompStatus(components []models.SysComponent) string {
	status := "green"
	yellowCount := 0
	redCount := 0
	for _, component := range components {
		if !component.Connected {
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
		if compRedCount >= 2 {
			redCount++
		} else if compRedCount == 1 {
			yellowCount++
		} else if compYellowCount > 0 {
			yellowCount++
		}
	}
	redStatus := float64(redCount / len(components))
	if redStatus >= 0.66 {
		status = "red"
	} else if redStatus >= 0.33 || yellowCount > 0 {
		status = "yellow"
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
		stringSize := strings.Split(parsedline[1], "Gi")
		storage_size, _ = strconv.ParseFloat(stringSize[0], 64)
	}
	return storage_size, nil
}
