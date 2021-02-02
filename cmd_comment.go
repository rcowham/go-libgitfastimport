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
)

// This file deals with comments, and with commands for which the
// specification says (or said) that the command "can be used anywhere
// in the stream that comments are accepted".

// comment /////////////////////////////////////////////////////////////////////

// CmdComment is a comment line; not a real command.
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

// CmdGetMark requests that the Backend to report back (over the
// auxiliary cat-blob stream) with the SHA-1 corresponding to the
// given Mark.
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
		return nil, errors.Wrap(err, "get-mark")
	}
	return c, nil
}

// cat-blob ////////////////////////////////////////////////////////////////////

// CmdCatBlob requests that the Backend to report back (over the
// auxiliary cat-blob stream) with the SHA-1 and content of the
// requested blob.  The blob can be specified either by a mark
// reference (":<idnum>") or by a full 40-byte SHA-1.
type CmdCatBlob struct {
	DataRef string
}

func (c CmdCatBlob) fiCmdClass() cmdClass {
	// Prior to git v2.22.0 this was 'cmdClassComment', but in
	// v2.22.0 it was changed to a stricter
	// 'cmdClassCommand|cmdClassInCommit|cmdClassInFileModify'.  I
	// don't want to implement cmdClassInFileModify for just this
	// one command, and also I want to have better backward
	// compatibility; so I'm keeping it as 'cmdClassComment'.
	return cmdClassComment
}
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

// CmdLs requests that the Backend to report back (over the auxiliary
// cat-blob stream) with information about the object at a path in the
// specified commit.  If inside of a commit, specifying the commit is
// optional, and the ongoing commit is used.  The commit can be
// specified either by a mark reference (":<idnum>") or by a full
// 40-byte SHA-1.
type CmdLs struct {
	DataRef string // optional if inside of a commit
	Path    Path
}

func (c CmdLs) fiCmdClass() cmdClass {
	// Prior to git v2.22.0 the docs said 'ls' was allowed
	// anywhere a comment was allowed, but that was never really
	// true, and in v2.22.0 the docs were updated to match the
	// code.
	if c.DataRef == "" {
		// Yeah, this will give slightly misleading info to
		// parser_registerCmd(), but that's OK,
		// parser_registerCmd() only really cares about the
		// cmdClassInCommand bit.
		return cmdClassInCommit
	}
	return cmdClassCommand | cmdClassInCommit
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
