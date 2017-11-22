package libfastimport

import (
	"git.lukeshu.com/go/libfastimport/textproto"
)

type CmdCommit struct {
	Ref       string
	Mark      int // optional; < 1 for non-use
	Author    *textproto.Ident
	Committer textproto.Ident
	Msg       string
	From      string
	Merge     []string
}

func (c CmdCommit) fiCmdClass() cmdClass { return cmdClassCommand }
func (c CmdCommit) fiCmdWrite(fiw *textproto.FIWriter) error {
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

	return ez.err
}

type CmdTag struct {
	RefName   string
	CommitIsh string
	Tagger    textproto.Ident
	Data      string
}

func (c CmdTag) fiCmdClass() cmdClass { return cmdClassCommand }
func (c CmdTag) fiCmdWrite(fiw *textproto.FIWriter) error {
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

func (c CmdReset) fiCmdClass() cmdClass { return cmdClassCommand }
func (c CmdReset) fiCmdWrite(fiw *textproto.FIWriter) error {
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

func (c CmdBlob) fiCmdClass() cmdClass { return cmdClassCommand }
func (c CmdBlob) fiCmdWrite(fiw *textproto.FIWriter) error {
	ez := &ezfiw{fiw: fiw}

	ez.WriteLine("blob")
	if c.Mark > 0 {
		ez.WriteMark(c.Mark)
	}
	ez.WriteData(c.Data)

	return ez.err
}

type CmdCheckpoint struct{}

func (c CmdCheckpoint) fiCmdClass() cmdClass { return cmdClassCommand }
func (c CmdCheckpoint) fiCmdWrite(fiw *textproto.FIWriter) error {
	return fiw.WriteLine("checkpoint")
}

type CmdProgress struct {
	Str string
}

func (c CmdProgress) fiCmdClass() cmdClass { return cmdClassCommand }
func (c CmdProgress) fiCmdWrite(fiw *textproto.FIWriter) error {
	return fiw.WriteLine("progress", c.Str)
}

type CmdFeature struct {
	Feature  string
	Argument string
}

func (c CmdFeature) fiCmdClass() cmdClass { return cmdClassCommand }
func (c CmdFeature) fiCmdWrite(fiw *textproto.FIWriter) error {
	if c.Argument != "" {
		return fiw.WriteLine("feature", c.Feature+"="+c.Argument)
	} else {
		return fiw.WriteLine("feature", c.Feature)
	}
}

type CmdOption struct {
	Option string
}

func (c CmdOption) fiCmdClass() cmdClass { return cmdClassCommand }
func (c CmdOption) fiCmdWrite(fiw *textproto.FIWriter) error {
	return fiw.WriteLine("option", c.Option)
}

type CmdDone struct{}

func (c CmdDone) fiCmdClass() cmdClass { return cmdClassCommand }
func (c CmdDone) fiCmdWrite(fiw *textproto.FIWriter) error {
	return fiw.WriteLine("done")
}
