package main

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
)

func TestParseArgs(t *testing.T) {
	orig := os.Args
	defer func() { os.Args = orig }()
	os.Args = []string{"mygrep", "-r", "-E", "pattern", "file1", "file2"}
	args := parseArgs()
	if !args.Recursive || args.Pattern != "pattern" || len(args.Paths) != 2 || args.Paths[0] != "file1" || args.Paths[1] != "file2" {
		t.Fatalf("unexpected args: %#v", args)
	}
}

func captureOutput(f func()) string {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	f()
	w.Close()
	os.Stdout = old
	var buf bytes.Buffer
	io.Copy(&buf, r)
	r.Close()
	return buf.String()
}

func TestGrepStdin(t *testing.T) {
	re, _ := Compile("foo")
	inR, inW, _ := os.Pipe()
	oldIn := os.Stdin
	os.Stdin = inR
	go func() {
		defer inW.Close()
		inW.WriteString("bar\nfoo\n")
	}()
	out := captureOutput(func() {
		if !grepStdin(re) {
			t.Fatalf("expected match")
		}
	})
	os.Stdin = oldIn
	if strings.TrimSpace(out) != "foo" {
		t.Fatalf("unexpected output %q", out)
	}
}

func TestGrepFile(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "test.txt")
	os.WriteFile(file, []byte("hello\nfoo\nbar\n"), 0644)
	re, _ := Compile("foo")
	out := captureOutput(func() {
		if !grepFile(re, file, false) {
			t.Fatalf("expected match")
		}
	})
	if strings.TrimSpace(out) != "foo" {
		t.Fatalf("unexpected output %q", out)
	}
	out2 := captureOutput(func() {
		if !grepFile(re, file, true) {
			t.Fatalf("expected match")
		}
	})
	expected := file + ":foo"
	if strings.TrimSpace(out2) != expected {
		t.Fatalf("unexpected output with prefix %q", out2)
	}
}

func TestGrepRecursive(t *testing.T) {
	root := t.TempDir()
	f1 := filepath.Join(root, "f1.txt")
	os.WriteFile(f1, []byte("foo\n"), 0644)
	sub := filepath.Join(root, "sub")
	os.Mkdir(sub, 0755)
	f2 := filepath.Join(sub, "f2.txt")
	os.WriteFile(f2, []byte("bar\nfoo\n"), 0644)
	re, _ := Compile("foo")
	out := captureOutput(func() {
		if !grepRecursive(re, root, true) {
			t.Fatalf("expected match")
		}
	})
	lines := strings.Split(strings.TrimSpace(out), "\n")
	if len(lines) != 2 {
		t.Fatalf("expected 2 lines, got %d: %q", len(lines), out)
	}
	expected1 := f1 + ":foo"
	expected2 := f2 + ":foo"
	sort.Strings(lines)
	if lines[0] != expected1 || lines[1] != expected2 {
		t.Fatalf("unexpected lines %v", lines)
	}
}
