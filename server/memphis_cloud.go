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
	"math/rand"
	"net/http"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/memphisdev/memphis/analytics"
	"github.com/memphisdev/memphis/db"
	"github.com/memphisdev/memphis/memphis_cache"
	"github.com/memphisdev/memphis/models"
	"github.com/memphisdev/memphis/utils"

	dockerClient "github.com/docker/docker/client"
	"github.com/gin-contrib/cors"

	"github.com/docker/docker/api/types"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const shouldCreateRootUserforGlobalAcc = true
const TENANT_SEQUENCE_START_ID = 2
const MAX_PARTITIONS = 10000

type BillingHandler struct{ S *Server }
type TenantHandler struct{ S *Server }
type LoginSchema struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type FunctionMetricsSchema struct {
	AverageProcessingTime float64 `json:"average_processing_time"`
	ErrorRate             float64 `json:"error_rate"`
	TotalInvocations      float64 `json:"total_invocations"`
}

type FunctionOverviewSchema struct {
	ID                int                   `json:"id"`
	Name              string                `json:"name"`
	StationId         int                   `json:"station_id"`
	Version           int                   `json:"version"`
	NextActiveStepId  int                   `json:"next_active_step_id"`
	PrevActiveStepId  int                   `json:"prev_active_step_id"`
	VisibleStep       int                   `json:"visible_step"`
	PartitionNumber   int                   `json:"partition_number"`
	Repo              string                `json:"repo"`
	Branch            string                `json:"branch"`
	Owner             string                `json:"owner"`
	Runtime           string                `json:"runtime"`
	OrderingMatter    bool                  `json:"ordering_matter"`
	ComputeEngine     string                `json:"compute_engine"`
	Activated         bool                  `json:"activated"`
	AddedBy           string                `json:"added_by"`
	SCM               string                `json:"scm"`
	InstalledId       int                   `json:"installed_id"`
	Metrics           FunctionMetricsSchema `json:"metrics"`
	PendingMessages   int                   `json:"pending_messages"`
	InProcessMessages int                   `json:"in_process_messages"`
	TenantName        string                `json:"tenant_name"`
}

type FunctionsOverviewResponse struct {
	Functions              []FunctionOverviewSchema `json:"functions"`
	TotalAwaitingMessages  int                      `json:"total_awaiting_messages"`
	TotalProcessedMessages int                      `json:"total_processed_messages"`
	TotalInvocations       int                      `json:"total_invocations"`
	AverageErrorRate       float64                  `json:"average_error_rate"`
	FunctionsExists        bool                     `json:"functions_exists"`
}

var ErrUpgradePlan = errors.New("to continue using Memphis, please upgrade your plan to a paid plan")

type MainOverviewData struct {
	TotalStations     int                               `json:"total_stations"`
	TotalMessages     uint64                            `json:"total_messages"`
	TotalDlsMessages  uint64                            `json:"total_dls_messages"`
	SystemComponents  []models.SystemComponents         `json:"system_components"`
	Stations          []models.ExtendedStationLight     `json:"stations"`
	K8sEnv            bool                              `json:"k8s_env"`
	BrokersThroughput []models.BrokerThroughputResponse `json:"brokers_throughput"`
	MetricsEnabled    bool                              `json:"metrics_enabled"`
	DelayedCgs        []models.DelayedCgResp            `json:"delayed_cgs"`
}

type SystemMessage struct {
	Id             string    `json:"id"`
	MessageType    string    `firestore:"message_type" json:"message_type"`
	MessagePayload string    `firestore:"message_payload" json:"message_payload"`
	StartTime      time.Time `firestore:"start_time" json:"start_time"`
	EndTime        time.Time `firestore:"end_time" json:"end_time"`
	UiPage         string    `firestore:"ui_page" json:"ui_page"`
}

type ProduceSchema struct {
	StationName     string            `json:"station_name" binding:"required"`
	PartitionNumber int               `json:"partition_number"`
	MsgPayload      string            `json:"message_payload" binding:"required"`
	MsgHdrs         map[string]string `json:"message_headers"`
	Amount          int               `json:"amount" binding:"required"`
	BypassSchema    bool              `json:"bypass_schema"`
	DataFormat      string            `json:"data_format"`
}

func InitializeBillingRoutes(router *gin.RouterGroup, h *Handlers) {
}

func InitializeTenantsRoutes(router *gin.RouterGroup, h *Handlers) {
}

func AddUsrMgmtCloudRoutes(userMgmtRoutes *gin.RouterGroup, userMgmtHandler UserMgmtHandler) {
}

func AddMonitoringCloudRoutes(monitoringRoutes *gin.RouterGroup, monitoringHandler MonitoringHandler) {
}

func getStationStorageType(storageType string) string {
	return strings.ToLower(storageType)
}

