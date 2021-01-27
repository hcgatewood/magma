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
	"fmt"
	"strings"
	"sync"
	"time"

	msync "magma/orc8r/lib/go/sync"

	"github.com/golang/glog"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
)

type ServiceRegistry struct {
	sync.RWMutex
	ServiceConnections Connections
	ServiceLocations   Locations

	cloudConnMu      sync.RWMutex
	cloudConnections map[string]cloudConnection
}

// Connections by service name.
type Connections map[string]*grpc.ClientConn

// Locations by service name.
type Locations map[string]ServiceLocation

type cloudConnection struct {
	*grpc.ClientConn
	expiration time.Time
}

var localKeepaliveParams = keepalive.ClientParameters{
	Time:                31 * time.Second,
	Timeout:             10 * time.Second,
	PermitWithoutStream: true,
}

// New creates and returns a new registry
func New() *ServiceRegistry {
	reg := &ServiceRegistry{
		ServiceConnections: map[string]*grpc.ClientConn{},
		ServiceLocations:   map[string]ServiceLocation{},
		cloudConnections:   map[string]cloudConnection{},
	}
	return reg
}

// AddService add a new service.
// If the service already exists, overwrites the service config.
func (r *ServiceRegistry) AddService(location ServiceLocation) {
	r.Lock()
	defer r.Unlock()
	location.Name = strings.ToLower(location.Name)

	r.addUnsafe(location)
}

// AddServices adds new services to the registry.
// If any services already exist, their locations will be overwritten
func (r *ServiceRegistry) AddServices(locations ...ServiceLocation) {
	r.Lock()
	defer r.Unlock()

	for _, location := range locations {
		location.Name = strings.ToLower(location.Name)
		r.addUnsafe(location)
	}
}

// GetAddress is an alias for GetServiceAddress.
// HACK: this alias solves different interface naming conventions across
// cloud and gateway. A better solution is to rectify the naming divergence.
func (r *ServiceRegistry) GetAddress(service string) (string, error) {
	return r.GetServiceAddress(service)
}

// GetServiceAddress returns the RPC address of the service.
// The service needs to be added to the registry before this.
func (r *ServiceRegistry) GetServiceAddress(service string) (string, error) {
	r.RLock()
	defer r.RUnlock()

	service = strings.ToLower(service)
	location, ok := r.ServiceLocations[service]
	if !ok {
		return "", fmt.Errorf("service %s not registered", service)
	}

	if location.Port == 0 {
		return location.Host, nil
	}
	return fmt.Sprintf("%s:%d", location.Host, location.Port), nil
}

// GetServicePort returns the listening port for the RPC service.
// The service needs to be added to the registry before this.
func (r *ServiceRegistry) GetServicePort(service string) (int, error) {
	r.RLock()
	defer r.RUnlock()

	service = strings.ToLower(service)
	location, ok := r.ServiceLocations[service]
	if !ok {
		return 0, fmt.Errorf("service %s not registered", service)
	}
	if location.Port == 0 {
		return 0, fmt.Errorf("service %s not available", service)
	}

	return location.Port, nil
}

func (r *ServiceRegistry) GetConn(service string) *grpc.ClientConn {
	return r.ServiceConnections[service]
}

func (r *ServiceRegistry) SetConn(service string, conn *grpc.ClientConn) {
	r.ServiceConnections[service] = conn
}

// GetConnection provides a gRPC connection to a service in the registry.
// The service needs to be added to the registry before this.
func (r *ServiceRegistry) GetConnection(service string) (*grpc.ClientConn, error) {
	return GetServiceConnection(service, r, GetDefaultGatewayDialOpts())
}

func (r *ServiceRegistry) GetConnectionImpl(ctx context.Context, service string, opts ...grpc.DialOption) (*grpc.ClientConn, error) {
	return GetServiceConnectionImpl(ctx, service, r, opts...)
}

