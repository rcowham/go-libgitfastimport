package libfastimport

type UnsupportedCommand string

func (e UnsupportedCommand) Error() string {
	return "Unsupported command: "+string(e)
}

type Parser struct {
	fir *FIReader

	cmd chan Cmd
}

func (p *Parser) GetCmd() (Cmd, error) {
	for p.cmd == nil {
		slice, err := p.fir.ReadSlice()
		if err != nil {
			return nil, err
		}
		err = p.putSlice(slice)
		if err != nil {
			return nil, err
		}
	}
	return <-p.cmd, nil
}

func (p *Parser) putSlice(slice []byte) error {
	if len(slice) < 1 {
		return UnsupportedCommand(slice)
	}
	switch slice[0] {
	case '#': // comment
	case 'b': // blob
	case 'c':
		if len(slice) < 2 {
			return UnsupportedCommand(slice)
		}
		switch slice[1] {
		case 'o': // commit
		case 'h': // checkpoint
		case 'a': // cat-blob
		default:
			return UnsupportedCommand(slice)
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
		return UnsupportedCommand(slice)
	}
	return nil // TODO
}
