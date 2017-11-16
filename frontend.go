package libfastimport

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
	"strings"

	"git.lukeshu.com/go/libfastimport/textproto"
)

type UnsupportedCommand string

func (e UnsupportedCommand) Error() string {
	return "Unsupported command: " + string(e)
}

func trimLinePrefix(line string, prefix string) string {
	if !strings.HasPrefix(line, prefix) {
		panic("line didn't have prefix")
	}
	if !strings.HasSuffix(line, "\n") {
		panic("line didn't have prefix")
	}
	return strings.TrimSuffix(strings.TrimPrefix(line, prefix), "\n")
}

// A Frontend is something that produces a fast-import stream; the
// Frontend object provides methods for reading from it.
type Frontend struct {
	fir *textproto.FIReader
	cbw *textproto.CatBlobWriter
	w   *bufio.Writer

	cmd chan Cmd
	err error
}

func NewFrontend(fastImport io.Reader, catBlob io.Writer) *Frontend {
	ret := &Frontend{}
	ret.fir = textproto.NewFIReader(fastImport)
	if catBlob != nil {
		ret.w = bufio.NewWriter(catBlob)
		ret.cbw = textproto.NewCatBlobWriter(ret.w)
	}
	ret.cmd = make(chan Cmd)
	go func() {
		ret.err = ret.parse()
		close(ret.cmd)
	}()
	return ret
}

func (f *Frontend) nextLine() (line string, err error) {
	for {
		line, err = f.fir.ReadLine()
		if err != nil {
			return
		}
		switch {
		case strings.HasPrefix(line, "#"):
			f.cmd <- CmdComment{Comment: line[1:]}
		case strings.HasPrefix(line, "cat-blob "):
			// 'cat-blob' SP <dataref> LF
			f.cmd <- CmdCatBlob{DataRef: trimLinePrefix(line, "cat-blob ")}
		case strings.HasPrefix(line, "get-mark :"):
			// 'get-mark' SP ':' <idnum> LF
			c := CmdGetMark{}
			c.Mark, err = strconv.Atoi(trimLinePrefix(line, "get-mark :"))
			if err != nil {
				line = ""
				err = fmt.Errorf("get-mark: %v", err)
				return
			}
			f.cmd <- c
		default:
			return
		}
	}
}

func parse_data(line string) (data string, err error) {
	nl := strings.IndexByte(line, '\n')
	if nl < 0 {
		return "", fmt.Errorf("data: expected newline: %v", data)
	}
	head := line[:nl+1]
	rest := line[nl+1:]
	if !strings.HasPrefix(head, "data ") {
		return "", fmt.Errorf("data: could not parse: %v", data)
	}
	if strings.HasPrefix(head, "data <<") {
		// Delimited format
		delim := trimLinePrefix(head, "data <<")
		suffix := "\n" + delim + "\n"
		if !strings.HasSuffix(rest, suffix) {
			return "", fmt.Errorf("data: did not find suffix: %v", suffix)
		}
		data = strings.TrimSuffix(rest, suffix)
	} else {
		// Exact byte count format
		size, err := strconv.Atoi(trimLinePrefix(head, "data "))
		if err != nil {
			return "", err
		}
		if size != len(rest) {
			panic("FIReader should not have let this happen")
		}
		data = rest
	}
	return
}

