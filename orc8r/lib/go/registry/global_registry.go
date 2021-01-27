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

// Package registry for Magma microservices
package registry

import (
	"google.golang.org/grpc"
)

const (
	GrpcMaxDelaySec        = 10
	GrpcMaxLocalTimeoutSec = 30
	GrpcMaxTimeoutSec      = 60
)

// globalRegistry is the global service registry instance
var globalRegistry = New()

// Get returns a reference to the instance of global platform registry
func Get() *ServiceRegistry {
	return globalRegistry
}

// AddService add a new service to global registry.
// If the service already exists, overwrites the service config.
func AddService(location ServiceLocation) {
	globalRegistry.AddService(location)
}

// AddServices adds new services to the global registry.
// If any services already exist, their locations will be overwritten
func AddServices(locations ...ServiceLocation) {
	globalRegistry.AddServices(locations...)
}

// GetServiceAddress returns the RPC address of the service from global registry
// The service needs to be added to the registry before this.
func GetServiceAddress(service string) (string, error) {
	return globalRegistry.GetServiceAddress(service)
}

// GetServicePort returns the listening port for the RPC service.
// The service needs to be added to the registry before this.
func GetServicePort(service string) (int, error) {
	return globalRegistry.GetServicePort(service)
}

// GetConnection provides a gRPC connection to a service in the registry.
func GetConnection(service string) (*grpc.ClientConn, error) {
	return GetServiceConnection(service, globalRegistry, GetDefaultGatewayDialOpts())
}
