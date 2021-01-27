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

package servicers

import (
	"fmt"
	"os"
	"strings"

	"magma/orc8r/cloud/go/services/service_registry/protos"

	"golang.org/x/net/context"
	k8score "k8s.io/api/core/v1"
	k8smeta "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8sclientcore "k8s.io/client-go/kubernetes/typed/core/v1"
)

const (
	namespaceEnvVar        = "SERVICE_REGISTRY_NAMESPACE"
	partOfLabel            = "app.kubernetes.io/part-of"
	partOfOrc8rApp         = "orc8r-app"
	orc8rServiceNamePrefix = "orc8r-"
	grpcPortName           = "grpc"
	httpPortName           = "http"
)

type k8sServiceRegistryServicer struct {
	client    k8sclientcore.CoreV1Interface
	namespace string
}

// NewK8sServiceRegistryServicer creates a new service registry servicer
// backed by Kubernetes.
func NewK8sServiceRegistryServicer(k8sClient k8sclientcore.CoreV1Interface) (protos.ServiceRegistryServer, error) {
	namespace := os.Getenv(namespaceEnvVar)
	if namespace == "" {
		return nil, fmt.Errorf("environment variable %s must be set to the Helm deployment's release namespace", namespaceEnvVar)
	}
	srv := &k8sServiceRegistryServicer{client: k8sClient, namespace: namespace}
	return srv, nil
}

func (k *k8sServiceRegistryServicer) ListAllServices(ctx context.Context, req *protos.ListAllServicesRequest) (*protos.ListAllServicesResponse, error) {
	res := &protos.ListAllServicesResponse{}
	listOptions := k8smeta.ListOptions{
		LabelSelector: fmt.Sprintf("%s=%s", partOfLabel, partOfOrc8rApp),
	}
	services, err := k.client.Services(k.namespace).List(listOptions)
	if err != nil {
		return nil, err
	}
	for _, s := range services.Items {
		res.Services = append(res.Services, k.convertK8sServiceNameToMagmaServiceName(s.Name))
	}
	return res, nil
}

func (k *k8sServiceRegistryServicer) FindServices(ctx context.Context, req *protos.FindServicesRequest) (*protos.FindServicesResponse, error) {
	res := &protos.FindServicesResponse{}
	listOptions := k8smeta.ListOptions{
		LabelSelector: fmt.Sprintf("%s=%s,%s=true", partOfLabel, partOfOrc8rApp, req.Label),
	}
	services, err := k.client.Services(k.namespace).List(listOptions)
	if err != nil {
		return nil, err
	}
	for _, s := range services.Items {
		res.Services = append(res.Services, k.convertK8sServiceNameToMagmaServiceName(s.Name))
	}
	return res, nil
}

func (k *k8sServiceRegistryServicer) GetAddress(ctx context.Context, req *protos.GetAddressRequest) (*protos.GetAddressResponse, error) {
	serviceAddress, err := k.getAddressForPortName(req.Service, grpcPortName)
	if err != nil {
		return nil, err
	}
	res := &protos.GetAddressResponse{Address: serviceAddress}
	return res, nil
}

func (k *k8sServiceRegistryServicer) GetHTTPServerAddress(ctx context.Context, req *protos.GetHTTPServerAddressRequest) (*protos.GetHTTPServerAddressResponse, error) {
	httpServerAddress, err := k.getAddressForPortName(req.Service, httpPortName)
	if err != nil {
		return nil, err
	}
	res := &protos.GetHTTPServerAddressResponse{Address: httpServerAddress}
	return res, nil
}

func (k *k8sServiceRegistryServicer) GetAnnotation(ctx context.Context, req *protos.GetAnnotationRequest) (*protos.GetAnnotationResponse, error) {
	service, err := k.getServiceForServiceName(req.Service)
	if err != nil {
		return nil, err
	}
	for annotation, value := range service.GetAnnotations() {
		if annotation == req.Annotation {
			return &protos.GetAnnotationResponse{AnnotationValue: value}, nil
		}
	}
	return nil, fmt.Errorf("annotation '%s' not found for service '%s'", req.Annotation, req.Service)
}

func (k *k8sServiceRegistryServicer) getAddressForPortName(serviceName string, portName string) (string, error) {
	service, err := k.getServiceForServiceName(serviceName)
	if err != nil {
		return "", err
	}
	for _, port := range service.Spec.Ports {
		if port.Name == portName {
			return fmt.Sprintf("%s:%d", service.Name, port.Port), nil
		}
	}
	return "", fmt.Errorf("could not find port '%s' for service '%s'", portName, serviceName)
}

func (k *k8sServiceRegistryServicer) getServiceForServiceName(serviceName string) (*k8score.Service, error) {
	orc8rListOption := k8smeta.ListOptions{
		LabelSelector: fmt.Sprintf("%s=%s", partOfLabel, partOfOrc8rApp),
	}

	formattedServiceName := k.convertMagmaServiceNameToK8sServiceName(serviceName)
	services, err := k.client.Services(k.namespace).List(orc8rListOption)
	if err != nil {
		return nil, err
	}
	for _, s := range services.Items {
		if s.Name == formattedServiceName {
			return &s, nil
		}
	}
	return nil, fmt.Errorf("could not find service '%s'", serviceName)
}

// Orc8r Helm services are formatted as orc8r-<service-name>. Magma convention
// is to use underscores in service names, so remove prefix and convert any
// hyphens in the K8s service name.
func (k *k8sServiceRegistryServicer) convertK8sServiceNameToMagmaServiceName(serviceName string) string {
	orc8rServiceNameSuffix := strings.TrimPrefix(serviceName, orc8rServiceNamePrefix)
	return strings.ReplaceAll(orc8rServiceNameSuffix, "-", "_")
}

// Orc8r helm services are formatted as orc8r-<service-name>. Magma convention
// is to use underscores in service names, so add prefix and convert any
// underscores to hyphens.
func (k *k8sServiceRegistryServicer) convertMagmaServiceNameToK8sServiceName(serviceName string) string {
	k8sServiceNameSuffix := strings.ReplaceAll(serviceName, "_", "-")
	return fmt.Sprintf("%s%s", orc8rServiceNamePrefix, k8sServiceNameSuffix)
}
