# mygrep

This repository contains an educational implementation of a minimal `grep` clone written in Go.

The code started as part of the [CodeCrafters](https://codecrafters.io) "Build Your Own grep" challenge, but the implementation and documentation have been adapted to fit my own style.

## Building

```sh
go build -o mygrep ./app
```

## Usage

```sh
./mygrep [-r] -E <pattern> [path ...]
```

Pass `-r` to search directories recursively. When no path is provided, input is read from standard input.

Example:

```sh
echo "hello\nworld" | ./mygrep -E "hello"
```

## Development

Use `./your_program.sh` to compile and run the project in a way that's compatible with the CodeCrafters interface.

```sh
./your_program.sh -E "pattern" file.txt
```

## Credits

Inspired by the "Build Your Own grep" challenge on [CodeCrafters](https://codecrafters.io).
