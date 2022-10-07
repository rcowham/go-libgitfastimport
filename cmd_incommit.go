// Copyright (C) 2017-2018, 2021  Luke Shumaker <lukeshu@lukeshu.com>
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
	"strconv"
	"strings"

	"github.com/pkg/errors"
	shlex "github.com/rcowham/go-libgitfastimport/shlex"
)

// M ///////////////////////////////////////////////////////////////////////////

// FileModify appears after a CmdCommit (and before a CmdCommitEnd),
// and causes the CmdCommit to add a new file or change the content of
// an existing file.  The content of the file is specified by giving
// either a mark reference (":<idnum>") or by a full 40-byte SHA-1.
//
// To specify the full content of the file inline, use
// FileModifyInline instead.
type FileModify struct {
	Mode    Mode
	Path    Path
	DataRef string
}

func (o FileModify) fiCmdClass() cmdClass { return cmdClassInCommit }
func (o FileModify) fiCmdWrite(fiw fiWriter) error {
	return fiw.WriteLine("M", o.Mode, o.DataRef, PathEscape(o.Path))
}
func init() { parser_registerCmd("M ", FileModify{}) }
func (FileModify) fiCmdRead(fir fiReader) (cmd Cmd, err error) {
	// NB: This parses both FileModify and FileModifyInline
	line, err := fir.ReadLine()
	if err != nil {
		return nil, err
	}
	str := trimLinePrefix(line, "M ")
	fields := strings.SplitN(str, " ", 3)
	if len(fields) != 3 {
		return nil, errors.Errorf("commit: malformed modify command: %q", line)
	}

	nMode, err := strconv.ParseUint(fields[0], 8, 18)
	if err != nil {
		return nil, err
	}

	ref := fields[1]
	path := PathUnescape(fields[2])

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
			Mode: Mode(nMode),
			Path: path,
			Data: data,
		}, nil
	} else {
		return FileModify{
			Mode:    Mode(nMode),
			Path:    path,
			DataRef: ref,
		}, nil
	}
}

// FileModifyInline appears after a CmdCommit (and before a
// CmdCommitEnd), and causes the CmdCommit to add a new file or change
// the content of an existing file.  The full content of the file are
// specified directly
//
// To instead specify the content with a mark reference (":<idnum>")
// or with a full 40-byte SHA-1, use FileModify instead.
type FileModifyInline struct {
	Mode Mode
	Path Path
	Data string
}

func (o FileModifyInline) fiCmdClass() cmdClass { return cmdClassInCommit }
func (o FileModifyInline) fiCmdWrite(fiw fiWriter) error {
	ez := &ezfiw{fiw: fiw}
	ez.WriteLine("M", o.Mode, "inline", PathEscape(o.Path))
	ez.WriteData(o.Data)
	return ez.err
}
func (FileModifyInline) fiCmdRead(fiReader) (Cmd, error) { panic("not reached") }

// D ///////////////////////////////////////////////////////////////////////////

// FileDelete appears after a CmdCommit (and before a CmdCommitEnd),
// and causes the CmdCommit to recursively remove a file or directory.
type FileDelete struct {
	Path Path
}

func (o FileDelete) fiCmdClass() cmdClass { return cmdClassInCommit }
func (o FileDelete) fiCmdWrite(fiw fiWriter) error {
	return fiw.WriteLine("D", PathEscape(o.Path))
}
func init() { parser_registerCmd("D ", FileDelete{}) }
func (FileDelete) fiCmdRead(fir fiReader) (cmd Cmd, err error) {
	line, err := fir.ReadLine()
	if err != nil {
		return nil, err
	}
	return FileDelete{Path: PathUnescape(trimLinePrefix(line, "D "))}, nil
}

// C ///////////////////////////////////////////////////////////////////////////

// FileCopy appears after a CmdCommit (and before a CmdCommitEnd),
// and causes the CmdCommit to recursively copy an existing file or
// subdirectory to a different location.
type FileCopy struct {
	Src Path
	Dst Path
}

