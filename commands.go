package libfastimport

import (
	"strconv"
)

type ezfiw struct {
	fiw *FIWriter
	err error
}

func (e *ezfiw) WriteLine(a ...interface{}) {
	if e.err == nil {
		e.err = e.fiw.WriteLine(a...)
	}
}

func (e *ezfiw) WriteData(data []byte) {
	if e.err == nil {
		e.err = e.fiw.WriteData(data)
	}
}

func (e *ezfiw) WriteMark(idnum int) {
	if e.err == nil {
		e.err = e.fiw.WriteLine("mark", ":"+strconv.Itoa(idnum))
	}
}

type Cmd interface {
	fiWriteCmd(*FIWriter) error
	fiReadCmd(*FIReader) error
}

type CmdCommit struct {
	Ref       string
	Mark      int // optional; < 1 for non-use
	Author    *UserTime
	Committer UserTime
	Msg       []byte
	Parents   []string
	Tree      []FileAction
}

func (c *CmdCommit) fiWriteCmd(fiw *FIWriter) error {
	ez := &ezfiw{fiw: fiw}

	ez.WriteLine("commit", c.Ref)
	if c.Mark > 0 {
		ez.WriteMark(c.Mark)
	}
	if c.Author != nil {
		ez.WriteLine("author", *c.Author)
	}
	ez.WriteLine("committer", c.Committer)
	ez.WriteData(c.Data)
	if len(c.Parents) > 0 {
		ez.WriteLine("from", c.Parents[0])
		if len(c.Parents) > 1 {
			for _, parent := range c.Parents[1:] {
				ez.WriteLine("merge", parent)
			}
		}
	}

	if ez.err != nil {
		return ez.err
	}

	for _, action := range c.Tree {
		err := action.fiWriteFA(fi)
		if err != nil {
			return err
		}
	}

	return nil
}

type CmdTag struct {
	RefName   string
	CommitIsh string
	Tagger    UserTime
	Data      []byte
}

func (c *CmdTag) fiWriteCmd(fiw *FIWriter) error {
	ez := &ezfiw{fiw: fiw}

	ez.WriteLine("tag", c.RefName)
	ez.WriteLine("from", c.CommitIsh)
	ez.WriteLine("tagger", c.Tagger)
	ez.WriteData(c.Data)

	return ez.err
}

type CmdReset struct {
	RefName   string
	CommitIsh string // optional
}

func (c *CmdReset) fiWriteCmd(fiw *FIWriter) error {
	ez := &ezfiw{fiw: fiw}

	ez.WriteLine("reset", c.RefName)
	if c.CommitIsh != "" {
		ez.WriteLine("from", c.CommitIsh)
	}

	return ez.err
}

type CmdBlob struct {
	Mark int
	Data []byte
}

func (c *CmdBlob) fiWriteCmd(fiw *FIWriter) error {
	ez := &ezfiw{fiw: fiw}

	ez.WriteLine("blob")
	if mark > 0 {
		ez.WriteMark(c.Mark)
	}
	ez.WriteData(c.Data)

	return ez.err
}

type CmdCheckpoint struct{}

func (c *CmdCheckpoint) fiWriteCmd(fiw *FIWriter) error {
	return fiw.WriteLine("checkpoint")
}

type CmdProgress struct {
	Str string
}

func (c *CmdProgress) fiWriteCmd(fiw *FIWriter) error {
	return fiw.WriteLine("progress", c.Str)
}

type CmdGetMark struct {
	Mark int
}

func (c *CmdGetmark) fiWriteCmd(fiw *FIWriter) error {
	return fiw.WriteLine("get-mark", ":"+strconv.Itoa(c.Mark))
}

type CmdCatBlob struct {
	DataRef string
}

func (c *CmdCatBlob) fiWriteCmd(fiw *FIWriter) error {
	return fiw.WriteLine("cat-blob", c.DataRef)
}

// See FileLs for using ls inside of a commit
type CmdLs struct {
	DataRef string
	Path    Path
}

func (c *CmdLs) fiWriteCmd(fiw *FIWriter) error {
	return fiw.WriteLine("ls", c.DataRef, c.Path)
}

type CmdFeature struct {
	Feature  string
	Argument string
}

func (c *CmdFeature) fiWriteCmd(fiw *FIWriter) error {
	if c.Argument != "" {
		return fiw.WriteLine("feature", c.Feature+"="+c.Argument)
	} else {
		return fiw.WriteLine("feature", c.Feature)
	}
}

type CmdOption struct {
	Option string
}

func (c *CmdOption) fiWriteCmd(fiw *FIWriter) error {
	return fiw.WriteLine("option", c.Option)
}

type CmdDone struct{}

func (c *CmdDone) fiWriteCmd(fiw *FIWriter) error {
	return fiw.WriteLine("done")
}

type CmdComment struct {
	Comment string
}

func (c *CmdComment) fiWriteCmd(fiw *FIWriter) error {
	return fiw.WriteLine("#" + c.Comment)
}
