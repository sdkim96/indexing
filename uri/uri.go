// Copyright 2026 Sungdong Kim
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package uri

import (
	"fmt"
	"strings"
)

// URI identifies the source of an resource.
// URI must follow the format of {scheme}://{path},
// For Example, "file:///path/to/file.txt" or "blob://container/blobname".
// The Provider or Writer will determine how to handle the URI based on its scheme.
type URI string

// Scheme returns the scheme part of the URI, which is the substring before "://".
func (u URI) Scheme() string {
	parts := strings.SplitN(string(u), "://", 2)
	if len(parts) != 2 {
		return ""
	}
	return parts[0]
}

// Path returns the path part of the URI, which is the substring after "://".
func (u URI) Path() string {
	parts := strings.SplitN(string(u), "://", 2)
	if len(parts) != 2 {
		return ""
	}
	return parts[1]
}

// Validate checks if the URI is valid, which means it has both a non-empty scheme and path.
func (u URI) Validate() error {
	if u.Scheme() == "" || u.Path() == "" {
		return fmt.Errorf("invalid URI: %q, must follow {scheme}://{path}", string(u))
	}
	return nil
}
