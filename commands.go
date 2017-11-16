package libfastimport

import (
	"strconv"

	"git.lukeshu.com/go/libfastimport/textproto"
)

type Cmd interface {
	fiWriteCmd(*textproto.FIWriter) error
}

type CmdCommit struct {
	Ref       string
	Mark      int // optional; < 1 for non-use
	Author    *textproto.UserTime
	Committer textproto.UserTime
	Msg       string
	From      string
	Merge     []string
	Tree      []FileAction
}

func (c CmdCommit) fiWriteCmd(fiw *textproto.FIWriter) error {
	ez := &ezfiw{fiw: fiw}

	ez.WriteLine("commit", c.Ref)
	if c.Mark > 0 {
		ez.WriteMark(c.Mark)
	}
	if c.Author != nil {
		ez.WriteLine("author", *c.Author)
	}
	ez.WriteLine("committer", c.Committer)
	ez.WriteData(c.Msg)
	if c.From != "" {
		ez.WriteLine("from", c.From)
	}
	for _, merge := range c.Merge {
		ez.WriteLine("merge", merge)
	}

	if ez.err != nil {
		return ez.err
	}

	for _, action := range c.Tree {
		err := action.fiWriteFA(fiw)
		if err != nil {
			return err
		}
	}

	return nil
}

type CmdTag struct {
	RefName   string
	CommitIsh string
	Tagger    textproto.UserTime
	Data      string
}

func (c CmdTag) fiWriteCmd(fiw *textproto.FIWriter) error {
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

func (c CmdReset) fiWriteCmd(fiw *textproto.FIWriter) error {
	ez := &ezfiw{fiw: fiw}

	ez.WriteLine("reset", c.RefName)
	if c.CommitIsh != "" {
		ez.WriteLine("from", c.CommitIsh)
	}

	return ez.err
}

type CmdBlob struct {
	Mark int
	Data string
}

func (c CmdBlob) fiWriteCmd(fiw *textproto.FIWriter) error {
	ez := &ezfiw{fiw: fiw}

	ez.WriteLine("blob")
	if c.Mark > 0 {
		ez.WriteMark(c.Mark)
	}
	ez.WriteData(c.Data)

	return ez.err
}

type CmdCheckpoint struct{}

func (c CmdCheckpoint) fiWriteCmd(fiw *textproto.FIWriter) error {
	return fiw.WriteLine("checkpoint")
}

type CmdProgress struct {
	Str string
}

func (c CmdProgress) fiWriteCmd(fiw *textproto.FIWriter) error {
	return fiw.WriteLine("progress", c.Str)
}

type CmdGetMark struct {
	Mark int
}

func (c CmdGetMark) fiWriteCmd(fiw *textproto.FIWriter) error {
	return fiw.WriteLine("get-mark", ":"+strconv.Itoa(c.Mark))
}

type CmdCatBlob struct {
	DataRef string
}

func (c CmdCatBlob) fiWriteCmd(fiw *textproto.FIWriter) error {
	return fiw.WriteLine("cat-blob", c.DataRef)
}

// See FileLs for using ls inside of a commit
type CmdLs struct {
	DataRef string
	Path    textproto.Path
}

func (c CmdLs) fiWriteCmd(fiw *textproto.FIWriter) error {
	return fiw.WriteLine("ls", c.DataRef, c.Path)
}

type CmdFeature struct {
	Feature  string
	Argument string
}

func (c CmdFeature) fiWriteCmd(fiw *textproto.FIWriter) error {
	if c.Argument != "" {
		return fiw.WriteLine("feature", c.Feature+"="+c.Argument)
	} else {
		return fiw.WriteLine("feature", c.Feature)
	}
}

type CmdOption struct {
	Option string
}

func (c CmdOption) fiWriteCmd(fiw *textproto.FIWriter) error {
	return fiw.WriteLine("option", c.Option)
}

type CmdDone struct{}

func (c CmdDone) fiWriteCmd(fiw *textproto.FIWriter) error {
	return fiw.WriteLine("done")
}

type CmdComment struct {
	Comment string
}

func (c CmdComment) fiWriteCmd(fiw *textproto.FIWriter) error {
	return fiw.WriteLine("#" + c.Comment)
}
