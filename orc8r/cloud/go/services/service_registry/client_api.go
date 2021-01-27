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

// Client API is private. Access through a Registry.

package service_registry

import (
	"context"

	"magma/orc8r/cloud/go/services/service_registry/protos"
	merrors "magma/orc8r/lib/go/errors"

	"github.com/golang/glog"
)

const (
	ServiceName = "service_registry"
)

// listAllServices returns the service names of all registered services.
func listAllServices() ([]string, error) {
	c, err := getClient()
	if err != nil {
		return nil, err
	}

	req := &protos.ListAllServicesRequest{}
	res, err := c.ListAllServices(context.Background(), req)
	if err != nil {
		return nil, err
	}

	return res.Services, nil
}

// findServices returns all services that have the provided label.
func findServices(label string) ([]string, error) {
	c, err := getClient()
	if err != nil {
		return nil, err
	}

	req := &protos.FindServicesRequest{
		Label: label,
	}
	res, err := c.FindServices(context.Background(), req)
	if err != nil {
		return []string{}, err
	}

	return res.GetServices(), nil
}

// getAddress return the address of the gRPC server for the provided
// service.
func getAddress(service string) (string, error) {
	c, err := getClient()
	if err != nil {
		return "", err
	}

	req := &protos.GetAddressRequest{
		Service: service,
	}
	res, err := c.GetAddress(context.Background(), req)
	if err != nil {
		return "", err
	}

	return res.GetAddress(), nil
}

// getHttpServerAddress returns the address of the HTTP server for the provided
// service.
func getHTTPServerAddress(service string) (string, error) {
	c, err := getClient()
	if err != nil {
		return "", err
	}

	req := &protos.GetHTTPServerAddressRequest{
		Service: service,
	}
	res, err := c.GetHTTPServerAddress(context.Background(), req)
	if err != nil {
		return "", err
	}

	return res.GetAddress(), nil
}

// getAnnotation returns the annotation value for the provided service and
// annotation.
func getAnnotation(service string, annotation string) (string, error) {
	c, err := getClient()
	if err != nil {
		return "", err
	}

	req := &protos.GetAnnotationRequest{
		Service:    service,
		Annotation: annotation,
	}
	res, err := c.GetAnnotation(context.Background(), req)
	if err != nil {
		return "", err
	}

	return res.GetAnnotationValue(), nil
}

func getClient() (protos.ServiceRegistryClient, error) {
	conn, err := GetConnection(ServiceName)
	if err != nil {
		initErr := merrors.NewInitError(err, ServiceName)
		glog.Error(initErr)
		return nil, initErr
	}
	return protos.NewServiceRegistryClient(conn), nil
}
