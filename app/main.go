package main

import (
	"bufio"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
)

// Usage: mygrep [-r] -E <pattern> [path ...]
func main() {
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

	re, err := Compile(pattern)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: invalid pattern: %v\n", err)
		os.Exit(2)
	}

	found := false

	if len(paths) == 0 {
		if recursive {
			paths = []string{"."}
		} else {
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
			if found {
				os.Exit(0)
			}
			os.Exit(1)
		}
	}

	multiPrefix := len(paths) > 1 || recursive

	for _, p := range paths {
		if recursive {
			walkErr := filepath.WalkDir(p, func(fpath string, d fs.DirEntry, err error) error {
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
		} else {
			fi, serr := os.Stat(p)
			if serr != nil {
				fmt.Fprintf(os.Stderr, "error: %v\n", serr)
				os.Exit(2)
			}
			if fi.IsDir() {
				fmt.Fprintf(os.Stderr, "mygrep: %s: Is a directory\n", p)
				os.Exit(2)
			}
			reader, oerr := os.Open(p)
			if oerr != nil {
				fmt.Fprintf(os.Stderr, "error: open %s: %v\n", p, oerr)
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
						fmt.Printf("%s:%s\n", p, line)
					} else {
						fmt.Println(line)
					}
					found = true
				}
			}
			if serr := scanner.Err(); serr != nil {
				fmt.Fprintf(os.Stderr, "error: read %s: %v\n", p, serr)
				os.Exit(2)
			}
			reader.Close()
		}
	}

	if found {
		os.Exit(0)
	}
	os.Exit(1)
}
