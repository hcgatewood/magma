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
	"testing"

	blobstore_test "magma/orc8r/cloud/go/blobstore/test"
	"magma/orc8r/cloud/go/orc8r"
	"magma/orc8r/cloud/go/service/test"
	"magma/orc8r/cloud/go/services/accessd"
	"magma/orc8r/cloud/go/services/accessd/protos"
	"magma/orc8r/cloud/go/services/accessd/servicers"
	"magma/orc8r/cloud/go/services/accessd/storage"
)

func StartTestService(t *testing.T) {
	srv, lis := test.NewService(t, orc8r.ModuleName, accessd.ServiceName)
	store := blobstore_test.NewSQLBlobstore(t, storage.AccessdTableBlobstore)
	accessdStore := storage.NewAccessdBlobstore(store)
	protos.RegisterAccessControlManagerServer(srv.GrpcServer, servicers.NewAccessdServer(accessdStore))
	go srv.MustRunTest(t, lis)
}
