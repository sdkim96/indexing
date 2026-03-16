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

func (u URI) Scheme() string {
	parts := strings.SplitN(string(u), "://", 2)
	if len(parts) != 2 {
		return ""
	}
	return parts[0]
}

func (u URI) Path() string {
	parts := strings.SplitN(string(u), "://", 2)
	if len(parts) != 2 {
		return ""
	}
	return parts[1]
}

func (u URI) Validate() error {
	if u.Scheme() == "" || u.Path() == "" {
		return fmt.Errorf("invalid URI: %q, must follow {scheme}://{path}", string(u))
	}
	return nil
}
