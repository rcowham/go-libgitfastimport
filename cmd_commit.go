package libfastimport

import (
	"git.lukeshu.com/go/libfastimport/textproto"
)

type FileModify struct {
	Mode    textproto.Mode
	Path    textproto.Path
	DataRef string
}

func (o FileModify) fiCmdClass() cmdClass { return cmdClassCommit }
func (o FileModify) fiCmdWrite(fiw *textproto.FIWriter) error {
	return fiw.WriteLine("M", o.Mode, o.DataRef, o.Path)
}

type FileModifyInline struct {
	Mode textproto.Mode
	Path textproto.Path
	Data string
}

func (o FileModifyInline) fiCmdClass() cmdClass { return cmdClassCommit }
func (o FileModifyInline) fiCmdWrite(fiw *textproto.FIWriter) error {
	ez := &ezfiw{fiw: fiw}
	ez.WriteLine("M", o.Mode, "inline", o.Path)
	ez.WriteData(o.Data)
	return ez.err
}

type FileDelete struct {
	Path textproto.Path
}

func (o FileDelete) fiCmdClass() cmdClass { return cmdClassCommit }
func (o FileDelete) fiCmdWrite(fiw *textproto.FIWriter) error {
	return fiw.WriteLine("D", o.Path)
}

type FileCopy struct {
	Src textproto.Path
	Dst textproto.Path
}

func (o FileCopy) fiCmdClass() cmdClass { return cmdClassCommit }
func (o FileCopy) fiCmdWrite(fiw *textproto.FIWriter) error {
	return fiw.WriteLine("C", o.Src, o.Dst)
}

type FileRename struct {
	Src string
	Dst string
}

func (o FileRename) fiCmdClass() cmdClass { return cmdClassCommit }
func (o FileRename) fiCmdWrite(fiw *textproto.FIWriter) error {
	return fiw.WriteLine("R", o.Src, o.Dst)
}

type FileDeleteAll struct{}

func (o FileDeleteAll) fiCmdClass() cmdClass { return cmdClassCommit }
func (o FileDeleteAll) fiCmdWrite(fiw *textproto.FIWriter) error {
	return fiw.WriteLine("deleteall")
}

type NoteModify struct {
	CommitIsh string
	DataRef   string
}

func (o NoteModify) fiCmdClass() cmdClass { return cmdClassCommit }
func (o NoteModify) fiCmdWrite(fiw *textproto.FIWriter) error {
	return fiw.WriteLine("N", o.DataRef, o.CommitIsh)
}

type NoteModifyInline struct {
	CommitIsh string
	Data      string
}

func (o NoteModifyInline) fiCmdClass() cmdClass { return cmdClassCommit }
func (o NoteModifyInline) fiCmdWrite(fiw *textproto.FIWriter) error {
	ez := &ezfiw{fiw: fiw}
	ez.WriteLine("N", "inline", o.CommitIsh)
	ez.WriteData(o.Data)
	return ez.err
}
