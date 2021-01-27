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

package test

import (
	"net"
	"testing"

	"magma/orc8r/cloud/go/service"
	"magma/orc8r/cloud/go/service/middleware/unary"
	"magma/orc8r/cloud/go/services/service_registry"
	"magma/orc8r/lib/go/registry"
	platform_service "magma/orc8r/lib/go/service"

	"github.com/labstack/echo"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
)

// NewService creates and registers a basic test Magma service on a dynamically
// selected available local port.
// Returns the newly created service and listener it was registered with.
func NewService(t *testing.T, moduleName string, serviceName string) (*platform_service.Service, net.Listener) {
	if t == nil {
		panic("for tests only")
	}

	lis, srvPort, err := getListener()
	if err != nil {
		t.Fatal(err)
	}

	location := registry.ServiceLocation{
		Name: serviceName,
		Host: "localhost",
		Port: srvPort,
	}
	// Add to both cloud and gateway registry so gateway-emulating code can
	// access the service
	service_registry.AddServices(location)
	registry.AddServices(location)

	srv, err := platform_service.NewServiceWithOptions(moduleName, serviceName, grpc.UnaryInterceptor(unary.MiddlewareHandler))
	if err != nil {
		t.Fatal(err)
	}

	return srv, lis
}

// NewOrchestratorService creates and registers a test Orchestrator service
// on a dynamically selected available local port for the gRPC server and HTTP
// echo server. Returns the newly created service and the gRPC listener it was
// registered with.
func NewOrchestratorService(
	t *testing.T,
	moduleName string,
	serviceName string,
	labels map[string]string,
	annotations map[string]string,
) (*service.OrchestratorService, net.Listener) {
	if t == nil {
		panic("for tests only")
	}

	if labels == nil {
		labels = map[string]string{}
	}
	if annotations == nil {
		annotations = map[string]string{}
	}

	srvLis, srvPort, err := getListener()
	if err != nil {
		t.Fatal(err)
	}
	echoSrv, echoPort, err := getEchoServer()
	if err != nil {
		t.Fatal(err)
	}

	location := registry.ServiceLocation{
		Name:        serviceName,
		Host:        "localhost",
		Port:        srvPort,
		EchoPort:    echoPort,
		Labels:      labels,
		Annotations: annotations,
	}
	// Add to both cloud and gateway registry so gateway-emulating code can
	// access the service
	service_registry.AddServices(location)
	registry.AddServices(location)

	platformService, err := platform_service.NewServiceWithOptions(moduleName, serviceName, grpc.UnaryInterceptor(unary.MiddlewareHandler))
	if err != nil {
		t.Fatal(err)
	}

	srv := &service.OrchestratorService{Service: platformService, EchoServer: echoSrv}
	return srv, srvLis
}

// getListener returns a listener at an available port, along with the chosen
// port.
func getListener() (net.Listener, int, error) {
	lis, err := net.Listen("tcp", "")
	if err != nil {
		return nil, 0, errors.Wrap(err, "error creating listener")
	}
	addr, err := net.ResolveTCPAddr("tcp", lis.Addr().String())
	if err != nil {
		return nil, 0, errors.Wrap(err, "error resolving TCP address")
	}
	return lis, addr.Port, err
}

// getEchoServer returns an echo HTTP server at an available port, along with
// the chosen port.
func getEchoServer() (*echo.Echo, int, error) {
	e := echo.New()
	e.HideBanner = true

	lis, port, err := getListener()
	if err != nil {
		return nil, 0, err
	}
	e.Listener = lis

	return e, port, nil
}
