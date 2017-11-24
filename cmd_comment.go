package libfastimport

import (
	"fmt"
	"strconv"
	"strings"
)

// comment /////////////////////////////////////////////////////////////////////

type CmdComment struct {
	Comment string
}

func (c CmdComment) fiCmdClass() cmdClass { return cmdClassComment }
func (c CmdComment) fiCmdWrite(fiw fiWriter) error {
	return fiw.WriteLine("#" + c.Comment)
}
func init() { parser_registerCmd("#", CmdComment{}) }
func (CmdComment) fiCmdRead(fir fiReader) (cmd Cmd, err error) {
	line, err := fir.ReadLine()
	if err != nil {
		return nil, err
	}
	return CmdComment{Comment: trimLinePrefix(line, "#")}, nil
}

// get-mark ////////////////////////////////////////////////////////////////////

type CmdGetMark struct {
	Mark int
}

func (c CmdGetMark) fiCmdClass() cmdClass { return cmdClassComment }
func (c CmdGetMark) fiCmdWrite(fiw fiWriter) error {
	return fiw.WriteLine("get-mark", ":"+strconv.Itoa(c.Mark))
}
func init() { parser_registerCmd("get-mark :", CmdGetMark{}) }
func (CmdGetMark) fiCmdRead(fir fiReader) (cmd Cmd, err error) {
	line, err := fir.ReadLine()
	if err != nil {
		return nil, err
	}
	c := CmdGetMark{}
	c.Mark, err = strconv.Atoi(trimLinePrefix(line, "get-mark :"))
	if err != nil {
		return nil, fmt.Errorf("get-mark: %v", err)
	}
	return c, nil
}

// cat-blob ////////////////////////////////////////////////////////////////////

type CmdCatBlob struct {
	DataRef string
}

func (c CmdCatBlob) fiCmdClass() cmdClass { return cmdClassComment }
func (c CmdCatBlob) fiCmdWrite(fiw fiWriter) error {
	return fiw.WriteLine("cat-blob", c.DataRef)
}
func init() { parser_registerCmd("cat-blob ", CmdCatBlob{}) }
func (CmdCatBlob) fiCmdRead(fir fiReader) (cmd Cmd, err error) {
	line, err := fir.ReadLine()
	if err != nil {
		return nil, err
	}
	return CmdCatBlob{DataRef: trimLinePrefix(line, "cat-blob ")}, nil
}

// ls //////////////////////////////////////////////////////////////////////////

type CmdLs struct {
	DataRef string // optional if inside of a commit
	Path    Path
}

// If you're thinking "but wait, parser_registerCmd will see CmdLs as
// cmdClassCommit, not cmdClassComment, that means it won't be allowed
// embedded inside other commands! (while still allowing it both
// inside and outside of a commit)", you're absolutely correct.
// That's the desired behavior.  It's a happy accident that the little
// fiCmdClass hack works out that way, instead of having to add even
// more complexity.

func (c CmdLs) fiCmdClass() cmdClass {
	if c.DataRef == "" {
		return cmdClassCommit
	}
	return cmdClassComment
}
func (c CmdLs) fiCmdWrite(fiw fiWriter) error {
	if c.DataRef == "" {
		return fiw.WriteLine("ls", c.Path)
	} else {
		return fiw.WriteLine("ls", c.DataRef, c.Path)
	}
}
func init() { parser_registerCmd("ls ", CmdLs{}) }
func (CmdLs) fiCmdRead(fir fiReader) (cmd Cmd, err error) {
	// 'ls' SP <dataref> SP <path> LF
	line, err := fir.ReadLine()
	if err != nil {
		return nil, err
	}
	str := trimLinePrefix(line, "ls ")
	sp := -1
	if !strings.HasPrefix(str, "\"") {
		sp = strings.IndexByte(line, ' ')
	}
	c := CmdLs{}
	c.Path = PathUnescape(str[sp+1:])
	if sp >= 0 {
		c.DataRef = str[:sp]
	}
	return c, nil
}
