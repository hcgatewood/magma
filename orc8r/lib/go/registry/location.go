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

package registry

import "fmt"

// ServiceLocation is an entry for the service registry which identifies a
// service by name and the host:port that it is running on.
type ServiceLocation struct {
	// Name of the service.
	Name string
	// Host name of the service.
	Host string
	// Port is the service's gRPC endpoint.
	Port int
	// EchoPort is the service's HTTP endpoint for providing obsidian handlers.
	EchoPort int
	// ProxyAliases provides the list of host:port aliases for the service.
	ProxyAliases map[string]int

	// Labels provide a way to identify the service.
	// Use cases include listing service mesh servicers the service implements.
	Labels map[string]string
	// Annotations provides a string-to-string map of per-service metadata.
	Annotations map[string]string
}

func (s ServiceLocation) HasLabel(label string) bool {
	_, ok := s.Labels[label]
	return ok
}

// String implements ServiceLocation stringer interface
// Returns string in the form: <service name> @ host:port (also known as: host:port, ...)
func (s ServiceLocation) String() string {
	alsoKnown := ""
	if len(s.ProxyAliases) > 0 {
		aliases := ""
		for host, port := range s.ProxyAliases {
			aliases += fmt.Sprintf(" %s:%d,", host, port)
		}
		alsoKnown = " (also known as:" + aliases[:len(aliases)-1] + ")"
	}
	return fmt.Sprintf("%s @ %s:%d%s", s.Name, s.Host, s.Port, alsoKnown)
}
