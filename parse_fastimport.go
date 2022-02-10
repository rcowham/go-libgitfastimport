// Copyright (C) 2017-2018, 2020-2021  Luke Shumaker <lukeshu@lukeshu.com>
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
	"io"
	"strings"

	"github.com/pkg/errors"
	"github.com/rcowham/go-libgitfastexport/textproto"
)

var parser_regularCmds = make(map[string]Cmd)
var parser_commentCmds = make(map[string]Cmd)

func parser_registerCmd(prefix string, cmd Cmd) {
	if cmdIs(cmd, cmdClassInCommand) {
		parser_commentCmds[prefix] = cmd
	} else {
		parser_regularCmds[prefix] = cmd
	}
}

var parser_regular func(line string) func(fiReader) (Cmd, error)
var parser_comment func(line string) func(fiReader) (Cmd, error)

func parser_compile(cmds map[string]Cmd) func(line string) func(fiReader) (Cmd, error) {
	// This assumes that 2 characters is enough to uniquely
	// identify a command, and that "#" is the only one-character
	// command.
	ch2map := make(map[string]string, len(cmds))
	for prefix := range cmds {
		var ch2 string
		if len(prefix) < 2 {
			if prefix != "#" {
				panic("the assumptions of parser_compile are invalid, it must be rewritten")
			}
			ch2 = prefix
		} else {
			ch2 = prefix[:2]
		}
		if _, dup := ch2map[ch2]; dup {
			panic("the assumptions of parser_compile are invalid, it must be rewritten")
		}
		ch2map[ch2] = prefix
	}
	return func(line string) func(fiReader) (Cmd, error) {
		n := 2
		if strings.HasPrefix(line, "#") {
			n = 1
		}
		if len(line) < n {
			return nil
		}
		prototype := cmds[ch2map[line[:n]]]
		if prototype == nil {
			return nil
		}
		return prototype.fiCmdRead
	}
}

type parser struct {
	fir *textproto.FIReader

	inCommit bool

	buf_line *string
	buf_err  error

	ret_cmd chan Cmd
	ret_err error
}

func newParser(fir *textproto.FIReader) *parser {
	if parser_regular == nil {
		parser_regular = parser_compile(parser_regularCmds)
	}
	if parser_comment == nil {
		parser_comment = parser_compile(parser_commentCmds)
	}

	ret := &parser{
		fir:     fir,
		ret_cmd: make(chan Cmd),
	}
	go func() {
		ret.ret_err = ret.parse()
		close(ret.ret_cmd)
	}()
	return ret
}

func (p *parser) ReadCmd() (Cmd, error) {
	cmd, ok := <-p.ret_cmd
	if !ok {
		return nil, p.ret_err
	}
	return cmd, nil
}

func (p *parser) parse() error {
	for {
		line, err := p.PeekLine()
		if err != nil {
			if err == io.EOF && p.inCommit {
				p.ret_cmd <- CmdCommitEnd{}
			}
			return err
		}
		subparser := parser_regular(line)
		if subparser == nil {
			return UnsupportedCommand(line)
		}
		cmd, err := subparser(p)
		if err != nil {
			return err
		}

		switch {
		case !cmdIs(cmd, cmdClassInCommit):
			if p.inCommit {
				p.ret_cmd <- CmdCommitEnd{}
			}
			_, p.inCommit = cmd.(CmdCommit)
		case !p.inCommit && !cmdIs(cmd, cmdClassCommand):
			return errors.Errorf("Got in-commit-only command outside of a commit: %[1]T(%#[1]v)", cmd)
		}

		p.ret_cmd <- cmd
	}
}

func (p *parser) PeekLine() (string, error) {
	for p.buf_line == nil && p.buf_err == nil {
		var line string
		line, p.buf_err = p.fir.ReadLine()
		p.buf_line = &line
		if p.buf_err != nil {
			return *p.buf_line, p.buf_err
		}
		subparser := parser_comment(line)
		if subparser != nil {
			var cmd Cmd
			cmd, p.buf_err = subparser(p)
			if p.buf_err != nil {
				return "", p.buf_err
			}
			p.ret_cmd <- cmd
		}
	}
	return *p.buf_line, p.buf_err
}

func (p *parser) ReadLine() (string, error) {
	line, err := p.PeekLine()
	p.buf_line = nil
	p.buf_err = nil
	return line, err
}
