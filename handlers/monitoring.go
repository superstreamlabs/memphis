// Copyright 2021-2022 The Memphis Authors
// Licensed under the GNU General Public License v3.0 (the “License”);
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// https://www.gnu.org/licenses/gpl-3.0.en.html
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an “AS IS” BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package handlers

import (
	"context"
	"flag"
	"io/ioutil"
	"memphis-broker/broker"
	"memphis-broker/db"
	"memphis-broker/logger"
	"memphis-broker/models"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

type MonitoringHandler struct{}

var clientset *kubernetes.Clientset

func clientSetConfig() error {
	var config *rest.Config
	var err error
	if configuration.DEV_ENV != "" { // dev environment is running locally and not inside the cluster
		// outside the cluster config
		var kubeconfig *string
		if home := homedir.HomeDir(); home != "" {
			kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "/Users/idanasulin/.kube/config")
		} else {
			kubeconfig = flag.String("kubeconfig", "", "/Users/idanasulin/.kube/config")
		}
		flag.Parse()
		config, err = clientcmd.BuildConfigFromFlags("", *kubeconfig)
		if err != nil {
			return err
		}
	} else {
		// in cluster config
		config, err = rest.InClusterConfig()
		if err != nil {
			return err
		}
	}

	clientset, err = kubernetes.NewForConfig(config)
	if err != nil {
		return err
	}

	return nil
}

func (mh MonitoringHandler) GetSystemComponents() ([]models.SystemComponent, error) {
	var components []models.SystemComponent
	if configuration.DOCKER_ENV != "" { // docker env
		_, err := http.Get("http://localhost:9000")
		if err != nil {
			components = append(components, models.SystemComponent{
				Component:   "ui",
				DesiredPods: 1,
				ActualPods:  0,
			})
		} else {
			components = append(components, models.SystemComponent{
				Component:   "ui",
				DesiredPods: 1,
				ActualPods:  1,
			})
		}

		if broker.IsConnectionAlive() {
			components = append(components, models.SystemComponent{
				Component:   "broker",
				DesiredPods: 1,
				ActualPods:  1,
			})
		} else {
			components = append(components, models.SystemComponent{
				Component:   "broker",
				DesiredPods: 1,
				ActualPods:  0,
			})
		}

		err = db.Client.Ping(context.TODO(), nil)
		if err != nil {
			components = append(components, models.SystemComponent{
				Component:   "mongodb",
				DesiredPods: 1,
				ActualPods:  0,
			})
		} else {
			components = append(components, models.SystemComponent{
				Component:   "mongodb",
				DesiredPods: 1,
				ActualPods:  1,
			})
		}

		components = append(components, models.SystemComponent{
			Component:   "control-plane",
			DesiredPods: 1,
			ActualPods:  1,
		})
	} else { // k8s env
		if clientset == nil {
			err := clientSetConfig()
			if err != nil {
				return components, err
			}
		}

		deploymentsClient := clientset.AppsV1().Deployments(configuration.K8S_NAMESPACE)
		deploymentsList, err := deploymentsClient.List(context.TODO(), metav1.ListOptions{})
		if err != nil {
			return components, err
		}

		for _, d := range deploymentsList.Items {
			if !strings.Contains(d.GetName(), "busybox") { // TODO remove it when busybox is getting fixed
				components = append(components, models.SystemComponent{
					Component:   d.GetName(),
					DesiredPods: int(*d.Spec.Replicas),
					ActualPods:  int(d.Status.ReadyReplicas),
				})
			}
		}

		statefulsetsClient := clientset.AppsV1().StatefulSets(configuration.K8S_NAMESPACE)
		statefulsetsList, err := statefulsetsClient.List(context.TODO(), metav1.ListOptions{})
		if err != nil {
			return components, err
		}
		for _, s := range statefulsetsList.Items {
			components = append(components, models.SystemComponent{
				Component:   s.GetName(),
				DesiredPods: int(*s.Spec.Replicas),
				ActualPods:  int(s.Status.ReadyReplicas),
			})
		}
	}

	return components, nil
}

func (mh MonitoringHandler) GetClusterInfo(c *gin.Context) {
	body, err := ioutil.ReadFile("version.conf")
	if err != nil {
		logger.Error("GetClusterInfo error: " + err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}
	c.IndentedJSON(200, gin.H{"version": string(body)})
}
