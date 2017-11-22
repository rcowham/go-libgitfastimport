package libfastimport

import (
	"bufio"
	"fmt"
	"io"
	"os"
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

	inCommit bool

	cmd chan Cmd
	err error
}

func NewFrontend(fastImport io.Reader, catBlob io.Writer, onErr func(error) error) *Frontend {
	ret := &Frontend{}
	ret.fir = textproto.NewFIReader(fastImport)
	if catBlob == nil {
		catBlob = os.Stdout
	}
	ret.w = bufio.NewWriter(catBlob)
	ret.cbw = textproto.NewCatBlobWriter(ret.w)
	ret.cmd = make(chan Cmd)
	go func() {
		ret.err = ret.parse()
		if onErr != nil {
			ret.err = onErr(ret.err)
		}
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
			c := CmdCommit{Ref: trimLinePrefix(line, "commit ")}

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
			if strings.HasPrefix(line, "author ") {
				author, err := textproto.ParseIdent(trimLinePrefix(line, "author "))
				if err != nil {
					return err
				}
				c.Author = &author
				line, err = f.nextLine()
				if err != nil {
					return err
				}
			}
			if !strings.HasPrefix(line, "committer ") {
				return fmt.Errorf("commit: expected committer command: %v", line)
			}
			c.Committer, err = textproto.ParseIdent(trimLinePrefix(line, "committer "))
			if err != nil {
				return err
			}

			line, err = f.nextLine()
			if err != nil {
				return err
			}
			if !strings.HasPrefix(line, "data ") {
				return fmt.Errorf("commit: expected data command: %v", line)
			}
			c.Msg, err = parse_data(line)
			if err != nil {
				return err
			}

			line, err = f.nextLine()
			if err != nil {
				return err
			}
			if strings.HasPrefix(line, "from ") {
				c.From = trimLinePrefix(line, "from ")
				line, err = f.nextLine()
				if err != nil {
					return err
				}
			}
			for strings.HasPrefix(line, "merge ") {
				c.Merge = append(c.Merge, trimLinePrefix(line, "merge "))
				line, err = f.nextLine()
				if err != nil {
					return err
				}
			}
			f.cmd <- c
		case strings.HasPrefix(line, "M "):
			fmt.Printf("line: %q\n", line)
			str := trimLinePrefix(line, "M ")
			sp1 := strings.IndexByte(str, ' ')
			sp2 := strings.IndexByte(str[sp1+1:], ' ')
			if sp1 < 0 || sp2 < 0 {
				return fmt.Errorf("commit: malformed modify command: %v", line)
			}
			nMode, err := strconv.ParseUint(str[:sp1], 8, 18)
			if err != nil {
				return err
			}
			ref := str[sp1+1 : sp2]
			path := textproto.PathUnescape(str[sp2+1:])
			if ref == "inline" {
				line, err = f.nextLine()
				if err != nil {
					return err
				}
				if !strings.HasPrefix(line, "data ") {
					return fmt.Errorf("commit: modify: expected data command: %v", line)
				}
				data, err := parse_data(line)
				if err != nil {
					return err
				}
				f.cmd <- FileModifyInline{Mode: textproto.Mode(nMode), Path: path, Data: data}
			} else {
				f.cmd <- FileModify{Mode: textproto.Mode(nMode), Path: path, DataRef: ref}
			}
		case strings.HasPrefix(line, "D "):
			f.cmd <- FileDelete{Path: textproto.PathUnescape(trimLinePrefix(line, "D "))}
		case strings.HasPrefix(line, "C "):
			// BUG(lukeshu): TODO: commit C not implemented
			panic("TODO: commit C not implemented")
		case strings.HasPrefix(line, "R "):
			// BUG(lukeshu): TODO: commit R not implemented
			panic("TODO: commit R not implemented")
		case strings.HasPrefix(line, "N "):
			str := trimLinePrefix(line, "N ")
			sp := strings.IndexByte(str, ' ')
			if sp < 0 {
				return fmt.Errorf("commit: malformed notemodify command: %v", line)
			}
			ref := str[:sp]
			commitish := str[sp+1:]
			if ref == "inline" {
				line, err = f.nextLine()
				if err != nil {
					return err
				}
				if !strings.HasPrefix(line, "data ") {
					return fmt.Errorf("commit: notemodify: expected data command: %v", line)
				}
				data, err := parse_data(line)
				if err != nil {
					return err
				}
				f.cmd <- NoteModifyInline{CommitIsh: commitish, Data: data}
			} else {
				f.cmd <- NoteModify{CommitIsh: commitish, DataRef: ref}
			}
		case line == "deleteall\n":
			f.cmd <- FileDeleteAll{}
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
			str := trimLinePrefix(line, "ls ")
			sp := -1
			if !strings.HasPrefix(str, "\"") {
				sp = strings.IndexByte(line, ' ')
			}
			c := CmdLs{}
			c.Path = textproto.PathUnescape(str[sp+1:])
			if sp >= 0 {
				c.DataRef = str[:sp]
			}
			f.cmd <- c
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
		switch cmd.fiCmdClass() {
		case cmdClassCommand:
			_, f.inCommit = cmd.(CmdCommit)
		case cmdClassCommit:
			if !f.inCommit {
				// BUG(lukeshu): idk what to do here
				panic("oops")
			}
		case cmdClassComment:
			/* do nothing */
		default:
			panic(fmt.Errorf("invalid cmdClass: %d", cmd.fiCmdClass()))
		}
		return cmd, nil
	}
	return nil, f.err
}

func (f *Frontend) RespondGetMark(sha1 string) error {
	err := f.cbw.WriteLine(sha1)
	if err != nil {
		return err
	}
	return f.w.Flush()
}

func (f *Frontend) RespondCatBlob(sha1 string, data string) error {
	err := f.cbw.WriteBlob(sha1, data)
	if err != nil {
		return err
	}
	return f.w.Flush()
}

func (f *Frontend) RespondLs(mode textproto.Mode, dataref string, path textproto.Path) error {
	var err error
	if mode == 0 {
		err = f.cbw.WriteLine("missing", path)
	} else {
		var t string
		switch mode {
		case textproto.ModeDir:
			t = "tree"
		case textproto.ModeGit:
			t = "commit"
		default:
			t = "blob"
		}
		err = f.cbw.WriteLine(mode, t, dataref+"\t"+textproto.PathEscape(path))
	}
	if err != nil {
		return err
	}
	return f.w.Flush()
}
