package libfastimport

type fiReader interface {
	PeekLine() (string, error)
	ReadLine() (string, error)
}

type fiWriter interface {
	WriteData(string) error
	WriteLine(a ...interface{}) error
}

type cmdClass int

const (
	cmdClassCommand cmdClass = 1 // may be a top-level command
	cmdClassCommit  cmdClass = 2 // may be used within in a commit

	cmdClassComment cmdClass = cmdClassCommand | cmdClassCommit
)

type Cmd interface {
	fiCmdRead(fiReader) (Cmd, error)
	fiCmdWrite(fiWriter) error
	fiCmdClass() cmdClass
}
