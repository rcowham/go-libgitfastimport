// Copyright (C) 2017-2018  Luke Shumaker <lukeshu@lukeshu.com>
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
	"fmt"
	"strconv"
	"strings"

	"git.lukeshu.com/go/libfastimport/textproto"
)

// commit //////////////////////////////////////////////////////////////////////

// CmdCommit requests that the Backend creates or updates a branch
// with a new commit.
//
// This command may be followed by zero or more "File" or "Note"
// commands to set the content of commit's tree.  When reading from a
// Frontend, that sequence of "File" and "Note" commands will be
// terminated by a CmdCommitEnd command.  It is not nescessary to
// manually emit a CmdCommitEnd when writing to a Backend.
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
func (c CmdCommit) fiCmdWrite(fiw fiWriter) error {
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
func init() { parser_registerCmd("commit ", CmdCommit{}) }
func (CmdCommit) fiCmdRead(fir fiReader) (cmd Cmd, err error) {
	ez := &ezfir{fir: fir}
	defer func() { err = ez.Defer() }()

	// 'commit' SP <ref> LF
	c := CmdCommit{Ref: trimLinePrefix(ez.ReadLine(), "commit ")}

	// mark?
	if strings.HasPrefix(ez.PeekLine(), "mark :") {
		c.Mark, err = strconv.Atoi(trimLinePrefix(ez.ReadLine(), "mark :"))
		ez.Errcheck(err)
	}

	// ('author' (SP <name>)? SP LT <email> GT SP <when> LF)?
	if strings.HasPrefix(ez.PeekLine(), "author ") {
		author, err := textproto.ParseIdent(trimLinePrefix(ez.ReadLine(), "author "))
		ez.Errcheck(err)
		c.Author = &author
	}

	// 'committer' (SP <name>)? SP LT <email> GT SP <when> LF
	if !strings.HasPrefix(ez.PeekLine(), "committer ") {
		ez.Errcheck(fmt.Errorf("commit: expected committer command: %v", ez.ReadLine()))
	}
	c.Committer, err = textproto.ParseIdent(trimLinePrefix(ez.ReadLine(), "committer "))
	ez.Errcheck(err)

	// data
	c.Msg, err = parse_data(ez.ReadLine())
	ez.Errcheck(err)

	// ('from' SP <commit-ish> LF)?
	if strings.HasPrefix(ez.PeekLine(), "from ") {
		c.From = trimLinePrefix(ez.ReadLine(), "from ")
	}

	// ('merge' SP <commit-ish> LF)*
	for strings.HasPrefix(ez.PeekLine(), "merge ") {
		c.Merge = append(c.Merge, trimLinePrefix(ez.ReadLine(), "merge "))
	}

	cmd = c
	return
}

// CmdCommitEnd indicates the Frontend will be sending no more "File"
// or "Note" commands that are "part of" the current CmdCommit.
//
// This is a synthesized command to simplify reading from a Frontend;
// it is not really a command in the stream.  It is thus not
// nescessary to send a CmdCommitEnd command when writing to a
// Backend.
type CmdCommitEnd struct{}

func (CmdCommitEnd) fiCmdClass() cmdClass                { return cmdClassCommit }
func (CmdCommitEnd) fiCmdWrite(fiw fiWriter) error       { return nil }
func (CmdCommitEnd) fiCmdRead(fir fiReader) (Cmd, error) { panic("not reached") }

// tag /////////////////////////////////////////////////////////////////////////

// CmdTag requests that the Backend creates an *annotated* tag
// referencing a specific commit.
//
// Hint: Use CmdReset to create a *lightweight* tag.
type CmdTag struct {
	RefName   string
	CommitIsh string
	Tagger    textproto.Ident
	Data      string
}

func (c CmdTag) fiCmdClass() cmdClass { return cmdClassCommand }
func (c CmdTag) fiCmdWrite(fiw fiWriter) error {
	ez := &ezfiw{fiw: fiw}

	ez.WriteLine("tag", c.RefName)
	ez.WriteLine("from", c.CommitIsh)
	ez.WriteLine("tagger", c.Tagger)
	ez.WriteData(c.Data)

	return ez.err
}
func init() { parser_registerCmd("tag ", CmdTag{}) }
func (CmdTag) fiCmdRead(fir fiReader) (cmd Cmd, err error) {
	ez := &ezfir{fir: fir}
	defer func() { err = ez.Defer() }()

	// 'tag' SP <name> LF
	c := CmdTag{RefName: trimLinePrefix(ez.ReadLine(), "tag ")}

	// 'from' SP <commit-ish> LF
	if !strings.HasPrefix(ez.PeekLine(), "from ") {
		ez.Errcheck(fmt.Errorf("tag: expected from command: %v", ez.ReadLine()))
	}
	c.CommitIsh = trimLinePrefix(ez.ReadLine(), "from ")

	// 'tagger' (SP <name>)? SP LT <email> GT SP <when> LF
	if !strings.HasPrefix(ez.PeekLine(), "tagger ") {
		ez.Errcheck(fmt.Errorf("tag: expected tagger command: %v", ez.ReadLine()))
	}
	c.Tagger, err = textproto.ParseIdent(trimLinePrefix(ez.ReadLine(), "tagger "))
	ez.Errcheck(err)

	// data
	c.Data, err = parse_data(ez.ReadLine())
	ez.Errcheck(err)

	cmd = c
	return
}

// reset ///////////////////////////////////////////////////////////////////////

// CmdReset requests that the Backend creates (or recreates) the named
// ref (usually a branch), optionally starting from a specific
// revision.
type CmdReset struct {
	RefName   string
	CommitIsh string // optional
}

func (c CmdReset) fiCmdClass() cmdClass { return cmdClassCommand }
func (c CmdReset) fiCmdWrite(fiw fiWriter) error {
	ez := &ezfiw{fiw: fiw}

	ez.WriteLine("reset", c.RefName)
	if c.CommitIsh != "" {
		ez.WriteLine("from", c.CommitIsh)
	}

	return ez.err
}
func init() { parser_registerCmd("reset ", CmdReset{}) }
func (CmdReset) fiCmdRead(fir fiReader) (cmd Cmd, err error) {
	ez := &ezfir{fir: fir}
	defer func() { err = ez.Defer() }()

	// 'reset' SP <ref> LF
	c := CmdReset{RefName: trimLinePrefix(ez.ReadLine(), "reset ")}

	// ('from' SP <commit-ish> LF)?
	if strings.HasPrefix(ez.PeekLine(), "from ") {
		c.CommitIsh = trimLinePrefix(ez.ReadLine(), "from ")
	}

	cmd = c
	return
}

// blob ////////////////////////////////////////////////////////////////////////

// CmdBlob requests that the Backend write file revision.  The blob
// can be later referred to by the specified Mark (if a Mark > 0 is
// given), or by pre-calculating the Git SHA-1 (though this is
// needlessly difficult, just specify a Mark).
type CmdBlob struct {
	Mark int // optional
	Data string
}

func (c CmdBlob) fiCmdClass() cmdClass { return cmdClassCommand }
func (c CmdBlob) fiCmdWrite(fiw fiWriter) error {
	ez := &ezfiw{fiw: fiw}

	ez.WriteLine("blob")
	if c.Mark > 0 {
		ez.WriteMark(c.Mark)
	}
	ez.WriteData(c.Data)

	return ez.err
}
func init() { parser_registerCmd("blob\n", CmdBlob{}) }
func (CmdBlob) fiCmdRead(fir fiReader) (cmd Cmd, err error) {
	ez := &ezfir{fir: fir}
	defer func() { err = ez.Defer() }()

	// 'blob' LF
	_ = ez.ReadLine()
	c := CmdBlob{}

	// mark?
	if strings.HasPrefix(ez.PeekLine(), "mark :") {
		c.Mark, err = strconv.Atoi(trimLinePrefix(ez.ReadLine(), "mark :"))
	}

	// data
	c.Data, err = parse_data(ez.ReadLine())
	ez.Errcheck(err)

	cmd = c
	return
}

// checkpoint //////////////////////////////////////////////////////////////////

// CmdCheckpoint requests that the Backend flush already-sent data.
type CmdCheckpoint struct{}

func (c CmdCheckpoint) fiCmdClass() cmdClass { return cmdClassCommand }
func (c CmdCheckpoint) fiCmdWrite(fiw fiWriter) error {
	return fiw.WriteLine("checkpoint")
}
func init() { parser_registerCmd("checkpoint\n", CmdCheckpoint{}) }
func (CmdCheckpoint) fiCmdRead(fir fiReader) (cmd Cmd, err error) {
	_, err = fir.ReadLine()
	if err != nil {
		return nil, err
	}
	return CmdCheckpoint{}, nil
}

// progress ////////////////////////////////////////////////////////////////////

// CmdProgress requests that the Backend print the given string to its
// standard output channel.
type CmdProgress struct {
	Str string
}

func (c CmdProgress) fiCmdClass() cmdClass { return cmdClassCommand }
func (c CmdProgress) fiCmdWrite(fiw fiWriter) error {
	return fiw.WriteLine("progress", c.Str)
}
func init() { parser_registerCmd("progress ", CmdProgress{}) }
func (CmdProgress) fiCmdRead(fir fiReader) (cmd Cmd, err error) {
	line, err := fir.ReadLine()
	if err != nil {
		return nil, err
	}
	return CmdProgress{Str: trimLinePrefix(line, "progress ")}, nil
}

// feature /////////////////////////////////////////////////////////////////////

// CmdFeature requests that the Backend immediately aborts with an
// error if it does not support the specified feature.
type CmdFeature struct {
	Feature  string
	Argument string
}

func (c CmdFeature) fiCmdClass() cmdClass { return cmdClassCommand }
func (c CmdFeature) fiCmdWrite(fiw fiWriter) error {
	if c.Argument != "" {
		return fiw.WriteLine("feature", c.Feature+"="+c.Argument)
	} else {
		return fiw.WriteLine("feature", c.Feature)
	}
}
func init() { parser_registerCmd("feature ", CmdFeature{}) }
func (CmdFeature) fiCmdRead(fir fiReader) (cmd Cmd, err error) {
	// 'feature' SP <feature> ('=' <argument>)? LF
	line, err := fir.ReadLine()
	if err != nil {
		return nil, err
	}
	str := trimLinePrefix(line, "feature ")
	eq := strings.IndexByte(str, '=')
	if eq < 0 {
		return CmdFeature{Feature: str}, nil
	} else {
		return CmdFeature{
			Feature:  str[:eq],
			Argument: str[eq+1:],
		}, nil
	}
}

// option //////////////////////////////////////////////////////////////////////

// CmdOption requests that the Backend changes its settings.
type CmdOption struct {
	Option string
}

func (c CmdOption) fiCmdClass() cmdClass { return cmdClassCommand }
func (c CmdOption) fiCmdWrite(fiw fiWriter) error {
	return fiw.WriteLine("option", c.Option)
}
func init() { parser_registerCmd("option ", CmdOption{}) }
func (CmdOption) fiCmdRead(fir fiReader) (cmd Cmd, err error) {
	// 'option' SP <option> LF
	line, err := fir.ReadLine()
	if err != nil {
		return nil, err
	}
	return CmdOption{Option: trimLinePrefix(line, "option ")}, nil
}

// done ////////////////////////////////////////////////////////////////////////

// CmdDone indicates to the Backend that no more commands will be
// sent.
type CmdDone struct{}

func (c CmdDone) fiCmdClass() cmdClass { return cmdClassCommand }
func (c CmdDone) fiCmdWrite(fiw fiWriter) error {
	return fiw.WriteLine("done")
}
func init() { parser_registerCmd("done\n", CmdDone{}) }
func (CmdDone) fiCmdRead(fir fiReader) (cmd Cmd, err error) {
	_, err = fir.ReadLine()
	if err != nil {
		return nil, err
	}
	return CmdDone{}, nil
}
