package libfastimport

import (
	"git.lukeshu.com/go/libfastimport/textproto"
)

type UnsupportedCommand string

func (e UnsupportedCommand) Error() string {
	return "Unsupported command: "+string(e)
}

type Parser struct {
	fir *textproto.FIReader

	cmd chan Cmd
}

func (p *Parser) GetCmd() (Cmd, error) {
	for p.cmd == nil {
		line, err := p.fir.ReadLine()
		if err != nil {
			return nil, err
		}
		err = p.putLine(line)
		if err != nil {
			return nil, err
		}
	}
	return <-p.cmd, nil
}

func (p *Parser) putLine(line string) error {
	if len(line) < 1 {
		return UnsupportedCommand(line)
	}
	switch line[0] {
	case '#': // comment
	case 'b': // blob
	case 'c':
		if len(line) < 2 {
			return UnsupportedCommand(line)
		}
		switch line[1] {
		case 'o': // commit
		case 'h': // checkpoint
		case 'a': // cat-blob
		default:
			return UnsupportedCommand(line)
		}
	case 'd': // done
	case 'f': // feature
	case 'g': // get-mark
	case 'l': // ls
	case 'o': // option
	case 'p': // progress
	case 'r': // reset
	case 't': // tag
	default:
		return UnsupportedCommand(line)
	}
	return nil // TODO
}
