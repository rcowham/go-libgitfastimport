package libfastimport

import (
	"fmt"
	"strconv"
	"strings"

	"git.lukeshu.com/go/libfastimport/textproto"
)

// M ///////////////////////////////////////////////////////////////////////////

type FileModify struct {
	Mode    textproto.Mode
	Path    textproto.Path
	DataRef string
}

func (o FileModify) fiCmdClass() cmdClass { return cmdClassCommit }
func (o FileModify) fiCmdWrite(fiw *textproto.FIWriter) error {
	return fiw.WriteLine("M", o.Mode, o.DataRef, o.Path)
}
func init() { parser_registerCmd("M ", FileModify{}) }
func (FileModify) fiCmdRead(fir fiReader) (cmd Cmd, err error) {
	// NB: This parses both FileModify and FileModifyInline
	line, err := fir.ReadLine()
	if err != nil {
		return nil, err
	}
	str := trimLinePrefix(line, "M ")
	sp1 := strings.IndexByte(str, ' ')
	sp2 := strings.IndexByte(str[sp1+1:], ' ')
	if sp1 < 0 || sp2 < 0 {
		return nil, fmt.Errorf("commit: malformed modify command: %v", line)
	}
	nMode, err := strconv.ParseUint(str[:sp1], 8, 18)
	if err != nil {
		return nil, err
	}
	ref := str[sp1+1 : sp2]
	path := textproto.PathUnescape(str[sp2+1:])
	if ref == "inline" {
		line, err = fir.ReadLine()
		if err != nil {
			return nil, err
		}
		data, err := parse_data(line)
		if err != nil {
			return nil, err
		}
		return FileModifyInline{
			Mode: textproto.Mode(nMode),
			Path: path,
			Data: data,
		}, nil
	} else {
		return FileModify{
			Mode:    textproto.Mode(nMode),
			Path:    path,
			DataRef: ref,
		}, nil
	}
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
func (FileModifyInline) fiCmdRead(fiReader) (Cmd, error) { panic("not reached") }

// D ///////////////////////////////////////////////////////////////////////////

type FileDelete struct {
	Path textproto.Path
}

func (o FileDelete) fiCmdClass() cmdClass { return cmdClassCommit }
func (o FileDelete) fiCmdWrite(fiw *textproto.FIWriter) error {
	return fiw.WriteLine("D", o.Path)
}
func init() { parser_registerCmd("D ", FileDelete{}) }
func (FileDelete) fiCmdRead(fir fiReader) (cmd Cmd, err error) {
	line, err := fir.ReadLine()
	if err != nil {
		return nil, err
	}
	return FileDelete{Path: textproto.PathUnescape(trimLinePrefix(line, "D "))}, nil
}

// C ///////////////////////////////////////////////////////////////////////////

type FileCopy struct {
	Src textproto.Path
	Dst textproto.Path
}

func (o FileCopy) fiCmdClass() cmdClass { return cmdClassCommit }
func (o FileCopy) fiCmdWrite(fiw *textproto.FIWriter) error {
	return fiw.WriteLine("C", o.Src, o.Dst)
}
func init() { parser_registerCmd("C ", FileDelete{}) }
func (FileCopy) fiCmdRead(fir fiReader) (cmd Cmd, err error) {
	// BUG(lukeshu): TODO: commit C not implemented
	panic("TODO: commit C not implemented")
}

// R ///////////////////////////////////////////////////////////////////////////

type FileRename struct {
	Src string
	Dst string
}

func (o FileRename) fiCmdClass() cmdClass { return cmdClassCommit }
func (o FileRename) fiCmdWrite(fiw *textproto.FIWriter) error {
	return fiw.WriteLine("R", o.Src, o.Dst)
}
func init() { parser_registerCmd("R ", FileDelete{}) }
func (FileRename) fiCmdRead(fir fiReader) (cmd Cmd, err error) {
	// BUG(lukeshu): TODO: commit R not implemented
	panic("TODO: commit R not implemented")
}

// deleteall ///////////////////////////////////////////////////////////////////

type FileDeleteAll struct{}

func (o FileDeleteAll) fiCmdClass() cmdClass { return cmdClassCommit }
func (o FileDeleteAll) fiCmdWrite(fiw *textproto.FIWriter) error {
	return fiw.WriteLine("deleteall")
}
func init() { parser_registerCmd("deleteall\n", FileDeleteAll{}) }
func (FileDeleteAll) fiCmdRead(fir fiReader) (cmd Cmd, err error) {
	_, err = fir.ReadLine()
	if err != nil {
		return nil, err
	}
	return FileDeleteAll{}, nil
}

// N ///////////////////////////////////////////////////////////////////////////

type NoteModify struct {
	CommitIsh string
	DataRef   string
}

func (o NoteModify) fiCmdClass() cmdClass { return cmdClassCommit }
func (o NoteModify) fiCmdWrite(fiw *textproto.FIWriter) error {
	return fiw.WriteLine("N", o.DataRef, o.CommitIsh)
}
func init() { parser_registerCmd("N ", NoteModify{}) }
func (NoteModify) fiCmdRead(fir fiReader) (cmd Cmd, err error) {
	// NB: This parses both NoteModify and NoteModifyInline
	line, err := fir.ReadLine()
	if err != nil {
		return nil, err
	}
	str := trimLinePrefix(line, "N ")
	sp := strings.IndexByte(str, ' ')
	if sp < 0 {
		return nil, fmt.Errorf("commit: malformed notemodify command: %v", line)
	}
	ref := str[:sp]
	commitish := str[sp+1:]
	if ref == "inline" {
		line, err = fir.ReadLine()
		if err != nil {
			return nil, err
		}
		data, err := parse_data(line)
		if err != nil {
			return nil, err
		}
		return NoteModifyInline{
			CommitIsh: commitish,
			Data:      data,
		}, nil
	} else {
		return NoteModify{
			CommitIsh: commitish,
			DataRef:   ref,
		}, nil
	}
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
func (NoteModifyInline) fiCmdRead(fiReader) (Cmd, error) { panic("not reached") }
