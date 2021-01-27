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

package service303

import (
	"github.com/golang/glog"
	"golang.org/x/net/context"

	"magma/orc8r/cloud/go/services/service_registry"
	"magma/orc8r/lib/go/errors"
	"magma/orc8r/lib/go/protos"
)

func GetServiceInfo(service string) (*protos.ServiceInfo, error) {
	client, err := getClient(service)
	if err != nil {
		return nil, err
	}
	return client.GetServiceInfo(context.Background(), new(protos.Void))
}

func GetMetrics(service string) (*protos.MetricsContainer, error) {
	client, err := getClient(service)
	if err != nil {
		return nil, err
	}
	return client.GetMetrics(context.Background(), new(protos.Void))
}

func StopService(service string) error {
	client, err := getClient(service)
	if err != nil {
		return err
	}
	_, err = client.StopService(context.Background(), new(protos.Void))
	return err
}

func SetLogLevel(service string, in *protos.LogLevelMessage) error {
	client, err := getClient(service)
	if err != nil {
		return err
	}
	_, err = client.SetLogLevel(context.Background(), in)
	return err
}

func SetLogVerbosity(service string, in *protos.LogVerbosity) error {
	client, err := getClient(service)
	if err != nil {
		return err
	}
	_, err = client.SetLogVerbosity(context.Background(), in)
	return err
}

func getClient(service string) (protos.Service303Client, error) {
	conn, err := service_registry.GetConnection(service)
	if err != nil {
		initErr := errors.NewInitError(err, "SERVICE303")
		glog.Error(initErr)
		return nil, initErr
	}
	return protos.NewService303Client(conn), nil
}
