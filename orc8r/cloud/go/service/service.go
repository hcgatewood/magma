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

// Package service outlines the Magma microservices framework in the cloud.
// The framework helps to create a microservice easily, and provides
// the common service logic like service303, config, etc.
package service

import (
	"context"
	"flag"
	"fmt"
	"net"
	"testing"

	"magma/orc8r/cloud/go/plugin"
	"magma/orc8r/cloud/go/service/middleware/unary"
	service_registry "magma/orc8r/cloud/go/services/service_registry"
	"magma/orc8r/lib/go/protos"
	platform_service "magma/orc8r/lib/go/service"

	"github.com/golang/glog"
	"github.com/labstack/echo"
	"google.golang.org/grpc"
)

const (
	RunEchoServerFlag = "run_echo_server"
)

var (
	runEchoServer bool
)

func init() {
	flag.BoolVar(&runEchoServer, RunEchoServerFlag, false, "Run echo HTTP server with service")
}

// OrchestratorService defines a service which extends the generic platform
// service with an optional HTTP server.
type OrchestratorService struct {
	*platform_service.Service

	// EchoServer runs on the echo_port specified in the registry.
	// This field will be nil for services that don't specify the
	// 'run_echo_server' flag.
	EchoServer *echo.Echo
}

// NewOrchestratorService returns a new gRPC orchestrator service
// implementing service303. If configured, it will also initialize an HTTP echo
// server as a part of the service. This service will implement a middleware
// interceptor to perform identity check. If your service does not or can not
// perform identity checks, (e.g., federation), use NewServiceWithOptions.
func NewOrchestratorService(moduleName string, serviceName string, serverOptions ...grpc.ServerOption) (*OrchestratorService, error) {
	flag.Parse()
	plugin.LoadAllPluginsFatalOnError(&plugin.DefaultOrchestratorPluginLoader{})

	err := service_registry.PopulateServices()
	if err != nil {
		return nil, err
	}

	serverOptions = append(serverOptions, grpc.UnaryInterceptor(unary.MiddlewareHandler))
	platformService, err := platform_service.NewServiceWithOptionsImpl(moduleName, serviceName, serverOptions...)
	if err != nil {
		return nil, err
	}

	echoPort, err := service_registry.GetHTTPServerPort(serviceName)
	if err != nil {
		return nil, err
	}

	echoSrv, err := getEchoServer(echoPort)
	if err != nil {
		return nil, err
	}

	return &OrchestratorService{Service: platformService, EchoServer: echoSrv}, nil
}

// Run runs the service. If the echo HTTP server is non-nil, both the HTTP
// server and gRPC server are run, blocking until an error occurs or a server
// stopped. If the HTTP server is nil, only the gRPC server is run, blocking
// until its interrupted by a signal or until the gRPC server is stopped.
func (s *OrchestratorService) Run() error {
	port, err := service_registry.GetPort(s.Type)
	if err != nil {
		return err
	}

	if s.EchoServer == nil {
		return s.Service.RunWithPort(port)
	}

	serverErr := make(chan error)
	go func() {
		err := s.Service.RunWithPort(port)
		shutdownErr := s.EchoServer.Shutdown(context.Background())
		if shutdownErr != nil {
			glog.Errorf("Error shutting down echo server: %v", shutdownErr)
		}
		serverErr <- err
	}()
	go func() {
		err := s.EchoServer.StartServer(s.EchoServer.Server)
		_, shutdownErr := s.Service.StopService(context.Background(), &protos.Void{})
		if shutdownErr != nil {
			glog.Errorf("Error shutting down orc8r service: %v", shutdownErr)
		}
		serverErr <- err
	}()
	return <-serverErr
}

// RunTest runs the test service on a given Listener and the HTTP on it's
// configured addr if exists. This function blocks by a signal or until a
// server is stopped.
func (s *OrchestratorService) RunTest(t *testing.T, lis net.Listener) {
	if t == nil {
		panic("for tests only")
	}
	serverErr := make(chan error)
	go func() {
		err := s.Service.RunTest(t, lis)
		serverErr <- err
	}()
	if s.EchoServer != nil {
		go func() {
			err := s.EchoServer.StartServer(s.EchoServer.Server)
			serverErr <- err
		}()
	}
	err := <-serverErr
	if err != nil {
		t.Fatal(err)
	}
}

// getEchoServer returns an echo HTTP server at the specified port.
func getEchoServer(port int) (*echo.Echo, error) {
	if !runEchoServer {
		return nil, nil
	}
	e := echo.New()
	e.HideBanner = true
	e.Server.Addr = fmt.Sprintf(":%d", port)
	return e, nil
}
