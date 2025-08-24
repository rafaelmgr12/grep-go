

# mygrep

mygrep is an open source, minimal grep clone written in Go, featuring a custom regular expression engine and robust file searching capabilities. The project is designed for learning, extensibility, and community collaboration.

## Table of Contents

- [mygrep](#mygrep)
  - [Table of Contents](#table-of-contents)
  - [Features](#features)
  - [Project Structure](#project-structure)
  - [Building](#building)
  - [Usage](#usage)
    - [Examples](#examples)
  - [Documentation](#documentation)
  - [Contributing](#contributing)
  - [License](#license)
  - [Credits](#credits)

## Features

- Recursive directory search (`-r`)
- Custom regex engine: groups, alternation, quantifiers (+, ?), character classes, anchors (^, $), escapes (\d, \w, etc.)
- Multiple file support
- Standard input support
- Extensible and well-documented codebase

## Project Structure

- `main.go`: CLI entry point and orchestration
- `search.go`: Argument parsing and file search logic
- `re.go`: Regular expression engine implementation
- `parser.go`: Regex pattern parsing utilities
- `state.go`: Regex matching state (if present)
- `go.mod`, `go.sum`: Go module files
- `docs/overview.md`: Extensive technical documentation

## Building

```sh
go build -o mygrep .
```

## Usage

```sh
./mygrep [-r] -E <pattern> [path ...]
```

- Use `-r` to search directories recursively
- If no path is provided, input is read from standard input
- Pattern must be provided with `-E`

### Examples

```sh
# Search for lines containing "hello" in standard input
echo -e "hello\nworld" | ./mygrep -E "hello"

# Search recursively for lines matching a pattern in all files under a directory
./mygrep -r -E "pattern" ./some_folder

# Search in a specific file
./mygrep -E "pattern" file.txt
```

## Documentation

Extensive documentation is available in [`docs/overview.md`](docs/overview.md). Please refer to it for:

- Design and implementation details
- Supported regex features and limitations
- File traversal and error handling
- Extension guidelines

## Contributing

We welcome contributions from the community! To get started:

1. Read [`docs/overview.md`](docs/overview.md) for technical details.
2. Fork the repository and create your feature branch.
3. Submit a pull request with a clear description of your changes.
4. For questions or suggestions, open an issue.

## License

This project is licensed under the MIT License. See the LICENSE file for details.

## Credits

Inspired by the "Build Your Own grep" challenge on [CodeCrafters](https://codecrafters.io).
