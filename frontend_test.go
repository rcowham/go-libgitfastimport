// Tests for frontend

package libfastimport

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func runCmd(cmdLine string) (string, error) {
	cmd := exec.Command("/bin/bash", "-c", cmdLine)
	stdout, err := cmd.Output()
	if err != nil {
		return string(stdout), err
	}
	return string(stdout), nil
}

func createGitRepo(t *testing.T) string {
	d := t.TempDir()
	os.Chdir(d)
	runCmd("git init -b main")
	return d
}

func writeToFile(fname, contents string) {
	f, err := os.Create(fname)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	fmt.Fprint(f, contents)
	if err != nil {
		panic(err)
	}
}

func writeToTempFile(contents string) string {
	f, err := os.CreateTemp("", "*.txt")
	if err != nil {
		panic(err)
	}
	defer f.Close()
	fmt.Fprint(f, contents)
	if err != nil {
		fmt.Println(err)
	}
	return f.Name()
}

func TestParseBasic(t *testing.T) {
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

	buf := strings.NewReader(input)
	f := NewFrontend(buf, nil, nil)
	for {
		cmd, err := f.ReadCmd()
		if err != nil {
			if err != io.EOF {
				t.Errorf("ERROR: Failed to read cmd: %v\n", err)
			}
			break
		}
		switch cmd.(type) {
		case CmdBlob:
			blob := cmd.(CmdBlob)
			assert.Equal(t, 1, blob.Mark)
		case CmdReset:
			r := cmd.(CmdReset)
			assert.Equal(t, "refs/heads/main", r.RefName)
		case CmdCommit:
			c := cmd.(CmdCommit)
			assert.Equal(t, "refs/heads/main", c.Ref)
			assert.Equal(t, "Robert Cowham", c.Author.Name)
			assert.Equal(t, "rcowham@perforce.com", c.Author.Email)
			assert.Equal(t, "test\n", c.Msg)
		case CmdCommitEnd:
		case FileModify:
			f := cmd.(FileModify)
			assert.Equal(t, "test.txt", f.Path.String())
			assert.Equal(t, "100644", f.Mode.String())
			assert.Equal(t, ":1", f.DataRef)
		case FileDelete, FileCopy, FileRename:
			t.Error("Unexpected")
		default:
			t.Errorf("Unexpected cmd: %+v\n", cmd)
		}
	}
}

func TestParseRename(t *testing.T) {
	d := createGitRepo(t)
	os.Chdir(d)
	src := "src.txt"
	dst := "dst.txt"
	writeToFile(src, "contents")
	runCmd("git add .")
	runCmd("git commit -m initial")
	runCmd(fmt.Sprintf("git mv %s %s", src, dst))
	runCmd("git add .")
	runCmd("git commit -m renamed")
	export := "export.txt"
	// fast-export with rename detection implemented
	output, err := runCmd(fmt.Sprintf("git fast-export --all -M > %s", export))
	if err != nil {
		t.Errorf("ERROR: Failed to git export '%s': %v\n", export, err)
	}
	assert.Equal(t, "", output)

	file, err := os.Open(export)
	if err != nil {
		t.Errorf("ERROR: Failed to open file '%s': %v\n", export, err)
	}

	buf := bufio.NewReader(file)
	f := NewFrontend(buf, nil, nil)
	cmds := make([]Cmd, 0)
	for {
		cmd, err := f.ReadCmd()
		if err != nil {
			if err != io.EOF {
				t.Errorf("ERROR: Failed to read cmd: %v\n", err)
			}
			break
		}
		cmds = append(cmds, cmd)
	}
	counts := map[string]int{}
	for _, cmd := range cmds {
		switch cmd.(type) {
		case FileRename:
			f := cmd.(FileRename)
			assert.Equal(t, "src.txt", f.Src.String())
			assert.Equal(t, "dst.txt", f.Dst.String())
			k := fmt.Sprintf("%T", cmd)
			if _, ok := counts[k]; ok {
				counts[k] += 1
			} else {
				counts[k] = 1
			}
		default:
			k := fmt.Sprintf("%T", cmd)
			if _, ok := counts[k]; ok {
				counts[k] += 1
			} else {
				counts[k] = 1
			}
		}
	}
	// assert.Equal(t, "", fmt.Sprintf("%+v", counts))
	assert.Equal(t, 1, counts["libfastimport.FileModify"])
	assert.Equal(t, 1, counts["libfastimport.FileRename"])
}

