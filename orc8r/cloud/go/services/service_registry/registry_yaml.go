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

	"magma/orc8r/lib/go/registry"

	"google.golang.org/grpc"
)

type yamlRegistry struct {
	sync.RWMutex
	locations registry.Locations
	conns     registry.Connections
}

func NewYAMLRegistry() Registry {
	return &yamlRegistry{locations: registry.Locations{}, conns: registry.Connections{}}
}

func (y *yamlRegistry) ListAllServices() ([]string, error) {
	y.RLock()
	defer y.RUnlock()
	var services []string
	for service := range y.locations {
		services = append(services, service)
	}
	return services, nil
}

func (y *yamlRegistry) FindServices(label string) ([]string, error) {
	y.RLock()
	defer y.RUnlock()
	var services []string
	for service, location := range y.locations {
		if location.HasLabel(label) {
			services = append(services, service)
		}
	}
	return services, nil
}

func (y *yamlRegistry) GetAddress(service string) (string, error) {
	y.RLock()
	defer y.RUnlock()

	service = strings.ToLower(service)
	location, ok := y.locations[service]
	if !ok {
		return "", fmt.Errorf("service %s not registered", service)
	}

	if location.Port == 0 {
		return location.Host, nil
	}
	return fmt.Sprintf("%s:%d", location.Host, location.Port), nil
}

func (y *yamlRegistry) GetPort(service string) (int, error) {
	y.RLock()
	defer y.RUnlock()

	service = strings.ToLower(service)
	location, ok := y.locations[service]
	if !ok {
		return 0, fmt.Errorf("service %s not registered", service)
	}
	if location.Port == 0 {
		return 0, fmt.Errorf("service %s not available", service)
	}

	return location.Port, nil
}

func (y *yamlRegistry) GetHTTPServerAddress(service string) (string, error) {
	y.RLock()
	defer y.RUnlock()

	service = strings.ToLower(service)
	location, ok := y.locations[service]
	if !ok {
		return "", fmt.Errorf("service %s not registered", service)
	}

	if location.EchoPort == 0 {
		return "", fmt.Errorf("service %s is not available", service)
	}
	return fmt.Sprintf("%s:%d", location.Host, location.EchoPort), nil
}

func (y *yamlRegistry) GetHTTPServerPort(service string) (int, error) {
	y.RLock()
	defer y.RUnlock()

	service = strings.ToLower(service)
	location, ok := y.locations[service]
	if !ok {
		return 0, fmt.Errorf("service %s not registered", service)
	}
	if location.EchoPort == 0 {
		return 0, fmt.Errorf("service %s HTTP server port not available", service)
	}
	return location.EchoPort, nil
}

func (y *yamlRegistry) GetAnnotation(service, annotationName string) (string, error) {
	y.RLock()
	defer y.RUnlock()

	service = strings.ToLower(service)
	location, ok := y.locations[strings.ToLower(service)]
	if !ok {
		return "", fmt.Errorf("service %s not registered", service)
	}

	annotationValue, ok := location.Annotations[annotationName]
	if !ok {
		return "", fmt.Errorf("service %s doesn't have annotation values for %s", service, annotationName)
	}

	return annotationValue, nil
}

func (y *yamlRegistry) AddServices(locations ...registry.ServiceLocation) {
	y.Lock()
	defer y.Unlock()

	for _, location := range locations {
		location.Name = strings.ToLower(location.Name)
		y.addUnsafe(location)
	}
}

func (y *yamlRegistry) RemoveService(service string) {
	y.Lock()
	defer y.Unlock()
	service = strings.ToLower(service)

	delete(y.locations, service)
	delete(y.conns, service)
}

func (y *yamlRegistry) RemoveServicesWithLabel(label string) {
	y.Lock()
	defer y.Unlock()

	for service, location := range y.locations {
		if location.HasLabel(label) {
			delete(y.locations, service)
			delete(y.conns, service)
		}
	}
}

func (y *yamlRegistry) GetConn(service string) *grpc.ClientConn {
	service = strings.ToLower(service)
	return y.conns[service]
}

func (y *yamlRegistry) SetConn(service string, conn *grpc.ClientConn) {
	service = strings.ToLower(service)
	y.conns[service] = conn
}

func (y *yamlRegistry) addUnsafe(location registry.ServiceLocation) {
	y.locations[location.Name] = location
	delete(y.conns, location.Name)
}