func (o FileCopy) fiCmdClass() cmdClass { return cmdClassInCommit }
func (o FileCopy) fiCmdWrite(fiw fiWriter) error {
	return fiw.WriteLine("C", PathEscape(o.Src), PathEscape(o.Dst))
}
func init() { parser_registerCmd("C ", FileCopy{}) }
func (FileCopy) fiCmdRead(fir fiReader) (cmd Cmd, err error) {
	line, err := fir.ReadLine()
	if err != nil {
		return nil, err
	}
	str := trimLinePrefix(line, "C ")
	fields, err := shlex.NewGitLexerString(str).Split()
	if err != nil {
		return nil, errors.Errorf("filecopy: malformed command: %q", err)
	}
	if len(fields) != 2 {
		return nil, errors.Errorf("filecopy: malformed command: %q", line)
	}
	return FileCopy{Src: PathUnescape(fields[0]), Dst: PathUnescape(fields[1])}, nil
}

// R ///////////////////////////////////////////////////////////////////////////

// FileRename appears after a CmdCommit (and before a CmdCommitEnd),
// and causes the CmdCommit to rename an existing file or subdirectory
// to a different location.
type FileRename struct {
	Src Path
	Dst Path
}

func (o FileRename) fiCmdClass() cmdClass { return cmdClassInCommit }
func (o FileRename) fiCmdWrite(fiw fiWriter) error {
	return fiw.WriteLine("R", PathEscape(o.Src), PathEscape(o.Dst))
}
func init() { parser_registerCmd("R ", FileRename{}) }
func (FileRename) fiCmdRead(fir fiReader) (cmd Cmd, err error) {
	line, err := fir.ReadLine()
	if err != nil {
		return nil, err
	}
	str := trimLinePrefix(line, "R ")
	fields, err := shlex.NewGitLexerString(str).Split()
	if err != nil {
		return nil, errors.Errorf("filerename: malformed command: %q", err)
	}
	if len(fields) != 2 {
		return nil, errors.Errorf("filerename: malformed command: %q", line)
	}
	return FileRename{Src: PathUnescape(fields[0]), Dst: PathUnescape(fields[1])}, nil
}

// deleteall ///////////////////////////////////////////////////////////////////

// FileDeleteAll appears after a CmdCommit (and before a
// CmdCommitEnd), and removes all files and directories from the
// CmdCommit.
type FileDeleteAll struct{}

func (o FileDeleteAll) fiCmdClass() cmdClass { return cmdClassInCommit }
func (o FileDeleteAll) fiCmdWrite(fiw fiWriter) error {
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

// NoteModify appears after a CmdCommit (and before a CmdCommitEnd),
// and causes the CmdCommit to add a new note describing CommitIsh or
// change the content of an existing note describing CommitIsh.  The
// content of the note is specified by giving either a mark reference
// (":<idnum>") or by a full 40-byte SHA-1.
//
// To specify the full content of the note inline, use
// NoteModifyInline instead.
type NoteModify struct {
	CommitIsh string
	DataRef   string
}

func (o NoteModify) fiCmdClass() cmdClass { return cmdClassInCommit }
func (o NoteModify) fiCmdWrite(fiw fiWriter) error {
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
		return nil, errors.Errorf("commit: malformed notemodify command: %q", line)
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

// NoteModifyInline appears after a CmdCommit (and before a
// CmdCommitEnd), and causes the CmdCommit to add a new note
// describing CommitIsh or change the content of an existing note
// describing CommitIsh.  The full content of the note is specified
// directly.
//
// To instead specify the content with a mark reference (":<idnum>")
// or with a full 40-byte SHA-1, use NoteModify instead.
type NoteModifyInline struct {
	CommitIsh string
	Data      string
}

func (o NoteModifyInline) fiCmdClass() cmdClass { return cmdClassInCommit }
func (o NoteModifyInline) fiCmdWrite(fiw fiWriter) error {
	ez := &ezfiw{fiw: fiw}
	ez.WriteLine("N", "inline", o.CommitIsh)
	ez.WriteData(o.Data)
	return ez.err
}
func (NoteModifyInline) fiCmdRead(fiReader) (Cmd, error) { panic("not reached") }
