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
			dataref := strings.TrimSuffix(strings.TrimPrefix(line, "cat-blob "), "\n")
			f.cmd <- CmdCatBlob{DataRef: dataref}
		case strings.HasPrefix(line, "get-mark :"):
			// 'get-mark' SP ':' <idnum> LF
			strIdnum := strings.TrimSuffix(strings.TrimPrefix(line, "get-mark :"), "\n")
			var nIdnum int
			nIdnum, err = strconv.Atoi(strIdnum)
			if err != nil {
				line = ""
				err = fmt.Errorf("get-mark: %v", err)
				return
			}
			f.cmd <- CmdGetMark{Mark: nIdnum}
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
	head := line[:nl]
	rest := line[nl+1:]
	if !strings.HasPrefix(head, "data ") {
		return "", fmt.Errorf("data: could not parse: %v", data)
	}
	if strings.HasPrefix(head, "data <<") {
		// Delimited format
		delim := strings.TrimPrefix(head, "data <<")
		suffix := "\n" + delim + "\n"
		if !strings.HasSuffix(rest, suffix) {
			return "", fmt.Errorf("data: did not find suffix: %v", suffix)
		}
		data = strings.TrimSuffix(rest, suffix)
	} else {
		// Exact byte count format
		strN := strings.TrimSuffix(head, "data ")
		intN, err := strconv.Atoi(strN)
		if err != nil {
			return "", err
		}
		if intN != len(rest) {
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
				str := strings.TrimSuffix(strings.TrimPrefix(line, "mark :"), "\n")
				c.Mark, err = strconv.Atoi(str)
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
		case strings.HasPrefix(line, "feature "):
			// 'feature' SP <feature> ('=' <argument>)? LF
			// TODO
		case strings.HasPrefix(line, "ls "):
			// 'ls' SP <dataref> SP <path> LF
			// TODO
		case strings.HasPrefix(line, "option "):
			// 'option' SP <option> LF
			// TODO
		case strings.HasPrefix(line, "progress "):
			// 'progress' SP <any> LF
			str := strings.TrimSuffix(strings.TrimPrefix(line, "progress "), "\n")
			f.cmd <- CmdProgress{Str: str}
		case strings.HasPrefix(line, "reset "):
			// 'reset' SP <ref> LF
			// ('from' SP <commit-ish> LF)?
			// TODO
		case strings.HasPrefix(line, "tag "):
			// 'tag' SP <name> LF
			// 'from' SP <commit-ish> LF
			// 'tagger' (SP <name>)? SP LT <email> GT SP <when> LF
			// data
			// TODO
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
