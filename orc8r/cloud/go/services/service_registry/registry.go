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
	"strings"

	"magma/orc8r/cloud/go/orc8r"
	"magma/orc8r/lib/go/registry"
	"magma/orc8r/lib/go/sync"
)

// Registry of services.
type Registry interface {
	MutableRegistry
	registry.Conns
	sync.RWMutex

	// ListAllServices lists the names of all registered services.
	ListAllServices() ([]string, error)

	// FindServices returns the names of all registered services that have
	// the passed label.
	FindServices(label string) ([]string, error)

	// GetAddress returns the address of the service's RPC server.
	GetAddress(service string) (string, error)

	// GetPort returns the port of the service's RPC server.
	GetPort(service string) (int, error)

	// GetHTTPServerAddress returns the address of the service's HTTP server.
	GetHTTPServerAddress(service string) (string, error)

	// GetHTTPServerPort returns the port of the service's HTTP server.
	GetHTTPServerPort(service string) (int, error)

	// GetAnnotation returns the value for the passed annotation name.
	GetAnnotation(service, annotationName string) (string, error)
}

// MutableRegistry adds methods to manipulate the services in the service
// registry.
type MutableRegistry interface {
	// AddServices adds new services to the registry.
	// If any services already exist, their locations will be overwritten
	AddServices(locations ...registry.ServiceLocation)

	// RemoveService removes a service from the registry.
	// Has no effect if the service does not exist.
	RemoveService(service string)

	// RemoveServicesWithLabel removes all services from the registry which
	// have the passed label.
	RemoveServicesWithLabel(label string)
}

// GetAnnotationList returns the comma-split fields of the value for the passed
// annotation name.
// First splits by field separator, then strips all whitespace
// (including newlines). Empty fields are removed.
func GetAnnotationListFromRegistry(reg Registry, service, annotationName string) ([]string, error) {
	annotationValue, err := reg.GetAnnotation(service, annotationName)
	if err != nil {
		return nil, err
	}

	var values []string
	for _, s := range strings.Split(annotationValue, orc8r.AnnotationFieldSeparator) {
		s = strings.Join(strings.Fields(s), "")
		if s != "" {
			values = append(values, s)
		}
	}

	return values, nil
}
