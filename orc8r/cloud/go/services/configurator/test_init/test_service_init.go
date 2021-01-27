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

package test_init

import (
	"fmt"
	"sync"
	"testing"

	"magma/orc8r/cloud/go/orc8r"
	"magma/orc8r/cloud/go/service/test"
	accessd_test_init "magma/orc8r/cloud/go/services/accessd/test_init"
	certifier_test_init "magma/orc8r/cloud/go/services/certifier/test_init"
	"magma/orc8r/cloud/go/services/configurator"
	"magma/orc8r/cloud/go/services/configurator/protos"
	"magma/orc8r/cloud/go/services/configurator/servicers"
	"magma/orc8r/cloud/go/services/configurator/storage"
	"magma/orc8r/cloud/go/sqorc"
)

func StartTestService(t *testing.T) {
	db, err := sqorc.Open("sqlite3", ":memory:?_foreign_keys=1")
	if err != nil {
		t.Fatalf("Could not initialize sqlite DB: %s", err)
	}
	idGenerator := sequentialIDGenerator{nextID: 1}
	storageFactory := storage.NewSQLConfiguratorStorageFactory(db, &idGenerator, sqorc.GetSqlBuilder())
	err = storageFactory.InitializeServiceStorage()
	if err != nil {
		t.Fatalf("Could not initialize storage: %s", err)
	}

	accessd_test_init.StartTestService(t)
	certifier_test_init.StartTestService(t)

	srv, lis := test.NewService(t, orc8r.ModuleName, configurator.ServiceName)
	nb, err := servicers.NewNorthboundConfiguratorServicer(storageFactory)
	if err != nil {
		t.Fatalf("Failed to create NB configurator servicer: %s", err)
	}
	protos.RegisterNorthboundConfiguratorServer(srv.GrpcServer, nb)

	sb, err := servicers.NewSouthboundConfiguratorServicer(storageFactory)
	if err != nil {
		t.Fatalf("Failed to create SB configurator servicer: %s", err)
	}
	protos.RegisterSouthboundConfiguratorServer(srv.GrpcServer, sb)

	go srv.MustRunTest(t, lis)
}

type sequentialIDGenerator struct {
	sync.Mutex
	nextID uint64
}

func (s *sequentialIDGenerator) New() string {
	s.Lock()
	defer s.Unlock()
	ret := fmt.Sprintf("%d", s.nextID)
	s.nextID++
	return ret
}
