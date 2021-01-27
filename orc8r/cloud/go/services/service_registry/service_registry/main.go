/*
 * Copyright 2020 The Magma Authors.
 *
 * This source code is licensed under the BSD-style license found in the
 * LICENSE file in the root directory of this source tree.
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package main

import (
	"magma/orc8r/cloud/go/orc8r"
	"magma/orc8r/cloud/go/service"
	"magma/orc8r/cloud/go/services/service_registry"
	"magma/orc8r/cloud/go/services/service_registry/protos"
	"magma/orc8r/cloud/go/services/service_registry/servicers"

	"github.com/golang/glog"
	"k8s.io/client-go/kubernetes"
	k8s_client "k8s.io/client-go/rest"
)

const (
	defaultK8sQPS   = 50
	defaultK8sBurst = 50
)

func main() {
	srv, err := service.NewOrchestratorService(orc8r.ModuleName, service_registry.ServiceName)
	if err != nil {
		glog.Fatalf("Error creating service_registry service %+v", err)
	}

	config, err := getK8sClientConfig()
	if err != nil {
		glog.Fatalf("Error querying k8s config: %+v", err)
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		glog.Fatalf("Error creating k8s clientset: %+v", err)
	}
	servicer, err := servicers.NewK8sServiceRegistryServicer(clientset.CoreV1())
	if err != nil {
		glog.Fatal(err)
	}
	protos.RegisterServiceRegistryServer(srv.GrpcServer, servicer)

	err = srv.Run()
	if err != nil {
		glog.Fatalf("Error while running service_registry service: %+v", err)
	}
}

func getK8sClientConfig() (*k8s_client.Config, error) {
	config, err := k8s_client.InClusterConfig()
	if err != nil {
		return nil, err
	}
	// TODO(hcgatewood): remove QPS and Burst overrides after adding cache.
	config.QPS = defaultK8sQPS
	config.Burst = defaultK8sBurst
	return config, err
}
