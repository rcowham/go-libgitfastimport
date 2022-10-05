// Copyright (C) 2017-2018, 2020-2021  Luke Shumaker <lukeshu@lukeshu.com>
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

package libfastimport

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"
)

// Ident is a tuple of a commiter's (or author's) name, email, and a
// timestamp with timezone.
//
// BUG(lukeshu): Ident (and ParseIdent) only supports the
// "raw"/"raw-permissive" date format (not "rfc2822" or "now")
type Ident struct {
	Name  string
	Email string
	Time  time.Time
}

func (ut Ident) String() string {
	if ut.Name == "" {
		return fmt.Sprintf("<%s> %d %s",
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

// ParseIdent parses a string containing an Ident.
//
// The format of this string is
//
//     <name> SP LT <email> GT SP <time> SP <offutc>
//
// Where <name> may contain a space, but not "<" or ">"; <time> is an
// integer number of seconds since the UNIX epoch (UTC); <offutc> is
// positive or negative 4-digit offset from UTC (for example, EST
// would be "-0500").
func ParseIdent(str string) (Ident, error) {
	ret := Ident{}
	lt := strings.IndexAny(str, "<>")
	if lt < 0 || str[lt] != '<' {
		return ret, errors.Errorf("Missing < in ident string: %q", str)
	}
	if lt > 0 {
		if str[lt-1] != ' ' {
			return ret, errors.Errorf("Missing space before < in ident string: %q", str)
		}
		ret.Name = str[:lt-1]
	}
	gt := lt + 1 + strings.IndexAny(str[lt+1:], "<>")
	if gt < lt+1 || str[gt] != '>' {
		return ret, errors.Errorf("Missing > in ident string: %q", str)
	}
	if str[gt+1] != ' ' {
		return ret, errors.Errorf("Missing space after > in ident string: %q", str)
	}
	ret.Email = str[lt+1 : gt]

	strWhen := str[gt+2:]
	sp := strings.IndexByte(strWhen, ' ')
	if sp < 0 {
		return ret, errors.Errorf("missing time zone in when: %q", str)
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

// Mode is a file mode as seen by git.
type Mode uint32 // 18 bits

const (
	ModeFil = Mode(0100644) // A regular file
	ModeExe = Mode(0100755) // An executable file
	ModeSym = Mode(0120000) // A symbolic link
	ModeGit = Mode(0160000) // A nested git repository (e.g. submodule)
	ModeDir = Mode(0040000) // A directory
)

func (m Mode) String() string {
	return fmt.Sprintf("%06o", m)
}

func (m Mode) GoString() string {
	return fmt.Sprintf("%07o", m)
}

// Path is a string storing a git path.
type Path string

// PathEscape escapes a path in case it contains special characters.
func PathEscape(path Path) string {
	if strings.HasPrefix(string(path), "\"") || strings.ContainsRune(string(path), ' ') || strings.ContainsRune(string(path), '\n') {
		return "\"" + strings.Replace(strings.Replace(strings.Replace(string(path), "\\", "\\\\", -1), "\"", "\\\"", -1), "\n", "\\n", -1) + "\""
	} else {
		return string(path)
	}
}

// PathUnescape unescapes a quoted path.
func PathUnescape(epath string) Path {
	if strings.HasPrefix(epath, "\"") {
		return Path(strings.Replace(strings.Replace(strings.Replace(epath[1:len(epath)-1], "\\n", "\n", -1), "\\\"", "\"", -1), "\\\\", "\\", -1))
	} else {
		return Path(epath)
	}
}

// String calls PathEscape on the Path.
func (p Path) String() string {
	return PathEscape(p)
}
