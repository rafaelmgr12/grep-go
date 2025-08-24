package main

import (
	"bufio"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
)

// Args holds parsed command-line arguments.
type Args struct {
	Recursive bool
	Pattern   string
	Paths     []string
}

// parseArgs parses command-line arguments and returns an Args struct.
func parseArgs() Args {
	var recursive bool
	i := 1
	if len(os.Args) > i && os.Args[i] == "-r" {
		recursive = true
		i++
	}
	if len(os.Args) <= i || os.Args[i] != "-E" {
		fmt.Fprintf(os.Stderr, "usage: mygrep [-r] -E <pattern> [path ...]\n")
		os.Exit(2)
	}
	i++
	if len(os.Args) <= i {
		fmt.Fprintf(os.Stderr, "error: missing pattern\n")
		os.Exit(2)
	}
	pattern := os.Args[i]
	i++
	paths := os.Args[i:]
	return Args{Recursive: recursive, Pattern: pattern, Paths: paths}
}

// grepStdin reads from standard input and prints matching lines.
func grepStdin(re *Regex) bool {
	found := false
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		line := scanner.Text()
		ok, matchErr := re.Match([]byte(line))
		if matchErr != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", matchErr)
			os.Exit(2)
		}
		if ok {
			fmt.Println(line)
			found = true
		}
	}
	if err := scanner.Err(); err != nil {
		fmt.Fprintf(os.Stderr, "error: read input: %v\n", err)
		os.Exit(2)
	}
	return found
}

// grepFile searches for matches in a single file.
func grepFile(re *Regex, path string, multiPrefix bool) bool {
	found := false
	fi, serr := os.Stat(path)
	if serr != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", serr)
		os.Exit(2)
	}
	if fi.IsDir() {
		fmt.Fprintf(os.Stderr, "mygrep: %s: Is a directory\n", path)
		os.Exit(2)
	}
	reader, oerr := os.Open(path)
	if oerr != nil {
		fmt.Fprintf(os.Stderr, "error: open %s: %v\n", path, oerr)
		os.Exit(2)
	}
	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		line := scanner.Text()
		ok, merr := re.Match([]byte(line))
		if merr != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", merr)
			os.Exit(2)
		}
		if ok {
			if multiPrefix {
				fmt.Printf("%s:%s\n", path, line)
			} else {
				fmt.Println(line)
			}
			found = true
		}
	}
	if serr := scanner.Err(); serr != nil {
		fmt.Fprintf(os.Stderr, "error: read %s: %v\n", path, serr)
		os.Exit(2)
	}
	reader.Close()
	return found
}

// grepRecursive searches for matches recursively in directories.
func grepRecursive(re *Regex, root string, multiPrefix bool) bool {
	found := false
	walkErr := filepath.WalkDir(root, func(fpath string, d fs.DirEntry, err error) error {
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			return err
		}
		if d.IsDir() {
			return nil
		}
		reader, oerr := os.Open(fpath)
		if oerr != nil {
			fmt.Fprintf(os.Stderr, "error: open %s: %v\n", fpath, oerr)
			return oerr
		}
		scanner := bufio.NewScanner(reader)
		for scanner.Scan() {
			line := scanner.Text()
			ok, merr := re.Match([]byte(line))
			if merr != nil {
				fmt.Fprintf(os.Stderr, "error: %v\n", merr)
				os.Exit(2)
			}
			if ok {
				if multiPrefix {
					fmt.Printf("%s:%s\n", fpath, line)
				} else {
					fmt.Println(line)
				}
				found = true
			}
		}
		if serr := scanner.Err(); serr != nil {
			fmt.Fprintf(os.Stderr, "error: read %s: %v\n", fpath, serr)
			return serr
		}
		reader.Close()
		return nil
	})
	if walkErr != nil {
		os.Exit(2)
	}
	return found
}
