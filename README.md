[![GoDoc](https://img.shields.io/static/v1?label=godoc&message=reference&color=blue)](https://pkg.go.dev/willnorris.com/go/tools)
[![Test Status](https://github.com/willnorris/tools/workflows/tests/badge.svg)](https://github.com/willnorris/tools/actions?query=workflow%3Atests)
[![Test Coverage](https://codecov.io/gh/willnorris/tools/branch/main/graph/badge.svg)](https://codecov.io/gh/willnorris/tools)

This repository contains assorted command line tools that aren't quite big
enough to justify their own repository.

## timestamp ##

The `timestamp` tool prints time in a variety of formats including unix
timestamp, RFC 3339, ordinal date, and epoch days.  Install by running:

    go get willnorris.com/go/tools/timestamp

Run `timestamp -help` for complete usage.

## License ##

These tools are copyright Google, but are not official Google products.
They are available under a [BSD License][].

[BSD License]: LICENSE