func (r *ServiceRegistry) addUnsafe(location ServiceLocation) {
	if r.ServiceLocations == nil {
		r.ServiceLocations = map[string]ServiceLocation{}
	}
	r.ServiceLocations[location.Name] = location
	delete(r.ServiceConnections, location.Name)
}

// GetDefaultCloudDialOpts is the default dial options for cloud services.
func GetDefaultCloudDialOpts() []grpc.DialOption {
	opts := []grpc.DialOption{
		grpc.WithBackoffMaxDelay(GrpcMaxDelaySec * time.Second),
		grpc.WithBlock(),
		grpc.WithUnaryInterceptor(CloudClientTimeoutInterceptor),
	}
	return opts
}

// GetDefaultGatewayDialOpts is the default dial options for gateway services.
func GetDefaultGatewayDialOpts() []grpc.DialOption {
	opts := []grpc.DialOption{
		grpc.WithBackoffMaxDelay(GrpcMaxDelaySec * time.Second),
		grpc.WithBlock(),
		grpc.WithUnaryInterceptor(TimeoutInterceptor),
	}
	return opts
}

// Conns provides methods for reading and writing gRPC connections.
// The caller is responsible for handling any required synchronization, e.g.
// locking a mutex.
type Conns interface {
	// GetConn returns the connection for the passed service, or nil if there
	// is none.
	GetConn(service string) *grpc.ClientConn

	// SetConn sets the connection for the passed service, overwriting any
	// previous value.
	SetConn(service string, conn *grpc.ClientConn)
}

// dialRegistry defines the minimal registry requirements needed to generate
// and set a client connection in the registry.
type dialRegistry interface {
	msync.RWMutex
	Conns
	GetAddress(service string) (string, error)
}

// GetServiceConnection provides a gRPC connection to a service in the registry.
func GetServiceConnection(service string, reg dialRegistry, opts []grpc.DialOption) (*grpc.ClientConn, error) {
	service = strings.ToLower(service)
	ctx, cancel := context.WithTimeout(context.Background(), GrpcMaxTimeoutSec*time.Second)
	defer cancel()
	return GetServiceConnectionImpl(ctx, service, reg, opts...)
}

func GetServiceConnectionImpl(ctx context.Context, service string, reg dialRegistry, opts ...grpc.DialOption) (*grpc.ClientConn, error) {
	service = strings.ToLower(service)

	// First try to get an existing connection with reader lock
	reg.RLock()
	conn := reg.GetConn(service)
	reg.RUnlock()
	if conn != nil {
		return conn, nil
	}

	// Attempt to connect outside of the lock
	// Each attempt to get client connection has a long timeout. Connecting
	// without the lock prevents callers from timing out waiting for the
	// lock to a bad connection.
	addr, err := reg.GetAddress(service)
	if err != nil {
		return nil, err
	}
	newConn, err := GetClientConnection(ctx, addr, opts...)
	if err != nil || newConn == nil {
		return newConn, fmt.Errorf("service %v connection error: %s", service, err)
	}

	reg.Lock()
	defer reg.Unlock()

	// Re-check after taking the lock
	conn = reg.GetConn(service)
	if conn != nil {
		// Another routine already added the connection for the service, clean up ours & return existing
		err := newConn.Close()
		if err != nil {
			glog.Errorf("Error closing unneeded gRPC connection: %v", err)
		}
		return conn, nil
	}

	reg.SetConn(service, newConn)
	return newConn, nil
}

// GetClientConnection provides a gRPC connection to a service on the address addr.
func GetClientConnection(ctx context.Context, addr string, opts ...grpc.DialOption) (*grpc.ClientConn, error) {
	opts = append(opts, grpc.WithInsecure())
	conn, err := grpc.DialContext(ctx, addr, opts...)
	if err != nil {
		return nil, fmt.Errorf("address: %s gRPC Dial error: %s", addr, err)
	} else if ctx.Err() != nil {
		return nil, ctx.Err()
	}
	return conn, nil
}
