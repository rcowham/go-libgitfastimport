// Tests for frontend

package libfastimport

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

type MyWriteCloser struct {
	*bufio.Writer
}

func (mwc *MyWriteCloser) Close() error {
	// Noop
	return nil
}

func TestWriteBasic(t *testing.T) {
	input := `blob
mark :1
data 5
test
reset refs/heads/main
commit refs/heads/main
mark :2
author Robert Cowham <rcowham@perforce.com> 1644399073 +0000
committer Robert Cowham <rcowham@perforce.com> 1644399073 +0000
data 5
test
M 100644 :1 test.txt

`

	inbuf := strings.NewReader(input)
	outbuf := new(bytes.Buffer)
	frontend := NewFrontend(inbuf, nil, nil)
	bw := bufio.NewWriter(outbuf)
	mwc := &MyWriteCloser{bw}
	backend := NewBackend(mwc, nil, nil)
	for {
		cmd, err := frontend.ReadCmd()
		if err != nil {
			if err != io.EOF {
				t.Errorf("ERROR: Failed to read cmd: %v\n", err)
			}
			break
		}
		switch cmd.(type) {
		case CmdBlob:
			blob := cmd.(CmdBlob)
			backend.Do(blob)
		case CmdReset:
			r := cmd.(CmdReset)
			backend.Do(r)
		case CmdCommit:
			c := cmd.(CmdCommit)
			backend.Do(c)
		case CmdCommitEnd:
			c := cmd.(CmdCommitEnd)
			backend.Do(c)
		case FileModify:
			f := cmd.(FileModify)
			backend.Do(f)
		case FileDelete, FileCopy, FileRename:
			t.Error("Unexpected")
		default:
			t.Errorf("Unexpected cmd: %+v\n", cmd)
		}
	}
	bw.Flush()
	outstr := outbuf.String()
	assert.Equal(t, input, outstr)
}

func TestRenameSpaces(t *testing.T) {
	d := createGitRepo(t)
	os.Chdir(d)
	src := "src space.txt"
	dst := "dst space.txt"
	writeToFile(src, "contents")
	runCmd("git add .")
	runCmd("git commit -m initial")
	runCmd(fmt.Sprintf("git mv '%s' '%s'", src, dst))
	runCmd("git add .")
	runCmd("git commit -m renamed")
	export := "export.txt"
	// fast-export with rename detection implemented
	output, err := runCmd(fmt.Sprintf("git fast-export --all -M > %s", export))
	if err != nil {
		t.Errorf("ERROR: Failed to git export '%s': %v\n", export, err)
	}
	assert.Equal(t, "", output)

	input, err := os.ReadFile(export)
	if err != nil {
		t.Errorf("ERROR: Failed to read file '%s': %v\n", export, err)
	}

	inbuf := strings.NewReader(string(input))
	frontend := NewFrontend(inbuf, nil, nil)
	outbuf := new(bytes.Buffer)
	bw := bufio.NewWriter(outbuf)
	mwc := &MyWriteCloser{bw}
	backend := NewBackend(mwc, nil, nil)
	for {
		cmd, err := frontend.ReadCmd()
		if err != nil {
			if err != io.EOF {
				t.Errorf("ERROR: Failed to read cmd: %v\n", err)
			}
			break
		}
		switch cmd.(type) {
		case CmdBlob, CmdReset, CmdCommit, CmdCommitEnd, FileModify, FileDelete, FileCopy, FileRename:
			backend.Do(cmd)
		default:
			t.Errorf("Unexpected cmd: %+v\n", cmd)
		}
	}
	bw.Flush()
	outstr := outbuf.String()
	assert.Equal(t, string(input), outstr)
}
