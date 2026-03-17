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

package input

import (
	"context"
	"errors"

	"github.com/sdkim96/indexing/uri"
)

// Provider reads the data from the source identified by the URI
// and returns an Input that can be used to read the data.
type Provider interface {

	// Provide reads the data from the source identified by the URI
	Provide(ctx context.Context, URI uri.URI) (Input, error)
}

var ErrUnsupportedSourceScheme error = errors.New("unsupported source key. check the scheme.")
var ErrInvalidSourceKey error = errors.New("invalid source key. must follow {scheme}://{path}")
