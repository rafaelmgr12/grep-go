
# mygrep

This repository contains an educational implementation of a minimal `grep` clone written in Go. The project demonstrates a custom regular expression engine and file searching logic, inspired by the [CodeCrafters](https://codecrafters.io) "Build Your Own grep" challenge.

## Project Structure

- `main.go`: Command-line interface, file and directory traversal, input handling.
- `re.go`: Custom regular expression engine (parsing, matching, quantifiers, groups, alternation).
- `parser.go`: Pattern parsing utilities (group and alternation handling).
- `state.go`: (if present) likely contains state/environment logic for regex matching.
- `go.mod`, `go.sum`: Go module files.

## Building

```sh
go build -o mygrep .
```

## Usage

```sh
./mygrep [-r] -E <pattern> [path ...]
```

- Pass `-r` to search directories recursively (searches all files under the given path).
- If no path is provided, input is read from standard input (one line at a time).
- Pattern must be provided with `-E` (basic regex syntax, see implementation for supported features).

### Examples

Search for lines containing "hello" in standard input:

```sh
echo -e "hello\nworld" | ./mygrep -E "hello"
```

Search recursively for lines matching a pattern in all files under a directory:

```sh
./mygrep -r -E "pattern" ./some_folder
```

Search in a specific file:

```sh
./mygrep -E "pattern" file.txt
```

## Features

- Recursive search (`-r`)
- Custom regex engine: supports groups, alternation, quantifiers (+, ?), character classes, anchors (^, $), and some escapes (\d, \w, etc.)
- Multiple file support
- Standard input support

## Development

You can use the provided build command or run directly with `go run`:

```sh
go run . -E "pattern" file.txt
```

## Credits

Inspired by the "Build Your Own grep" challenge on [CodeCrafters](https://codecrafters.io).
