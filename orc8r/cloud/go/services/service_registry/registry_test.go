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

package service_registry

import (
	"testing"

	"magma/orc8r/lib/go/registry"

	"github.com/stretchr/testify/assert"
)

func TestServiceRegistry_GetAnnotationFields(t *testing.T) {
	tests := []struct {
		name            string
		annotationValue string
		want            []string
	}{
		{
			name:            "empty",
			annotationValue: "",
			want:            nil,
		},
		{
			name:            "all whitespace",
			annotationValue: "  \n\n  ",
			want:            nil,
		},
		{
			name:            "single element",
			annotationValue: "42",
			want:            []string{"42"},
		},
		{
			name:            "multiple elements",
			annotationValue: "42,foo",
			want:            []string{"42", "foo"},
		},
		{
			name:            "multiple elements with whitespace",
			annotationValue: "  42 ,\n  foo  ",
			want:            []string{"42", "foo"},
		},
		{
			name:            "trailing separator",
			annotationValue: "  a,       b, c,\n\nd,    e,\n\n  f,  \n  ",
			want:            []string{"a", "b", "c", "d", "e", "f"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := NewYAMLRegistry()
			location := registry.ServiceLocation{
				Name:        "srv",
				Annotations: map[string]string{"annotationName": tt.annotationValue},
			}
			r.AddServices(location)
			got, err := GetAnnotationListFromRegistry(r, "srv", "annotationName")
			assert.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}
