# mygrep Documentation

This documentation provides an in-depth overview of the design, implementation, and usage of the `mygrep` project—a minimal grep clone written in Go with a custom regular expression engine.

## Table of Contents

1. Introduction
2. Project Structure
3. Usage
4. Regular Expression Engine
5. File and Directory Traversal
6. Error Handling
7. Extending the Project
8. References

---

## 1. Introduction

`mygrep` is an educational project that demonstrates how to build a command-line tool for searching text using regular expressions. Unlike most grep clones, it features a custom regex engine implemented from scratch, supporting basic regex features such as groups, alternation, quantifiers, character classes, and anchors.

## 2. Project Structure

- `main.go`: Handles command-line arguments, input/output, and file traversal.
- `re.go`: Implements the custom regular expression engine, including parsing and matching logic.
- `parser.go`: Provides utilities for parsing regex patterns, handling groups and alternation.
- `state.go`: (if present) Manages state/environment for regex matching, such as group captures.
- `go.mod`, `go.sum`: Go module files for dependency management.
- `README.md`: Project overview and quick usage guide.
- `docs/`: This documentation folder.

## 3. Usage

### Buildingb

```sh
go build -o mygrep .
```

### Command-Line Options

```
./mygrep [-r] -E <pattern> [path ...]
```

- `-r`: Recursively search directories.
- `-E <pattern>`: Specify the regex pattern to search for.
- `[path ...]`: One or more files or directories to search. If omitted, reads from standard input.

### Examples

- Search for "hello" in standard input:
  ```sh
  echo -e "hello\nworld" | ./mygrep -E "hello"
  ```
- Search recursively in a directory:
  ```sh
  ./mygrep -r -E "pattern" ./some_folder
  ```
- Search in a specific file:
  ```sh
  ./mygrep -E "pattern" file.txt
  ```

## 4. Regular Expression Engine

The custom regex engine supports:

- **Anchors**: `^` (start of line), `$` (end of line)
- **Quantifiers**: `+` (one or more), `?` (zero or one)
- **Groups**: Parentheses for capturing groups, e.g., `(abc)`
- **Alternation**: `|` for top-level alternation, e.g., `foo|bar`
- **Character Classes**: `[abc]`, `[^abc]`
- **Escapes**: `\d` (digit), `\w` (word character), and backreferences (`\1`, `\2`, ...)
- **Dot**: `.` matches any character

### Implementation Highlights

- **Parsing**: The engine parses the pattern into atoms, handling groups, character classes, and alternation.
- **Matching**: Recursive matching functions handle quantifiers, groups, and backtracking.
- **Group Indexing**: Tracks group positions for backreferences.
- **Alternation**: Splits patterns at top-level `|` for alternation logic.

### Limitations

- Does not support all PCRE features (e.g., `{n,m}` quantifiers, lookahead/lookbehind).
- Character classes are basic and do not support ranges (e.g., `[a-z]`).
- Backreferences are limited to single-digit groups.

## 5. File and Directory Traversal

- Uses Go's `filepath.WalkDir` for recursive directory traversal when `-r` is specified.
- For each file, reads line by line and applies the regex engine.
- Supports multiple files and prints the filename as a prefix when searching more than one file or recursively.
- If no path is provided, reads from standard input.

## 6. Error Handling

- Invalid patterns or file errors print a message to `stderr` and exit with code 2.
- If no matches are found, exits with code 1.
- Successful matches exit with code 0.

## 7. Extending the Project

To add features or improve the regex engine:

- Enhance the parser in `parser.go` to support more regex syntax (e.g., ranges, more escapes).
- Add new quantifiers or support for non-greedy matching.
- Improve error messages and diagnostics.
- Add unit tests for the regex engine and CLI behavior.

## 8. References

- [CodeCrafters: Build Your Own grep](https://codecrafters.io)
- [Go Documentation](https://golang.org/doc/)
- [Regular Expressions - Wikipedia](https://en.wikipedia.org/wiki/Regular_expression)

---

For questions or contributions, feel free to open an issue or pull request.
