package libfastimport

import (
	"fmt"
	"strings"
	"time"
)

type UserTime struct {
	Name  string
	Email string
	Time  time.Time
}

func (ut UserTime) String() string {
	if ut.Name == "" {
		return fmt.Sprintf("<%s> %d %s",
			ut.Name,
			ut.Email,
			ut.Time.Unix(),
			ut.Time.Format("-0700"))
	} else {
		return fmt.Sprintf("%s <%s> %d %s",
			ut.Name,
			ut.Email,
			ut.Time.Unix(),
			ut.Time.Format("-0700"))
	}
}

type Mode uint32 // 18 bits

var (
	ModeFil = Mode(0100644)
	ModeExe = Mode(0100755)
	ModeSym = Mode(0120000)
	ModeGit = Mode(0160000)
	ModeDir = Mode(0040000)
)

func (m Mode) String() string {
	return fmt.Sprintf("%06o", m)
}

func PathEscape(path string) string {
	if strings.HasPrefix(path, "\"") || strings.ContainsRune("\n") {
		return "\"" + strings.Replace(strings.Replace(strings.Replace(path, "\\", "\\\\", -1), "\"", "\\\"", -1), "\n", "\\n", -1) + "\""
	} else {
		return path
	}
}

func PathUnescape(epath string) string {
	if strings.HasPrefix(epath, "\"") {
		return strings.Replace(strings.Replace(strings.Replace(epath[1:len(epath)-1], "\\n", "\n", -1), "\\\"", "\"", -1), "\\\\", "\\", -1)
	} else {
		return epath
	}
}

type Path string

func (p Path) String() string {
	return PathEscape(string(p))
}