func TestParseRenameSpaces(t *testing.T) {
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

	file, err := os.Open(export)
	if err != nil {
		t.Errorf("ERROR: Failed to open file '%s': %v\n", export, err)
	}

	buf := bufio.NewReader(file)
	f := NewFrontend(buf, nil, nil)
	cmds := make([]Cmd, 0)
	for {
		cmd, err := f.ReadCmd()
		if err != nil {
			if err != io.EOF {
				t.Errorf("ERROR: Failed to read cmd: %v\n", err)
			}
			break
		}
		cmds = append(cmds, cmd)
	}
	counts := map[string]int{}
	for _, cmd := range cmds {
		switch cmd.(type) {
		case FileRename:
			f := cmd.(FileRename)
			assert.Equal(t, src, string(f.Src))
			assert.Equal(t, dst, string(f.Dst))
			k := fmt.Sprintf("%T", cmd)
			if _, ok := counts[k]; ok {
				counts[k] += 1
			} else {
				counts[k] = 1
			}
		default:
			k := fmt.Sprintf("%T", cmd)
			if _, ok := counts[k]; ok {
				counts[k] += 1
			} else {
				counts[k] = 1
			}
		}
	}
	// assert.Equal(t, "", fmt.Sprintf("%+v", counts))
	assert.Equal(t, 1, counts["libfastimport.FileModify"])
	assert.Equal(t, 1, counts["libfastimport.FileRename"])
}

func TestParseRenameQuotes(t *testing.T) {
	d := createGitRepo(t)
	os.Chdir(d)
	src := "src's.txt"
	dst := "dst's.txt"
	writeToFile(src, "contents")
	runCmd("git add .")
	runCmd("git commit -m initial")
	output, err := runCmd(fmt.Sprintf("git mv \"%s\" \"%s\"", src, dst))
	if err != nil {
		t.Errorf("ERROR: Failed to git mv %v", err)
	}
	runCmd("git add .")
	runCmd("git commit -m renamed")
	export := "export.txt"
	// fast-export with rename detection implemented
	output, err = runCmd(fmt.Sprintf("git fast-export --all -M"))
	if err != nil {
		t.Errorf("ERROR: Failed to git export '%s': %v\n", export, err)
	}

	buf := strings.NewReader(output)
	f := NewFrontend(buf, nil, nil)
	cmds := make([]Cmd, 0)
	for {
		cmd, err := f.ReadCmd()
		if err != nil {
			if err != io.EOF {
				t.Errorf("ERROR: Failed to read cmd: %v\n", err)
			}
			break
		}
		cmds = append(cmds, cmd)
	}
	counts := map[string]int{}
	for _, cmd := range cmds {
		switch cmd.(type) {
		case FileRename:
			f := cmd.(FileRename)
			assert.Equal(t, src, string(f.Src))
			assert.Equal(t, dst, string(f.Dst))
			k := fmt.Sprintf("%T", cmd)
			if _, ok := counts[k]; ok {
				counts[k] += 1
			} else {
				counts[k] = 1
			}
		default:
			k := fmt.Sprintf("%T", cmd)
			if _, ok := counts[k]; ok {
				counts[k] += 1
			} else {
				counts[k] = 1
			}
		}
	}
	// assert.Equal(t, "", fmt.Sprintf("%+v", counts))
	assert.Equal(t, 1, counts["libfastimport.FileModify"])
	assert.Equal(t, 1, counts["libfastimport.FileRename"])
}
