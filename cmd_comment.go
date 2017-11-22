package libfastimport

import (
	"strconv"

	"git.lukeshu.com/go/libfastimport/textproto"
)

type CmdComment struct {
	Comment string
}

func (c CmdComment) fiCmdClass() cmdClass { return cmdClassComment }
func (c CmdComment) fiCmdWrite(fiw *textproto.FIWriter) error {
	return fiw.WriteLine("#" + c.Comment)
}

type CmdGetMark struct {
	Mark int
}

func (c CmdGetMark) fiCmdClass() cmdClass { return cmdClassComment }
func (c CmdGetMark) fiCmdWrite(fiw *textproto.FIWriter) error {
	return fiw.WriteLine("get-mark", ":"+strconv.Itoa(c.Mark))
}

type CmdCatBlob struct {
	DataRef string
}

func (c CmdCatBlob) fiCmdClass() cmdClass { return cmdClassComment }
func (c CmdCatBlob) fiCmdWrite(fiw *textproto.FIWriter) error {
	return fiw.WriteLine("cat-blob", c.DataRef)
}

type CmdLs struct {
	DataRef string // optional if inside of a commit
	Path    textproto.Path
}

func (c CmdLs) fiCmdClass() cmdClass {
	if c.DataRef == "" {
		return cmdClassCommit
	}
	return cmdClassComment
}
func (c CmdLs) fiCmdWrite(fiw *textproto.FIWriter) error {
	if c.DataRef == "" {
		return fiw.WriteLine("ls", c.Path)
	} else {
		return fiw.WriteLine("ls", c.DataRef, c.Path)
	}
}

