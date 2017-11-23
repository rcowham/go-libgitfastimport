package textproto

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// BUG(lukeshu): Only supports the "raw" date format (not "rfc2822" or
// "now")
type Ident struct {
	Name  string
	Email string
	Time  time.Time
}

func (ut Ident) String() string {
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

func ParseIdent(str string) (Ident, error) {
	ret := Ident{}
	lt := strings.IndexAny(str, "<>")
	if lt < 0 || str[lt] != '<' {
		return ret, fmt.Errorf("Missing < in ident string: %v", str)
	}
	if lt > 0 {
		if str[lt-1] != ' ' {
			return ret, fmt.Errorf("Missing space before < in ident string: %v", str)
		}
		ret.Name = str[:lt-1]
	}
	gt := lt + 1 + strings.IndexAny(str[lt+1:], "<>")
	if gt < lt+1 || str[gt] != '>' {
		return ret, fmt.Errorf("Missing > in ident string: %v", str)
	}
	if str[gt+1] != ' ' {
		return ret, fmt.Errorf("Missing space after > in ident string: %v", str)
	}
	ret.Email = str[lt+1 : gt]

	strWhen := str[gt+2:]
	sp := strings.IndexByte(strWhen, ' ')
	if sp < 0 {
		return ret, fmt.Errorf("missing time zone in when: %v", str)
	}
	sec, err := strconv.ParseInt(strWhen[:sp], 10, 64)
	if err != nil {
		return ret, err
	}
	tzt, err := time.Parse("-0700", strWhen[sp+1:])
	if err != nil {
		return ret, err
	}
	ret.Time = time.Unix(sec, 0).In(tzt.Location())

	return ret, nil
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

func (m Mode) GoString() string {
	return fmt.Sprintf("%07o", m)
}

func PathEscape(path Path) string {
	if strings.HasPrefix(string(path), "\"") || strings.ContainsRune(string(path), '\n') {
		return "\"" + strings.Replace(strings.Replace(strings.Replace(string(path), "\\", "\\\\", -1), "\"", "\\\"", -1), "\n", "\\n", -1) + "\""
	} else {
		return string(path)
	}
}

func PathUnescape(epath string) Path {
	if strings.HasPrefix(epath, "\"") {
		return Path(strings.Replace(strings.Replace(strings.Replace(epath[1:len(epath)-1], "\\n", "\n", -1), "\\\"", "\"", -1), "\\\\", "\\", -1))
	} else {
		return Path(epath)
	}
}

type Path string

func (p Path) String() string {
	return PathEscape(p)
}
