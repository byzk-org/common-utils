package commonutils

import (
	"path"
	"runtime"
	"strings"
)

func PathJoin(ele ...string) string {
	str := path.Join(ele...)
	if runtime.GOOS == "windows" {
		str = strings.ReplaceAll(str, "/", "\\")
		str = strings.ReplaceAll(str, "\\\\", "\\")
	}
	return str
}
