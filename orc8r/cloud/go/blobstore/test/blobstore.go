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
	"fmt"
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"

	"magma/orc8r/cloud/go/blobstore"
	"magma/orc8r/cloud/go/sqorc"
	"magma/orc8r/cloud/go/sqorc/test"
)

// NewSQLBlobstore returns a new blobstore storage factory utilizing the
// singleton in-memory database.
func NewSQLBlobstore(t *testing.T, tableName string) blobstore.BlobStorageFactory {
	if t == nil {
		panic("for tests only")
	}
	fact, err := NewSQLBlobstoreForServices(tableName)
	assert.NoError(t, err)
	return fact
}

// NewSQLBlobstoreForServices is same as NewSQLBlobstore, but for use in
// validation-oriented services.
// Prefer NewSQLBlobstore wherever possible.
func NewSQLBlobstoreForServices(tableName string) (blobstore.BlobStorageFactory, error) {
	db, err := test.GetSharedMemoryDB()
	if err != nil {
		return nil, err
	}

	// Since the backing storage is process-shared, drop the table if it exists
	// to provide a clean slate across test cases
	_, err = db.Exec(fmt.Sprintf("DROP TABLE IF EXISTS %s", tableName))
	if err != nil {
		return nil, errors.Wrapf(err, "drop test SQL blobstore table: %s", tableName)
	}

	store := blobstore.NewSQLBlobStorageFactory(tableName, db, sqorc.GetSqlBuilder())
	err = store.InitializeFactory()
	if err != nil {
		return nil, err
	}

	return store, nil
}
