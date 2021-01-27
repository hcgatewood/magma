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
	"context"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"magma/orc8r/lib/go/registry"

	"github.com/golang/glog"
	"google.golang.org/grpc"
)

const (
	serviceRegistryModeEnvVar = "SERVICE_REGISTRY_MODE"
	k8sRegistryMode           = "k8s"
	yamlRegistryMode          = "yaml"
)

var (
	globalReg = NewRegistry()
)

func NewRegistry() Registry {
	mode := os.Getenv(serviceRegistryModeEnvVar)
	switch mode {
	case yamlRegistryMode:
		return NewYAMLRegistry()
	case k8sRegistryMode:
		return NewK8sRegistry()
	default:
		// Default to local registry for tests, but with warning
		glog.Warningf(
			"Unrecognized service registry mode ('%s'). Set %s environment variable to '%s' or '%s'.",
			mode, serviceRegistryModeEnvVar, yamlRegistryMode, k8sRegistryMode,
		)
		return NewYAMLRegistry()
	}
}

// MustPopulateServices is same as PopulateServices but fails on errors.
func MustPopulateServices() {
	if err := PopulateServices(); err != nil {
		glog.Fatalf("Error populating services: %+v", err)
	}
}

// PopulateServices populates the service registry based on the per-module
// config files at /etc/magma/configs/MODULE_NAME/service_registry.yml.
func PopulateServices() error {
	serviceConfigs, err := registry.LoadServiceRegistryConfigs()
	if err != nil {
		return err
	}
	AddServices(serviceConfigs...)
	return nil
}

// todo hcg standardize grpc and http ports for k8s deployment

func GetConnection(service string) (*grpc.ClientConn, error) {
	return registry.GetServiceConnection(service, globalReg, registry.GetDefaultCloudDialOpts())
}

// GetGatewayConnectionForTest returns a connection to a service, emulating
// the connection from a gateway.
func GetGatewayConnectionForTest(t *testing.T, service string) *grpc.ClientConn {
	if t == nil {
		panic("for tests only")
	}
	service = strings.ToLower(service)

	addr, err := globalReg.GetAddress(service)
	assert.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), registry.GrpcMaxTimeoutSec*time.Second)
	defer cancel()
	// Get a fresh, non-cached connection to avoid sharing connections with
	// the cloud registry
	conn, err := registry.GetClientConnection(ctx, addr, registry.GetDefaultGatewayDialOpts()...)
	assert.NoError(t, err)
	return conn
}

func GetConnectionWithOptions(ctx context.Context, service string, opts ...grpc.DialOption) (*grpc.ClientConn, error) {
	return registry.GetServiceConnectionImpl(ctx, service, globalReg, opts...)
}

func ListAllServices() ([]string, error) {
	return globalReg.ListAllServices()
}

func FindServices(label string) ([]string, error) {
	return globalReg.FindServices(label)
}

func GetAddress(service string) (string, error) {
	return globalReg.GetAddress(service)
}

func GetPort(service string) (int, error) {
	return globalReg.GetPort(service)
}

func GetHTTPServerAddress(service string) (string, error) {
	return globalReg.GetHTTPServerAddress(service)
}

func GetHTTPServerPort(service string) (int, error) {
	return globalReg.GetHTTPServerPort(service)
}

func GetAnnotation(service, annotationName string) (string, error) {
	return globalReg.GetAnnotation(service, annotationName)
}

// GetAnnotationList returns the comma-split fields of the value for the passed
// annotation name.
// First splits by field separator, then strips all whitespace
// (including newlines). Empty fields are removed.
func GetAnnotationList(service, annotationName string) ([]string, error) {
	return GetAnnotationListFromRegistry(globalReg, service, annotationName)
}

func AddServices(locations ...registry.ServiceLocation) {
	globalReg.AddServices(locations...)
}

func RemoveService(service string) {
	globalReg.RemoveService(service)
}

func RemoveServicesWithLabel(label string) {
	globalReg.RemoveServicesWithLabel(label)
}
