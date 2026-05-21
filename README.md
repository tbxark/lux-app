# lux-app

`lux-app` is a desktop downloader built with [Fyne](https://fyne.io/) and [lux](https://github.com/iawia002/lux). It provides a simple graphical interface for downloading videos, audio, captions, and playlists.

## Features

- Enter multiple URLs at once.
- Choose an output folder.
- Configure stream format, Cookie, User-Agent, and Referer.
- Support playlist downloads, audio-only downloads, captions, and multi-threaded downloads.
- Save commonly used settings automatically.

## Requirements

- Go 1.26 or later
- A local build environment with CGO support

`zig` is recommended for cross-compiling Windows or Linux builds.

## Development

```sh
go run .
```

## Build

Build for the current platform:

```sh
make build
```

Run tests:

```sh
make test
```

Build release artifacts:

```sh
make dist
```

Build all supported platforms:

```sh
make dist-all
```

## License

This project is licensed under the MIT License. See [LICENSE](LICENSE) for details.
