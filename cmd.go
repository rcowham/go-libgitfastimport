package libfastimport

import (
	"git.lukeshu.com/go/libfastimport/textproto"
)

type fiReader interface {
	PeekLine() (string, error)
	ReadLine() (string, error)
}

type cmdClass int

const (
	cmdClassCommand cmdClass = 1 // may be a top-level command
	cmdClassCommit  cmdClass = 2 // may be used within in a commit

	cmdClassComment cmdClass = cmdClassCommand | cmdClassCommit
)

type Cmd interface {
	fiCmdRead(fiReader) (Cmd, error)
	fiCmdWrite(*textproto.FIWriter) error
	fiCmdClass() cmdClass
}
