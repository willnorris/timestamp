# canon

canon is a tool to add [canonical import paths][] to Go packages.

[canonical import paths]: https://golang.org/doc/go1.4#canonicalimports

For example, given a file located at `$GOPATH/example.com/foo/foo.go` with the
following contents:

```go
// Package foo docs here.
package foo
```

canon will update this file as:

```go
// Package foo docs here.
package foo // import "example.com/foo"
```

canon will only modify a single go source file per package.  If there are
multiple source files for the package, canon will try to use the file that
declares the package-level documentation.  If there is no package documentation,
the behavior of selecting a file is undefined.