func GetStationMaxAge(retentionType, tenantName string, retentionValue int) time.Duration {
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

	created, err := db.UpsertUserUpdatePassword(ROOT_USERNAME, "root", hashedPwdString, "", false, 1, serv.MemphisGlobalAccountString())
	if err != nil {
		return err
	}

	shouldSendAnalytics, _ := shouldSendAnalytics()
	if created && shouldSendAnalytics {
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

			ip := serv.getIp()
			analyticsParams := map[string]interface{}{"installation-type": installationType, "device-id": deviceIdValue, "source": configuration.INSTALLATION_SOURCE, "ip": ip}
			analytics.SendEvent("", "", analyticsParams, "installation")

			if configuration.EXPORTER {
				analytics.SendEvent("", "", analyticsParams, "enable-exporter")
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
				storage_size = getUnixStorageSize()
				perc := 0
				if storage_size > 0 {
					perc = int(math.Ceil((float64(v.JetStream.Stats.Store) / storage_size) * 100))
				}
				storageComp = models.CompStats{
					Total:      shortenFloat(storage_size),
					Current:    shortenFloat(float64(v.JetStream.Stats.Store)),
					Percentage: perc,
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
			storage_size := getUnixStorageSize()
			if strings.Contains(containerName, "metadata") {
				dbStorageUsage, err := getDbStorageUsage()
				if err != nil {
					return components, metricsEnabled, err
				}
				perc := 0
				if storage_size > 0 {
					perc = int(math.Ceil((float64(dbStorageUsage) / float64(storage_size)) * 100))
				}
				storageStat = models.CompStats{
					Total:      shortenFloat(storage_size),
					Current:    shortenFloat(dbStorageUsage),
					Percentage: perc,
				}
				containerName = strings.TrimPrefix(containerName, "memphis-")
			} else if strings.Contains(containerName, "memphis-1") {
				v, err := serv.Varz(nil)
				if err != nil {
					return components, metricsEnabled, err
				}

				perc := 0
				if storage_size > 0 {
					perc = int(math.Ceil((float64(v.JetStream.Stats.Store) / storage_size) * 100))
				}
				storageStat = models.CompStats{
					Total:      shortenFloat(storage_size),
					Current:    shortenFloat(float64(v.JetStream.Stats.Store)),
					Percentage: perc,
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
			storage_size = getUnixStorageSize()
			perc := 0
			if storage_size > 0 {
				perc = int(math.Ceil((float64(v.JetStream.Stats.Store) / storage_size) * 100))
			}
			storageComp = models.CompStats{
				Total:      shortenFloat(storage_size),
				Current:    shortenFloat(float64(v.JetStream.Stats.Store)),
				Percentage: perc,
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
			storage_size := getUnixStorageSize()
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
				dbStorageUsage, err := getDbStorageUsage()
				if err != nil {
					return components, metricsEnabled, err
				}
				perc := 0
				if storage_size > 0 {
					perc = int(math.Ceil((float64(dbStorageUsage) / float64(storage_size)) * 100))
				}
				storageStat = models.CompStats{
					Total:      shortenFloat(storage_size),
					Current:    shortenFloat(dbStorageUsage),
					Percentage: perc,
				}

			} else if strings.Contains(containerName, "cluster") {
				v, err := serv.Varz(nil)
				if err != nil {
					return components, metricsEnabled, err
				}
				perc := 0
				if storage_size > 0 {
					perc = int(math.Ceil((float64(v.JetStream.Stats.Store) / storage_size) * 100))
				}
				storageStat = models.CompStats{
					Total:      shortenFloat(storage_size),
					Current:    shortenFloat(float64(v.JetStream.Stats.Store)),
					Percentage: perc,
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
						serv.Warnf("GetSystemComponents: k8s metrics not installed: %v", err.Error())
						noMetricsInstalledLog = true
					}
					continue
				} else if strings.Contains(err.Error(), "is forbidden") {
					metricsEnabled = false
					allComponents = append(allComponents, defaultSystemComp(pod.Name, true))
					if !noMetricsPermissionLog {
						serv.Warnf("GetSystemComponents: No permissions for k8s metrics: %v", err.Error())
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
					storageUsage, err = getDbStorageUsage()
					if err != nil {
						return components, metricsEnabled, err
					}
				} else if strings.Contains(strings.ToLower(pod.Name), "memphis-0") {
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
		if request.LogSource == "empty" || request.LogSource == "" {
			filterSubject = fmt.Sprintf("%s.%s.%s", syslogsStreamName, "*", filterSubjectSuffix)
		} else if request.LogSource != "empty" && request.LogType != "external" {
			filterSubject = fmt.Sprintf("%s.%s.%s", syslogsStreamName, logSource, filterSubjectSuffix)
		} else if request.LogSource != "empty" && request.LogType == "external" {
			filterSubject = fmt.Sprintf("%s.%s.%s.%s", syslogsStreamName, logSource, "extern", ">")
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
		analyticsParams := make(map[string]interface{})
		analytics.SendEvent(user.TenantName, user.Username, analyticsParams, "user-enter-syslogs-page")
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

func (s *Server) InitializeCloudComponents() error {
	return nil
}

func (s *Server) UploadTenantUsageToDB() error {
	return nil
}

func IncrementEventCounter(tenantName string, eventType string, size int64, amount int64, subj string, msg []byte, hdr []byte) {
}

func (ch ConfigurationsHandler) EditClusterConfig(c *gin.Context) {
	user, err := getUserDetailsFromMiddleware(c)
	if err != nil {
		serv.Errorf("EditClusterConfig at getUserDetailsFromMiddleware: %v", err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}

	var body models.EditClusterConfigSchema
	ok := utils.Validate(c, &body, false, nil)
	if !ok {
		return
	}
	if ch.S.opts.DlsRetentionHours[user.TenantName] != body.DlsRetention {
		err := changeDlsRetention(body.DlsRetention, user.TenantName)
		if err != nil {
			serv.Errorf("[tenant: %v][user: %v]EditConfigurations at changeDlsRetention: %v", user.TenantName, user.Username, err.Error())
			c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
			return
		}
	}
	if ch.S.opts.GCProducersConsumersRetentionHours != body.GCProducersConsumersRetentionHours {
		err := changeGCProducersConsumersRetentionHours(body.GCProducersConsumersRetentionHours, user.TenantName)
		if err != nil {
			serv.Errorf("[tenant: %v][user: %v]EditConfigurations at changeGCProducersConsumersRetentionHours: %v", user.TenantName, user.Username, err.Error())
			c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
			return
		}
	}
	if ch.S.opts.LogsRetentionDays != body.LogsRetention {
		err := changeLogsRetention(body.LogsRetention)
		if err != nil {
			serv.Errorf("[tenant: %v][user: %v]EditConfigurations at changeLogsRetention: %v", user.TenantName, user.Username, err.Error())
			c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
			return
		}
	}
	if ch.S.opts.TieredStorageUploadIntervalSec != body.TSTimeSec {
		if body.TSTimeSec > 3600 || body.TSTimeSec < 5 {
			serv.Errorf("[tenant: %v][user: %v]EditConfigurations: Tiered storage time can't be less than 5 seconds or more than 60 minutes", user.TenantName, user.Username)
			c.AbortWithStatusJSON(SHOWABLE_ERROR_STATUS_CODE, gin.H{"message": "Tiered storage time can't be less than 5 seconds or more than 60 minutes"})
		} else {
			err := changeTSTime(body.TSTimeSec)
			if err != nil {
				serv.Errorf("[tenant: %v][user: %v]EditConfigurations at changeTSTime: %v", user.TenantName, user.Username, err.Error())
				c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
				return
			}
		}
	}

	brokerHost := strings.ToLower(body.BrokerHost)
	if ch.S.opts.BrokerHost != brokerHost {
		err := EditClusterCompHost("broker_host", brokerHost)
		if err != nil {
			serv.Errorf("[tenant: %v][user: %v]EditConfigurations at EditClusterCompHost broker_host: %v", user.TenantName, user.Username, err.Error())
			c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
			return
		}
	}

	uiHost := strings.ToLower(body.UiHost)
	if ch.S.opts.UiHost != uiHost {
		err := EditClusterCompHost("ui_host", uiHost)
		if err != nil {
			serv.Errorf("[tenant: %v][user: %v]EditConfigurations at EditClusterCompHost ui_host: %v", user.TenantName, user.Username, err.Error())
			c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
			return
		}
	}

	restGWHost := strings.ToLower(body.RestGWHost)
	if ch.S.opts.RestGwHost != restGWHost {
		err := EditClusterCompHost("rest_gw_host", restGWHost)
		if err != nil {
			serv.Errorf("[tenant: %v][user: %v]EditConfigurations at EditClusterCompHost rest_gw_host: %v", user.TenantName, user.Username, err.Error())
			c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
			return
		}
	}

	if ch.S.opts.MaxPayload != int32(body.MaxMsgSizeMb) {
		err := changeMaxMsgSize(body.MaxMsgSizeMb)
		if err != nil {
			serv.Errorf("[tenant: %v][user: %v]EditConfigurations at changeMaxMsgSize: %v", user.TenantName, user.Username, err.Error())
			c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
			return
		}
	}

	// send signal to reload config
	err = serv.sendInternalAccountMsgWithReply(serv.MemphisGlobalAccount(), CONFIGURATIONS_RELOAD_SIGNAL_SUBJ, _EMPTY_, nil, _EMPTY_, true)
	if err != nil {
		serv.Errorf("[tenant: %v][user: %v]EditConfigurations at sendInternalAccountMsgWithReply: %v", user.TenantName, user.Username, err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}

	shouldSendAnalytics, _ := shouldSendAnalytics()
	if shouldSendAnalytics {
		user, _ := getUserDetailsFromMiddleware(c)
		analyticsParams := make(map[string]interface{})
		analytics.SendEvent(user.TenantName, user.Username, analyticsParams, "user-update-cluster-config")
	}

	c.IndentedJSON(200, gin.H{
		"dls_retention":                        body.DlsRetention,
		"logs_retention":                       body.LogsRetention,
		"broker_host":                          brokerHost,
		"ui_host":                              uiHost,
		"rest_gw_host":                         restGWHost,
		"tiered_storage_time_sec":              body.TSTimeSec,
		"max_msg_size_mb":                      int32(body.MaxMsgSizeMb),
		"gc_producer_consumer_retention_hours": body.GCProducersConsumersRetentionHours,
	})
}

func (ch ConfigurationsHandler) GetClusterConfig(c *gin.Context) {
	user, err := getUserDetailsFromMiddleware(c)
	if err != nil {
		serv.Errorf("GetClusterConfig at getUserDetailsFromMiddleware: %v", err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}

	shouldSendAnalytics, _ := shouldSendAnalytics()
	if shouldSendAnalytics {
		analyticsParams := make(map[string]interface{})
		analytics.SendEvent(user.TenantName, user.Username, analyticsParams, "user-enter-cluster-config-page")
	}
	c.IndentedJSON(200, gin.H{
		"dls_retention":                        ch.S.opts.DlsRetentionHours[user.TenantName],
		"logs_retention":                       ch.S.opts.LogsRetentionDays,
		"broker_host":                          ch.S.opts.BrokerHost,
		"ui_host":                              ch.S.opts.UiHost,
		"rest_gw_host":                         ch.S.opts.RestGwHost,
		"tiered_storage_time_sec":              ch.S.opts.TieredStorageUploadIntervalSec,
		"max_msg_size_mb":                      ch.S.opts.MaxPayload / 1024 / 1024,
		"gc_producer_consumer_retention_hours": ch.S.opts.GCProducersConsumersRetentionHours,
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

func (th TenantHandler) CreateTenant(c *gin.Context) {
	// use the func changeDlsRetention(DEFAULT_DLS_RETENTION_HOURS, tenantName) when creating a new tenant
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
		serv.Errorf("[tenant: %v][user: %v]Login at authenticateUser: User %v: %v", user.TenantName, user.Username, body.Username, err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}
	if !authenticated || user.UserType == "application" {
		c.AbortWithStatusJSON(401, gin.H{"message": "Unauthorized"})
		return
	}

	token, refreshToken, err := CreateTokens(user)
	if err != nil {
		serv.Errorf("[tenant: %v][user: %v]Login at CreateTokens: User %v: %v", user.TenantName, user.Username, body.Username, err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}

	if !user.AlreadyLoggedIn {
		err = db.UpdateUserAlreadyLoggedIn(user.ID)
		if err != nil {
			serv.Errorf("[tenant: %v][user: %v]Login at UpdateUserAlreadyLoggedIn: User %v: %v", user.TenantName, user.Username, body.Username, err.Error())
			c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
			return
		}
	}

	env := "K8S"
	if configuration.DOCKER_ENV != "" || configuration.LOCAL_CLUSTER_ENV {
		env = "docker"
	}
	exist, tenant, err := db.GetTenantByName(user.TenantName)
	if err != nil {
		serv.Errorf("[tenant: %v][user: %v]Login at GetTenantByName: User %v: %v", user.TenantName, user.Username, body.Username, err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}
	if !exist {
		serv.Warnf("[tenant: %v][user: %v]Login: User %v: tenant %v does not exist", user.TenantName, user.Username, body.Username, user.TenantName)
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}
	decriptionKey := getAESKey()
	decryptedUserPassword, err := DecryptAES(decriptionKey, tenant.InternalWSPass)
	if err != nil {
		serv.Errorf("Login: User " + body.Username + ": " + err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}

	serv.Noticef("[tenant: %v][user: %v] has logged in", user.TenantName, user.Username)
	shouldSendAnalytics, _ := shouldSendAnalytics()
	if shouldSendAnalytics {
		analyticsParams := make(map[string]interface{})
		analytics.SendEvent(user.TenantName, user.Username, analyticsParams, "user-login")
	}

	domain := ""
	secure := false
	c.SetCookie("memphis-jwt-refresh-token", refreshToken, REFRESH_JWT_EXPIRES_IN_MINUTES*60*1000, "/", domain, secure, true)
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
		"internal_ws_pass":        decryptedUserPassword,
		"dls_retention":           serv.opts.DlsRetentionHours[user.TenantName],
		"logs_retention":          serv.opts.LogsRetentionDays,
		"max_msg_size_mb":         serv.opts.MaxPayload / 1024 / 1024,
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
		serv.Errorf("AddUser: %v", err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}

	var subscription, pending bool
	team := strings.ToLower(body.Team)
	teamError := validateUserTeam(team)
	if teamError != nil {
		serv.Warnf("[tenant: %v][user: %v]AddUser at validateUserTeam: %v", user.TenantName, user.Username, teamError.Error())
		c.AbortWithStatusJSON(SHOWABLE_ERROR_STATUS_CODE, gin.H{"message": teamError.Error()})
		return
	}
	position := strings.ToLower(body.Position)
	positionError := validateUserPosition(position)
	if positionError != nil {
		serv.Warnf("[tenant: %v][user: %v]AddUser at validateUserPosition: %v", user.TenantName, user.Username, positionError.Error())
		c.AbortWithStatusJSON(SHOWABLE_ERROR_STATUS_CODE, gin.H{"message": positionError.Error()})
		return
	}
	fullName := strings.ToLower(body.FullName)
	fullNameError := validateUserFullName(fullName)
	if fullNameError != nil {
		serv.Warnf("[tenant: %v][user: %v]AddUser at validateUserFullName: %v", user.TenantName, user.Username, fullNameError.Error())
		c.AbortWithStatusJSON(SHOWABLE_ERROR_STATUS_CODE, gin.H{"message": fullNameError.Error()})
		return
	}
	owner := user.Username
	description := strings.ToLower(body.Description)
	descriptionError := validateUserDescription(description)
	if descriptionError != nil {
		serv.Warnf("[tenant: %v][user: %v]AddUser at validateUserDescription: %v", user.TenantName, user.Username, descriptionError.Error())
		c.AbortWithStatusJSON(SHOWABLE_ERROR_STATUS_CODE, gin.H{"message": descriptionError.Error()})
		return
	}

	if user.TenantName != DEFAULT_GLOBAL_ACCOUNT {
		user.TenantName = strings.ToLower(user.TenantName)
	}
	username := strings.ToLower(body.Username)
	usernameError := validateUsername(username)
	if usernameError != nil {
		serv.Warnf("[tenant: %v][user: %v]AddUser at validateUsername: %v", user.TenantName, user.Username, usernameError.Error())
		c.AbortWithStatusJSON(SHOWABLE_ERROR_STATUS_CODE, gin.H{"message": usernameError.Error()})
		return
	}
	exist, _, err := memphis_cache.GetUser(username, user.TenantName, true)
	if err != nil {
		serv.Errorf("[tenant: %v][user: %v]AddUser at GetUserByUsername: User %v: %v", user.TenantName, user.Username, body.Username, err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}
	if exist {
		errMsg := fmt.Sprintf("A user with the name %v already exists", body.Username)
		serv.Warnf("[tenant: %v][user: %v]CreateUser: %v", user.TenantName, user.Username, errMsg)
		c.AbortWithStatusJSON(SHOWABLE_ERROR_STATUS_CODE, gin.H{"message": errMsg})
		return
	}

	userType := strings.ToLower(body.UserType)
	userTypeError := validateUserType(userType)
	if userTypeError != nil {
		serv.Warnf("[tenant: %v][user: %v]AddUser at validateUserType: %v", user.TenantName, user.Username, userTypeError.Error())
		c.AbortWithStatusJSON(SHOWABLE_ERROR_STATUS_CODE, gin.H{"message": userTypeError.Error()})
		return
	}

	avatarId := 1
	if body.AvatarId > 0 {
		avatarId = body.AvatarId
	}

	if body.Password == "" {
		serv.Warnf("[tenant: %v][user: %v]AddUser: Password was not provided for user %v", user.TenantName, user.Username, username)
		c.AbortWithStatusJSON(SHOWABLE_ERROR_STATUS_CODE, gin.H{"message": "Password was not provided"})
		return
	}
	passwordErr := validatePassword(body.Password)
	if passwordErr != nil {
		serv.Warnf("[tenant: %v][user: %v]AddUser validate password : User %v: %v", user.TenantName, user.Username, body.Username, passwordErr.Error())
		c.AbortWithStatusJSON(SHOWABLE_ERROR_STATUS_CODE, gin.H{"message": passwordErr.Error()})
		return
	}

	var password string
	if userType == "management" {

		hashedPwd, err := bcrypt.GenerateFromPassword([]byte(body.Password), bcrypt.MinCost)
		if err != nil {
			serv.Errorf("[tenant: %v][user: %v]AddUser at GenerateFromPassword: User %v: %v", user.TenantName, user.Username, body.Username, err.Error())
			c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
			return
		}
		password = string(hashedPwd)
	}

	var brokerConnectionCreds string
	if userType == "application" {
		fullName = ""
		subscription = false
		pending = false
		if configuration.USER_PASS_BASED_AUTH {
			if body.Password == "" {
				serv.Warnf("[tenant: %v][user: %v]AddUser: Password was not provided for user %v", user.TenantName, user.Username, username)
				c.AbortWithStatusJSON(SHOWABLE_ERROR_STATUS_CODE, gin.H{"message": "Password was not provided"})
				return
			}
			password, err = EncryptAES([]byte(body.Password))
			if err != nil {
				serv.Errorf("[tenant: %v][user: %v]AddUser at EncryptAES: User %v: %v", user.TenantName, user.Username, body.Username, err.Error())
				c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
				return
			}
		} else {
			brokerConnectionCreds = configuration.CONNECTION_TOKEN
		}
	}
	newUser, err := db.CreateUser(username, userType, password, fullName, subscription, avatarId, user.TenantName, pending, team, position, owner, description)
	if err != nil {
		if strings.Contains(err.Error(), "already exist") {
			serv.Warnf("[tenant: %v][user: %v]CreateUserManagement user already exists: %v", user.TenantName, user.Username, err.Error())
			c.AbortWithStatusJSON(SHOWABLE_ERROR_STATUS_CODE, gin.H{"message": err.Error()})
			return
		}
		serv.Errorf("[tenant: %v][user: %v]AddUser at CreateUser: User %v: %v", user.TenantName, user.Username, body.Username, err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}

	err = memphis_cache.SetUser(newUser)
	if err != nil {
		serv.Errorf("[tenant: %v][user: %v]AddUser at writing to the user cache error: %v", user.TenantName, user.Username, err)
	}

	shouldSendAnalytics, _ := shouldSendAnalytics()
	if shouldSendAnalytics {
		analyticsParams := map[string]interface{}{
			"username": username,
		}
		analytics.SendEvent(user.TenantName, user.Username, analyticsParams, "user-add-user")
	}

	if userType == "application" && configuration.USER_PASS_BASED_AUTH {
		// send signal to reload config
		err = serv.sendInternalAccountMsgWithReply(serv.MemphisGlobalAccount(), CONFIGURATIONS_RELOAD_SIGNAL_SUBJ, _EMPTY_, nil, _EMPTY_, true)
		if err != nil {
			serv.Errorf("[tenant: %v][user: %v]AddUser at sendInternalAccountMsgWithReply: User %v: %v", user.TenantName, user.Username, body.Username, err.Error())
			c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
			return
		}
	}

	serv.Noticef("[tenant: %v][user: %v]User %v has been created", user.TenantName, user.Username, username)
	c.IndentedJSON(200, gin.H{
		"id":                      newUser.ID,
		"username":                username,
		"full_name":               fullName,
		"user_type":               userType,
		"created_at":              newUser.CreatedAt,
		"already_logged_in":       false,
		"avatar_id":               avatarId,
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
		serv.Errorf("RemoveUser: User %v: %v", body.Username, err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}
	if user.Username == username {
		serv.Warnf("[tenant: %v][user: %v]RemoveUser: You can not remove your own user", user.TenantName, user.Username)
		c.AbortWithStatusJSON(SHOWABLE_ERROR_STATUS_CODE, gin.H{"message": "You can not remove your own user"})
		return
	}

	exist, userToRemove, err := memphis_cache.GetUser(username, user.TenantName, false)
	if err != nil {
		serv.Errorf("[tenant: %v][user: %v]RemoveUser at GetUserByUsername: User %v: %v", user.TenantName, user.Username, body.Username, err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}
	if !exist {
		serv.Warnf("[tenant: %v][user: %v]RemoveUser: User does not exist", user.TenantName, user.Username)
		c.AbortWithStatusJSON(SHOWABLE_ERROR_STATUS_CODE, gin.H{"message": "User does not exist"})
		return
	}
	if userToRemove.UserType == "root" {
		serv.Warnf("[tenant: %v][user: %v]RemoveUser: You can not remove the root user", user.TenantName, user.Username)
		c.AbortWithStatusJSON(SHOWABLE_ERROR_STATUS_CODE, gin.H{"message": "You can not remove the root user"})
		return
	}

	err = updateDeletedUserResources(userToRemove)
	if err != nil {
		serv.Errorf("[tenant: %v][user: %v]RemoveUser at updateDeletedUserResources: User %v: %v", user.TenantName, user.Username, body.Username, err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": err.Error()})
		return
	}

	err = db.DeleteUser(username, userToRemove.TenantName)
	if err != nil {
		serv.Errorf("[tenant: %v][user: %v]RemoveUser at DeleteUser: User %v: %v", user.TenantName, user.Username, body.Username, err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}

	SendUserDeleteCacheUpdate([]string{username}, user.TenantName)

	if userToRemove.UserType == "application" && configuration.USER_PASS_BASED_AUTH {
		// send signal to reload config
		err = serv.sendInternalAccountMsgWithReply(serv.MemphisGlobalAccount(), CONFIGURATIONS_RELOAD_SIGNAL_SUBJ, _EMPTY_, nil, _EMPTY_, true)
		if err != nil {
			serv.Errorf("[tenant: %v][user: %v]RemoveUser at sendInternalAccountMsgWithReply: User %v: %v", user.TenantName, user.Username, body.Username, err.Error())
			c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
			return
		}
	}

	shouldSendAnalytics, _ := shouldSendAnalytics()
	if shouldSendAnalytics {
		analyticsParams := map[string]interface{}{
			"username": username,
		}
		analytics.SendEvent(user.TenantName, user.Username, analyticsParams, "user-remove-user")
	}

	serv.Noticef("[tenant: %v][user: %v]User %v has been deleted by user %v", user.TenantName, user.Username, username, user.Username)
	c.IndentedJSON(200, gin.H{})
}

func (umh UserMgmtHandler) RemoveMyUser(c *gin.Context) {
	user, err := getUserDetailsFromMiddleware(c)
	if err != nil {
		serv.Errorf("RemoveMyUser at getUserDetailsFromMiddleware: %v", err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}

	if user.UserType != "root" {
		serv.Warnf("RemoveMyUser: Only root user can remove the entire account")
		c.AbortWithStatusJSON(SHOWABLE_ERROR_STATUS_CODE, gin.H{"message": "Only root user can remove the entire account"})
		return
	}

	username := strings.ToLower(user.Username)
	tenantName := user.TenantName
	if user.TenantName != DEFAULT_GLOBAL_ACCOUNT {
		user.TenantName = strings.ToLower(user.TenantName)
	}
	err = removeTenantResources(tenantName, user)
	if err != nil {
		serv.Errorf("[tenant: %v][user: %v]RemoveMyUser at removeTenantResources: User %v: %v", tenantName, username, username, err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": err.Error()})
		return
	}

	shouldSendAnalytics, _ := shouldSendAnalytics()
	if shouldSendAnalytics {
		analyticsParams := make(map[string]interface{})
		analytics.SendEvent(user.TenantName, user.Username, analyticsParams, "user-remove-himself")
	}

	serv.Noticef("[tenant: %v][user: %v]Tenant %v has been deleted", tenantName, username, user.TenantName)
	c.IndentedJSON(200, gin.H{})
}

func (s *Server) RefreshFirebaseFunctionsKey() {
}

func shouldPersistSysLogs() bool {
	return true
}

func (umh UserMgmtHandler) EditAnalytics(c *gin.Context) {
	var body models.EditAnalyticsSchema
	ok := utils.Validate(c, &body, false, nil)
	if !ok {
		return
	}

	flag := "false"
	if body.SendAnalytics {
		flag = "true"
	}

	err := db.EditConfigurationValue("analytics", flag, serv.MemphisGlobalAccountString())
	if err != nil {
		serv.Errorf("EditAnalytics: " + err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}

	if !body.SendAnalytics {
		user, _ := getUserDetailsFromMiddleware(c)
		analyticsParams := make(map[string]interface{})
		analytics.SendEvent(user.TenantName, user.Username, analyticsParams, "user-disable-analytics")
	}

	c.IndentedJSON(200, gin.H{})
}

func (s *Server) GetCustomDeploymentId() string {
	return ""
}

func (s *Server) sendLogToAnalytics(label string, log []byte) {
	switch label {
	case "ERR":
		shouldSend, err := shouldSendAnalytics()
		if err != nil || !shouldSend {
			return
		}
		analyticsParams := map[string]interface{}{"err_source": s.getLogSource(), "err_log": string(log)}
		analytics.SendEvent("", "", analyticsParams, "error")
	default:
		return
	}
}

func (mh MonitoringHandler) getMainOverviewDataDetails(tenantName string) (MainOverviewData, error) {
	var wg sync.WaitGroup
	var mu sync.Mutex
	mainOverviewData := &MainOverviewData{}
	generalErr := new(error)

	streams, err := serv.memphisAllStreamsInfo(tenantName)
	if err != nil {
		return MainOverviewData{}, err
	}

	wg.Add(4)
	go func(streamsInfo []*StreamInfo) {
		stationsHandler := StationsHandler{S: mh.S}
		stations, totalMessages, totalDlsMsgs, err := stationsHandler.GetAllStationsDetailsLight(false, tenantName, streamsInfo)
		if err != nil {
			*generalErr = err
			wg.Done()
			return
		}
		mu.Lock()
		mainOverviewData.TotalStations = len(stations)
		mainOverviewData.Stations = stations
		mainOverviewData.TotalMessages = totalMessages
		mainOverviewData.TotalDlsMessages = totalDlsMsgs
		mu.Unlock()
		wg.Done()
	}(streams)

	go func() {
		systemComponents, metricsEnabled, err := mh.GetSystemComponents()
		if err != nil {
			*generalErr = err
			wg.Done()
			return
		}
		mu.Lock()
		mainOverviewData.SystemComponents = systemComponents
		mainOverviewData.MetricsEnabled = metricsEnabled
		mu.Unlock()
		wg.Done()
	}()

	go func() {
		brokersThroughputs, err := mh.GetBrokersThroughputs(tenantName)
		if err != nil {
			*generalErr = err
			wg.Done()
			return
		}
		mu.Lock()
		mainOverviewData.BrokersThroughput = brokersThroughputs
		mu.Unlock()
		wg.Done()
	}()

	go func(streamsInfo []*StreamInfo) {
		consumersHandler := ConsumersHandler{S: mh.S}
		delayedConsumersMap, err := consumersHandler.GetDelayedCgsByTenant(tenantName, streamsInfo)
		if err != nil {
			*generalErr = err
			wg.Done()
			return
		}
		mu.Lock()
		mainOverviewData.DelayedCgs = delayedConsumersMap
		mu.Unlock()
		wg.Done()
	}(streams)
	wg.Wait()
	if *generalErr != nil {
		return MainOverviewData{}, *generalErr
	}

	k8sEnv := true
	if configuration.DOCKER_ENV == "true" || configuration.LOCAL_CLUSTER_ENV {
		k8sEnv = false
	}
	mainOverviewData.K8sEnv = k8sEnv
	return *mainOverviewData, nil
}

func (umh UserMgmtHandler) RefreshToken(c *gin.Context) {
	user, err := getUserDetailsFromMiddleware(c)
	if err != nil {
		serv.Errorf("RefreshToken: " + err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}
	username := user.Username
	_, systemKey, err := db.GetSystemKey("analytics", serv.MemphisGlobalAccountString())
	if err != nil {
		serv.Errorf("RefreshToken: User " + username + ": " + err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}
	sendAnalytics, _ := strconv.ParseBool(systemKey.Value)
	exist, user, err := memphis_cache.GetUser(username, user.TenantName, true)
	if err != nil {
		serv.Errorf("RefreshToken: User " + username + ": " + err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}
	if !exist {
		serv.Warnf("RefreshToken: user " + username + " does not exist")
		c.AbortWithStatusJSON(401, gin.H{"message": "Unauthorized"})
		return
	}

	token, refreshToken, err := CreateTokens(user)
	if err != nil {
		serv.Errorf("RefreshToken: User " + username + ": " + err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}

	env := "K8S"
	if configuration.DOCKER_ENV != "" || configuration.LOCAL_CLUSTER_ENV {
		env = "docker"
	}

	exist, tenant, err := db.GetTenantByName(user.TenantName)
	if err != nil {
		serv.Errorf("RefreshToken: User " + username + ": " + err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}
	if !exist {
		serv.Warnf("RefreshToken: User " + username + ": tenant " + user.TenantName + " does not exist")
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}
	decriptionKey := getAESKey()
	decryptedUserPassword, err := DecryptAES(decriptionKey, tenant.InternalWSPass)
	if err != nil {
		serv.Errorf("RefreshToken: User " + username + ": " + err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}

	domain := ""
	secure := true
	c.SetCookie("memphis-jwt-refresh-token", refreshToken, REFRESH_JWT_EXPIRES_IN_MINUTES*60*1000, "/", domain, secure, true)
	c.IndentedJSON(200, gin.H{
		"jwt":                     token,
		"expires_in":              JWT_EXPIRES_IN_MINUTES * 60 * 1000,
		"user_id":                 user.ID,
		"username":                user.Username,
		"user_type":               user.UserType,
		"created_at":              user.CreatedAt,
		"already_logged_in":       user.AlreadyLoggedIn,
		"avatar_id":               user.AvatarId,
		"send_analytics":          sendAnalytics,
		"env":                     env,
		"namespace":               serv.opts.K8sNamespace,
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
		"internal_ws_pass":        decryptedUserPassword,
		"dls_retention":           serv.opts.DlsRetentionHours[user.TenantName],
		"logs_retention":          serv.opts.LogsRetentionDays,
		"max_msg_size_mb":         serv.opts.MaxPayload / 1024 / 1024,
	})
}

func (mh MonitoringHandler) GetBrokersThroughputs(tenantName string) ([]models.BrokerThroughputResponse, error) {
	uid := serv.memphis.nuid.Next()
	durableName := "$memphis_fetch_throughput_consumer_" + uid
	var msgs []StoredMsg
	var throughputs []models.BrokerThroughputResponse
	streamInfo, err := serv.memphisStreamInfo(serv.MemphisGlobalAccountString(), throughputStreamNameV1)
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
		Replicas:      1,
	}
	err = serv.memphisAddConsumer(serv.MemphisGlobalAccountString(), throughputStreamNameV1, &cc)
	if err != nil {
		return throughputs, err
	}

	responseChan := make(chan StoredMsg)
	subject := fmt.Sprintf(JSApiRequestNextT, throughputStreamNameV1, durableName)
	reply := durableName + "_reply"
	req := []byte(strconv.FormatUint(amount, 10))

	sub, err := serv.subscribeOnAcc(serv.MemphisGlobalAccount(), reply, reply+"_sid", func(_ *client, subject, reply string, msg []byte) {
		go func(respCh chan StoredMsg, subject, reply string, msg []byte) {
			// ack
			serv.sendInternalAccountMsg(serv.MemphisGlobalAccount(), reply, []byte(_EMPTY_))
			rawTs := tokenAt(reply, 8)
			seq, _, _ := ackReplyInfo(reply)

			intTs, err := strconv.Atoi(rawTs)
			if err != nil {
				serv.Errorf("[tenant: %v]GetBrokersThroughputs: %v", tenantName, err.Error())
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

	serv.sendInternalAccountMsgWithReply(serv.MemphisGlobalAccount(), subject, reply, nil, req, true)
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
	serv.unsubscribeOnAcc(serv.MemphisGlobalAccount(), sub)
	time.AfterFunc(500*time.Millisecond, func() {
		serv.memphisRemoveConsumer(serv.MemphisGlobalAccountString(), throughputStreamNameV1, durableName)
	})

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
			Read:      brokerThroughput.ReadMap[tenantName],
		})
		mapEntry.Write = append(m[brokerThroughput.Name].Write, models.ThroughputWriteResponse{
			Timestamp: msg.Time,
			Write:     brokerThroughput.WriteMap[tenantName],
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

func (s *Server) validateAccIdInUsername(username string) bool {
	return true
}

func (s *Server) SendBillingAlertWhenNeeded() error {
	return nil
}

func shouldSendAnalytics() (bool, error) {
	if configuration.ENV == "staging" || configuration.ENV == "dev" {
		return false, nil
	}
	return true, nil
	// exist, systemKey, err := db.GetSystemKey("analytics", serv.MemphisGlobalAccountString())
	// if err != nil {
	// 	return false, err
	// }
	// if !exist {
	// 	return false, nil
	// }

	// if systemKey.Value == "true" {
	// 	return true, nil
	// } else {
	// 	return false, nil
	// }
}

func validateAmountOfMessagesToProduce(amount int) error {
	if amount <= 0 || amount > 1 {
		return errors.New("amount of messages to produce has to be positive and not larger than 1")
	}

	return nil
}

func validatePayloadLength(payload string) error {
	if len(payload) > 100 {
		return errors.New("max message payload length is 100 characters")
	}

	return nil
}

func validatePartitionToProduce(partitionNumber int) error {
	if partitionNumber > 0 {
		return errors.New("you can't produce to a specific partition")
	}

	return nil
}

func TenantSeqInitialize() error {
	err := db.SetTenantSequence(TENANT_SEQUENCE_START_ID)
	if err != nil {
		return err
	}
	return nil
}

func GetAvailableReplicas(replicas int) int {
	return replicas
}

func validateReplicas(replicas int) error {
	if replicas > 5 {
		return errors.New("max replicas in a cluster is 5")
	}

	return nil
}

func (s *Server) Force3ReplicationsForExistingStations() error {
	return nil
}

func getStationReplicas(replicas int) int {
	if replicas <= 0 {
		return 1
	} else if replicas == 2 || replicas == 4 {
		return 3
	} else if replicas > 5 {
		return 5
	}
	return replicas
}

func getDefaultReplicas() int {
	return 1
}

func updateSystemLiveness() {
	shouldSend, _ := shouldSendAnalytics()
	if shouldSend {
		stationsHandler := StationsHandler{S: serv}
		stations, totalMessages, totalDlsMsgs, err := stationsHandler.GetAllStationsDetailsLight(false, "", nil)
		if err != nil {
			serv.Warnf("updateSystemLiveness: %v", err.Error())
			return
		}

		producersCount, err := db.CountAllActiveProudcers()
		if err != nil {
			serv.Warnf("updateSystemLiveness: %v", err.Error())
			return
		}

		consumersCount, err := db.CountAllActiveConsumers()
		if err != nil {
			serv.Warnf("updateSystemLiveness: %v", err.Error())
			return
		}

		analyticsParams := map[string]interface{}{"total-messages": strconv.Itoa(int(totalMessages)), "total-dls-messages": strconv.Itoa(int(totalDlsMsgs)), "total-stations": strconv.Itoa(len(stations)), "active-producers": strconv.Itoa(int(producersCount)), "active-consumers": strconv.Itoa(int(consumersCount))}
		analytics.SendEvent("", "", analyticsParams, "system-is-up")
	}
}

func (umh UserMgmtHandler) GetRelevantSystemMessages() ([]SystemMessage, error) {
	return []SystemMessage{}, nil
}

func (s *Server) SetDlsRetentionForExistTenants() error {
	return nil
}

func (sh StationsHandler) Produce(c *gin.Context) {
	var body ProduceSchema
	ok := utils.Validate(c, &body, false, nil)
	if !ok {
		return
	}

	user, err := getUserDetailsFromMiddleware(c)
	if err != nil {
		serv.Errorf("Produce: could not get user from middleware: %v", err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}

	err = validateAmountOfMessagesToProduce(body.Amount)
	if err != nil {
		serv.Warnf("[tenant: %v][user: %v]Produce at validateAmountOfMessagesToProduce: Station %v: %v", user.TenantName, user.Username, body.StationName, err.Error())
		c.AbortWithStatusJSON(SHOWABLE_ERROR_STATUS_CODE, gin.H{"message": err.Error()})
		return
	}

	err = validatePayloadLength(body.MsgPayload)
	if err != nil {
		serv.Warnf("[tenant: %v][user: %v]Produce at validatePayloadLength: Station %v: %v", user.TenantName, user.Username, body.StationName, err.Error())
		c.AbortWithStatusJSON(SHOWABLE_ERROR_STATUS_CODE, gin.H{"message": err.Error()})
		return
	}

	err = validatePartitionToProduce(body.PartitionNumber)
	if err != nil {
		serv.Warnf("[tenant: %v][user: %v]Produce at validatePartitionToProduce: Station %v: %v", user.TenantName, user.Username, body.StationName, err.Error())
		c.AbortWithStatusJSON(SHOWABLE_ERROR_STATUS_CODE, gin.H{"message": err.Error()})
		return
	}

	stationName, err := StationNameFromStr(body.StationName)
	if err != nil {
		serv.Warnf("[tenant: %v][user: %v]Produce at StationNameFromStr: At station %v: %v", user.TenantName, user.Username, body.StationName, err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}

	exist, station, err := db.GetStationByName(stationName.Ext(), user.TenantName)
	if err != nil {
		serv.Errorf("[tenant: %v][user: %v]Produce at GetStationByName: At station %v: %v", user.TenantName, user.Username, body.StationName, err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}
	if !exist {
		errMsg := fmt.Sprintf("Station %v does not exist", body.StationName)
		serv.Warnf("[tenant: %v][user: %v]Produce: %v", user.TenantName, user.Username, errMsg)
		c.AbortWithStatusJSON(SHOWABLE_ERROR_STATUS_CODE, gin.H{"message": errMsg})
		return
	}

	subject := ""
	shouldRoundRobin := false
	if station.Version == 0 {
		subject = fmt.Sprintf("%s.final", stationName.Intern())
	} else {
		shouldRoundRobin = true
	}

	account, err := serv.lookupAccount(user.TenantName)
	if err != nil {
		serv.Errorf("[tenant: %v][user: %v]Produce at lookupAccount: %v", user.TenantName, user.Username, err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}

	if body.MsgHdrs == nil {
		body.MsgHdrs = make(map[string]string)
	}
	body.MsgHdrs["$memphis_producedBy"] = "UI"
	body.MsgHdrs["$memphis_connectionId"] = "UI"
	if shouldRoundRobin {
		rand.Seed(time.Now().UnixNano())
		randomIndex := rand.Intn(len(station.PartitionsList))
		subject = fmt.Sprintf("%s$%v.final", stationName.Intern(), station.PartitionsList[randomIndex])
	}
	serv.sendInternalAccountMsgWithHeadersWithEcho(account, subject, body.MsgPayload, body.MsgHdrs)

	c.IndentedJSON(200, gin.H{})
}

type GraphOverviewResponse struct {
	Stations map[int]models.StationLight `json:"stations"`
}

func (mh MonitoringHandler) getGraphOverview(tenantName string) (GraphOverviewResponse, error) {
	return GraphOverviewResponse{}, nil
}

func (s *Server) CreateDefaultEntitiesOnMemphisAccount() error {
	defaultStationName := "default"
	exist, user, err := db.GetRootUser(serv.MemphisGlobalAccountString())
	if err != nil {
		return err
	}
	if !exist {
		return errors.New("root user does not exist")
	}

	stationName, err := StationNameFromStr(defaultStationName)
	if err != nil {
		return err
	}

	schemaName, err := CreateDefaultSchema(user.Username, user.TenantName, user.ID)
	if err != nil {
		return err
	}

	_, _, err = CreateDefaultStation(serv.MemphisGlobalAccountString(), serv, stationName, user.ID, user.Username, schemaName, 1)
	if err != nil {
		return err
	}

	return nil
}

func ScheduledCloudCacheRefresh() {
}

func ValidataAccessToFeature(tenantName, featureName string) bool {
	return true
}

func ValidataUsageLimitOfFeature(tenantName, featureName string, amount int) (bool, int) {
	return true, 10000 // this is the number of the max partitions
}

func validateRetentionPolicyUsage(tenantName, retentionType string, retentionValue int) bool {
	return true
}

func InitializeCloudComponents() error {
	return nil
}

func (s *Server) ListenForCloudCacheUpdates() error {
	return nil
}

func (c *client) AccountConnExceeded() {
	c.sendErrAndErr(ErrTooManyAccountConnections.Error())
}

func IsStorageLimitExceeded(tenantName string) bool {
	return false
}

func validateProducersCount(stationId int, tenantName string) error {
	return nil
}

func InitializeCloudFunctionRoutes(functionsHandler FunctionsHandler, functionsRoutes *gin.RouterGroup) {
}

// Integrations

func (it IntegrationsHandler) GetSourecCodeBranches(c *gin.Context) {
	c.IndentedJSON(401, "Unautorized")
}

func InitializeCloudStationRoutes(stationsHandler StationsHandler, stationsRoutes *gin.RouterGroup) {}

func validatePartitionNumber(partitionsList []int, partition int) bool {
	for _, val := range partitionsList {
		if val == partition {
			return true
		}
	}
	return false
}

func GetStationAttachedFunctionsByPartitions(stationID int, partitionsList []int) ([]db.FunctionSchema, error) {
	return []db.FunctionSchema{}, nil
}

func getInternalUserPassword() string {
	return configuration.ROOT_PASSWORD
}

func sendDeleteAllFunctionsReqToMS(user models.User, tenantName, scmType, repo, branch, computeEngine, owner string, uninstall bool) error {
	return nil
}

func sendCloneFunctionReqToMS(connectedRepo interface{}, user models.User, scm string, bodyToUpdate models.CreateIntegrationSchema, index int) {
}

func GetAllFirstActiveFunctionsIDByStationID(stationId int, tenantName string) (map[int]int, error) {
	return map[int]int{}, nil
}
