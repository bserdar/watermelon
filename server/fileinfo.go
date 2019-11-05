package server

import (
	"os"
	"strings"
	"time"
)

// CommonFileInfo is an os.FileInfo
type CommonFileInfo struct {
	FileName    string
	FileSize    int64
	FileMode    os.FileMode
	FileModTime time.Time
	FileIsDir   bool
}

// Name returns the file name
func (c CommonFileInfo) Name() string { return c.FileName }

// Size returns the file size
func (c CommonFileInfo) Size() int64 { return c.FileSize }

// Mode returns the file mode
func (c CommonFileInfo) Mode() os.FileMode { return c.FileMode }

// ModTime returns the file modification time
func (c CommonFileInfo) ModTime() time.Time { return c.FileModTime }

// IsDir returns true if this is a directory
func (c CommonFileInfo) IsDir() bool { return c.FileIsDir }

// Sys returns nil
func (c CommonFileInfo) Sys() interface{} { return nil }

// ParseFileMode parses -rwxrwxrwx string
func ParseFileMode(str string) os.FileMode {
	var ret os.FileMode
	if len(str) < 10 {
		return os.FileMode(0)
	}
	const typestr = "dalTLDpSugct?"
	ix := strings.IndexByte(typestr, str[0])
	if ix != -1 {
		ret |= os.FileMode(1 << uint32(32-1-ix))
	}
	rwx := func(input string) os.FileMode {
		var out os.FileMode
		if len(input) >= 3 {
			for i := 0; i < 3; i++ {
				if input[i] == 'r' {
					out |= 4
				} else if input[i] == 'w' {
					out |= 2
				} else if input[i] == 'x' {
					out |= 1
				}
			}
		}
		return out
	}
	return ret | (rwx(str[1:]) << 6) | (rwx(str[4:]) << 3) | (rwx(str[7:]))
}