func (f *Frontend) parse() error {
	line, err := f.nextLine()
	if err != nil {
		return err
	}
	for {
		switch {
		case line == "blob\n":
			// 'blob' LF
			// mark?
			// data
			c := CmdBlob{}
			line, err = f.nextLine()
			if err != nil {
				return err
			}
			if strings.HasPrefix(line, "mark :") {
				c.Mark, err = strconv.Atoi(trimLinePrefix(line, "mark :"))
				if err != nil {
					return err
				}
				line, err = f.nextLine()
				if err != nil {
					return err
				}
			}
			if !strings.HasPrefix(line, "data ") {
				return fmt.Errorf("Unexpected command in blob: %q", line)
			}
			c.Data, err = parse_data(line)
			if err != nil {
				return err
			}
			f.cmd <- c
		case line == "checkpoint\n":
			f.cmd <- CmdCheckpoint{}
		case line == "done\n":
			f.cmd <- CmdDone{}
		case strings.HasPrefix(line, "commit "):
			// 'commit' SP <ref> LF
			// mark?
			// ('author' (SP <name>)? SP LT <email> GT SP <when> LF)?
			// 'committer' (SP <name>)? SP LT <email> GT SP <when> LF
			// data
			// ('from' SP <commit-ish> LF)?
			// ('merge' SP <commit-ish> LF)*
			// (filemodify | filedelete | filecopy | filerename | filedeleteall | notemodify)*
			// TODO
			continue
		case strings.HasPrefix(line, "feature "):
			// 'feature' SP <feature> ('=' <argument>)? LF
			str := trimLinePrefix(line, "feature ")
			eq := strings.IndexByte(str, '=')
			if eq < 0 {
				f.cmd <- CmdFeature{
					Feature: str,
				}
			} else {
				f.cmd <- CmdFeature{
					Feature:  str[:eq],
					Argument: str[eq+1:],
				}
			}
		case strings.HasPrefix(line, "ls "):
			// 'ls' SP <dataref> SP <path> LF
			sp1 := strings.IndexByte(line, ' ')
			sp2 := strings.IndexByte(line[sp1+1:], ' ')
			lf := strings.IndexByte(line[sp2+1:], '\n')
			if sp1 < 0 || sp2 < 0 || lf < 0 {
				return fmt.Errorf("ls: outside of a commit both <dataref> and <path> are required: %v", line)
			}
			f.cmd <- CmdLs{
				DataRef: line[sp1+1 : sp2],
				Path:    textproto.PathUnescape(line[sp2+1 : lf]),
			}
		case strings.HasPrefix(line, "option "):
			// 'option' SP <option> LF
			f.cmd <- CmdOption{Option: trimLinePrefix(line, "option ")}
		case strings.HasPrefix(line, "progress "):
			// 'progress' SP <any> LF
			f.cmd <- CmdProgress{Str: trimLinePrefix(line, "progress ")}
		case strings.HasPrefix(line, "reset "):
			// 'reset' SP <ref> LF
			// ('from' SP <commit-ish> LF)?
			c := CmdReset{RefName: trimLinePrefix(line, "reset ")}
			line, err = f.nextLine()
			if err != nil {
				return err
			}
			if strings.HasPrefix(line, "from ") {
				c.CommitIsh = trimLinePrefix(line, "from ")
				line, err = f.nextLine()
				if err != nil {
					return err
				}
			}
			f.cmd <- c
			continue
		case strings.HasPrefix(line, "tag "):
			// 'tag' SP <name> LF
			// 'from' SP <commit-ish> LF
			// 'tagger' (SP <name>)? SP LT <email> GT SP <when> LF
			// data
			c := CmdTag{RefName: trimLinePrefix(line, "tag ")}

			line, err = f.nextLine()
			if err != nil {
				return err
			}
			if !strings.HasPrefix(line, "from ") {
				return fmt.Errorf("tag: expected from command: %v", line)
			}
			c.CommitIsh = trimLinePrefix(line, "from ")

			line, err = f.nextLine()
			if err != nil {
				return err
			}
			if !strings.HasPrefix(line, "tagger ") {
				return fmt.Errorf("tag: expected tagger command: %v", line)
			}
			c.Tagger, err = textproto.ParseIdent(trimLinePrefix(line, "tagger "))
			if err != nil {
				return err
			}

			line, err = f.nextLine()
			if err != nil {
				return err
			}
			c.Data, err = parse_data(line)
			if err != nil {
				return err
			}
			f.cmd <- c
		default:
			return UnsupportedCommand(line)
		}
		line, err = f.nextLine()
		if err != nil {
			return err
		}
	}
}

func (f *Frontend) ReadCmd() (Cmd, error) {
	cmd, ok := <-f.cmd
	if ok {
		return cmd, nil
	}
	return nil, f.err
}

func (f *Frontend) RespondGetMark(sha1 string) error {
	// TODO
	return f.w.Flush()
}

func (f *Frontend) RespondCatBlob(sha1 string, data string) error {
	// TODO
	return f.w.Flush()
}

func (f *Frontend) RespondLs(mode textproto.Mode, dataref string, path textproto.Path) error {
	// TODO
	return f.w.Flush()
}
