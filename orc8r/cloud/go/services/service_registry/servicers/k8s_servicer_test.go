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
	"testing"

	"github.com/stretchr/testify/assert"
	"golang.org/x/net/context"
	k8score "k8s.io/api/core/v1"
	k8smeta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
	k8sclientcore "k8s.io/client-go/kubernetes/typed/core/v1"

	"magma/orc8r/cloud/go/orc8r"
	"magma/orc8r/cloud/go/services/service_registry/protos"
)

func TestK8sListAllServices(t *testing.T) {
	namespace := "test_namespace_listall"
	servicer, mockClient := setupTest(t, namespace)
	req := &protos.ListAllServicesRequest{}
	res, err := servicer.ListAllServices(context.Background(), req)
	assert.NoError(t, err)
	assert.Equal(t, []string{"service1", "service_2"}, res.Services)

	err = mockClient.Services(namespace).Delete("orc8r-service-2", &k8smeta.DeleteOptions{})
	assert.NoError(t, err)
	res, err = servicer.ListAllServices(context.Background(), req)
	assert.NoError(t, err)
	assert.Equal(t, []string{"service1"}, res.Services)
}

func TestK8sFindServices(t *testing.T) {
	namespace := "test_namespace_find"
	servicer, mockClient := setupTest(t, namespace)
	req := &protos.FindServicesRequest{
		Label: "label1",
	}
	res, err := servicer.FindServices(context.Background(), req)
	assert.NoError(t, err)
	assert.Equal(t, []string{"service1"}, res.Services)

	req.Label = "label2"
	res, err = servicer.FindServices(context.Background(), req)
	assert.NoError(t, err)
	assert.Equal(t, []string{"service1", "service_2"}, res.Services)

	err = mockClient.Services(namespace).Delete("orc8r-service1", &k8smeta.DeleteOptions{})
	assert.NoError(t, err)
	res, err = servicer.FindServices(context.Background(), req)
	assert.NoError(t, err)
	assert.Equal(t, []string{"service_2"}, res.Services)
}

func TestK8sGetAddress(t *testing.T) {
	namespace := "test_namespace_getaddr"
	servicer, mockClient := setupTest(t, namespace)
	req := &protos.GetAddressRequest{Service: "service1"}

	res, err := servicer.GetAddress(context.Background(), req)
	assert.NoError(t, err)
	assert.Equal(t, res.Address, fmt.Sprintf("orc8r-service1:%d", orc8r.GRPCServicePort))

	mockClient.Services(namespace).Delete("orc8r-service1", &k8smeta.DeleteOptions{})
	_, err = servicer.GetAddress(context.Background(), req)
	assert.Error(t, err)
}

func TestK8sGetHTTPServerAddress(t *testing.T) {
	namespace := "test_namespace_gethttp"
	servicer, mockClient := setupTest(t, namespace)
	req := &protos.GetHTTPServerAddressRequest{Service: "service1"}

	res, err := servicer.GetHTTPServerAddress(context.Background(), req)
	assert.NoError(t, err)
	assert.Equal(t, res.Address, fmt.Sprintf("orc8r-service1:%d", orc8r.HTTPServerPort))

	mockClient.Services(namespace).Delete("orc8r-service1", &k8smeta.DeleteOptions{})
	_, err = servicer.GetHTTPServerAddress(context.Background(), req)
	assert.Error(t, err)
}

func TestK8sGetAnnotation(t *testing.T) {
	namespace := "test_namespace_getannotation"
	servicer, mockClient := setupTest(t, namespace)
	req := &protos.GetAnnotationRequest{Service: "service1", Annotation: "annotation2"}

	res, err := servicer.GetAnnotation(context.Background(), req)
	assert.NoError(t, err)
	assert.Equal(t, "bar,baz", res.AnnotationValue)

	mockClient.Services(namespace).Delete("orc8r-service1", &k8smeta.DeleteOptions{})
	_, err = servicer.GetAnnotation(context.Background(), req)
	assert.Error(t, err)
}

func setupTest(t *testing.T, namespace string) (protos.ServiceRegistryServer, k8sclientcore.CoreV1Interface) {
	mockClient := fake.NewSimpleClientset().CoreV1()
	os.Setenv(namespaceEnvVar, namespace)

	servicer, err := NewK8sServiceRegistryServicer(mockClient)
	assert.NoError(t, err)
	createK8sServices(t, mockClient, namespace)
	return servicer, mockClient
}

func createK8sServices(t *testing.T, mockClient k8sclientcore.CoreV1Interface, namespace string) {
	svc1 := &k8score.Service{
		ObjectMeta: k8smeta.ObjectMeta{
			Namespace: namespace,
			Name:      "orc8r-service1",
			Labels: map[string]string{
				partOfLabel: partOfOrc8rApp,
				"label1":    "true",
				"label2":    "true",
			},
			Annotations: map[string]string{
				"annotation1": "foo",
				"annotation2": "bar,baz",
			},
		},
		Spec: k8score.ServiceSpec{
			Ports: []k8score.ServicePort{
				{
					Name: grpcPortName,
					Port: orc8r.GRPCServicePort,
				},
				{
					Name: httpPortName,
					Port: orc8r.HTTPServerPort,
				},
			},
			Type:      k8score.ServiceTypeClusterIP,
			ClusterIP: "127.0.0.1",
		},
	}
	svc2 := &k8score.Service{
		ObjectMeta: k8smeta.ObjectMeta{
			Namespace: namespace,
			Name:      "orc8r-service-2",
			Labels: map[string]string{
				partOfLabel: partOfOrc8rApp,
				"label2":    "true",
				"label3":    "true",
			},
			Annotations: map[string]string{
				"annotation3": "roo",
				"annotation4": "par,zaz",
			},
		},
		Spec: k8score.ServiceSpec{
			Ports: []k8score.ServicePort{
				{
					Name: grpcPortName,
					Port: orc8r.GRPCServicePort,
				},
			},
			Type:      k8score.ServiceTypeClusterIP,
			ClusterIP: "127.0.0.1",
		},
	}
	svc3 := &k8score.Service{
		ObjectMeta: k8smeta.ObjectMeta{
			Namespace: namespace,
			Name:      fmt.Sprintf("%s-%s", "nonorc8r", "service3"),
		},
		Spec: k8score.ServiceSpec{
			Type:      k8score.ServiceTypeClusterIP,
			ClusterIP: "127.0.0.1",
		},
	}
	_, err := mockClient.Services(namespace).Create(svc1)
	assert.NoError(t, err)
	_, err = mockClient.Services(namespace).Create(svc2)
	assert.NoError(t, err)
	_, err = mockClient.Services(namespace).Create(svc3)
	assert.NoError(t, err)
}
