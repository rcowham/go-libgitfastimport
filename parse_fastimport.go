package libfastimport

import (
	"fmt"
	"strings"

	"git.lukeshu.com/go/libfastimport/textproto"
)

var parser_regularCmds = make(map[string]Cmd)
var parser_commentCmds = make(map[string]Cmd)

func parser_registerCmd(prefix string, cmd Cmd) {
	switch cmd.fiCmdClass() {
	case cmdClassCommand, cmdClassCommit:
		parser_regularCmds[prefix] = cmd
	case cmdClassComment:
		parser_commentCmds[prefix] = cmd
	default:
		panic(fmt.Errorf("invalid cmdClass: %d", cmd.fiCmdClass()))
	}
}

var parser_regular func(line string) func(fiReader) (Cmd, error)
var parser_comment func(line string) func(fiReader) (Cmd, error)

func parser_compile(cmds map[string]Cmd) func(line string) func(fiReader) (Cmd, error) {
	// This assumes that 2 characters is enough to uniquely
	// identify a command, and that "#" is the only one-chara
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

		switch cmd.fiCmdClass() {
		case cmdClassCommand:
			if p.inCommit {
				p.ret_cmd <- CmdCommitEnd{}
			}
			_, p.inCommit = cmd.(CmdCommit)
		case cmdClassCommit:
			if !p.inCommit {
				return fmt.Errorf("Got in-commit-only command outside of a commit: %[1]T(%#[1]v)", cmd)
			}
		case cmdClassComment:
			/* do nothing */
		default:
			panic(fmt.Errorf("invalid cmdClass: %d", cmd.fiCmdClass()))
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
