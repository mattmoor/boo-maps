/*
Copyright 2018 Matt Moore

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package names

import (
	"fmt"

	"github.com/mattmoor/boo-maps/pkg/apis/boos/v1alpha1"
)

// ImmutableMap gives the name of the next snapshot of this map.
func ImmutableMap(i *v1alpha1.MutableMap) string {
	return fmt.Sprintf("%s-%05d", i.Name, i.Generation)
}
