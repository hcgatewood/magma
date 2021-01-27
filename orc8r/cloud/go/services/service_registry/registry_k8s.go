/*
 Copyright 2020 The Magma Authors.

 This source code is licensed under the BSD-style license found in the
 LICENSE file in the root directory of this source tree.

 Unless required by applicable law or agreed to in writing, software
 distributed under the License is distributed on an "AS IS" BASIS,
 WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 See the License for the specific language governing permissions and
 limitations under the License.
*/

package service_registry

import (
	"fmt"
	"strings"
	"sync"

	"magma/orc8r/cloud/go/orc8r"
	"magma/orc8r/lib/go/registry"

	"google.golang.org/grpc"
)

type k8sRegistry struct {
	sync.RWMutex
	conns registry.Connections
}

func NewK8sRegistry() Registry {
	return &k8sRegistry{conns: registry.Connections{}}
}

// AddServices etc are noops for immutable registries.
func (k *k8sRegistry) AddServices(locations ...registry.ServiceLocation) {}
func (k *k8sRegistry) RemoveService(service string)                      {}
func (k *k8sRegistry) RemoveServicesWithLabel(label string)              {}

func (k *k8sRegistry) ListAllServices() ([]string, error) {
	return listAllServices()
}

func (k *k8sRegistry) FindServices(label string) ([]string, error) {
	return findServices(label)
}

func (k *k8sRegistry) GetAddress(service string) (string, error) {
	service = strings.ToLower(service)
	if service == ServiceName {
		return getServiceRegistryAddress(), nil
	}
	return getAddress(service)
}

func (k *k8sRegistry) GetPort(service string) (int, error) {
	return orc8r.GRPCServicePort, nil
}

func (k *k8sRegistry) GetHTTPServerAddress(service string) (string, error) {
	service = strings.ToLower(service)
	return getHTTPServerAddress(service)
}

func (k *k8sRegistry) GetHTTPServerPort(service string) (int, error) {
	return orc8r.HTTPServerPort, nil
}

func (k *k8sRegistry) GetAnnotation(service, annotationName string) (string, error) {
	service = strings.ToLower(service)
	return getAnnotation(service, annotationName)
}

func (k *k8sRegistry) GetConn(service string) *grpc.ClientConn {
	service = strings.ToLower(service)
	return k.conns[service]
}

func (k *k8sRegistry) SetConn(service string, conn *grpc.ClientConn) {
	service = strings.ToLower(service)
	k.conns[service] = conn
}

// getServiceRegistryAddress uses hardcoded address for service_registry
// service since we can't dynamically discover the service registry service
// itself.
func getServiceRegistryAddress() string {
	return fmt.Sprintf("orc8r-service-registry:%d", orc8r.GRPCServicePort)
}
