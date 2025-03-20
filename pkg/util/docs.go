package util

import (
	"fmt"
)

// DocsURL returns a full documentation URL for the current version of Shoutrrr with the path appended.
// If the path contains a leading slash, it is stripped.
func DocsURL(path string) string {
	// strip leading slash if present
	if len(path) > 0 && path[0] == '/' {
		path = path[1:]
	}
	return fmt.Sprintf("https://github.com/serverleader/shoutrrr/blob/main/docs/%s", path)
}
