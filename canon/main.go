// Copyright 2013 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// The canon command adds canonical import paths to Go packages.
package main // import "willnorris.com/go/tools/canon"

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"go/ast"
	"go/build"
	"go/format"
	"go/parser"
	"go/token"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
)

// flags
var (
	dryrun = flag.Bool("n", false, "dry run: show changes, but don't apply them")
)

func usage() {
	fmt.Fprint(os.Stderr, "usage: canon [packages]\n")
	flag.PrintDefaults()
}

func main() {
	flag.Usage = usage
	flag.Parse()

	if len(flag.Args()) == 0 {
		flag.Usage()
		return
	}

	if err := fixPackages(flag.Args()...); err != nil {
		log.Fatalf("error listing packages: %v", err)
	}
}

func fixPackages(packages ...string) error {
	pkgs, err := list(packages...)
	if err != nil {
		return err
	}

	for _, pkg := range pkgs {
		if len(pkg.GoFiles) == 0 {
			// skip packages with no go files
			continue
		}
		if pkg.ImportComment != "" {
			if pkg.ImportComment != pkg.ImportPath {
				return fmt.Errorf("package %q does not having matching import comment %q", pkg.ImportPath, pkg.ImportComment)
			}
			continue
		}
		if err := fixPackage(pkg); err != nil {
			return err
		}
	}

	return nil
}

func fixPackage(pkg *build.Package) error {
	for _, file := range pkg.GoFiles {
		filename := filepath.Join(pkg.Dir, file)

		fset := token.NewFileSet()
		pf, err := parser.ParseFile(fset, filename, nil, parser.ParseComments)
		if err != nil {
			return err
		}

		if pf.Doc != nil {
			return rewriteFile(fset, pf, filename, pkg.ImportPath)
		}
	}

	// no files have package docs.  look for file that matches pkg.Name
	for _, file := range pkg.GoFiles {
		if file == pkg.Name+".go" {
			return parseAndRewriteFile(file, pkg)
		}
	}

	log.Printf("can't find file to rewrite for package: %q (%v)", pkg.ImportPath, pkg.Name)
	return nil
}

func parseAndRewriteFile(file string, pkg *build.Package) error {
	filename := filepath.Join(pkg.Dir, file)
	fset := token.NewFileSet()
	pf, err := parser.ParseFile(fset, filename, nil, parser.ParseComments)
	if err != nil {
		return err
	}

	return rewriteFile(fset, pf, filename, pkg.ImportPath)
}

// rewrite filename to include importPath.
func rewriteFile(fset *token.FileSet, pf *ast.File, filename, importPath string) error {
	log.Printf("package: %q, rewriting %q", importPath, filename)

	// add comment containing canonical import path
	cmap := ast.NewCommentMap(fset, pf, pf.Comments)
	if cmap == nil {
		cmap = make(ast.CommentMap)
	}
	com := &ast.Comment{Slash: pf.Name.End(), Text: `// import "` + importPath + `"`}
	cmap[pf.Name] = []*ast.CommentGroup{{List: []*ast.Comment{com}}}
	pf.Comments = cmap.Comments()

	if !*dryrun {
		var buf bytes.Buffer
		if err := format.Node(&buf, fset, pf); err != nil {
			return err
		}
		return ioutil.WriteFile(filename, buf.Bytes(), 0644)
	}

	return nil
}

// list runs 'go list' with the specified arguments and returns the metadata
// for matching packages.
func list(args ...string) ([]*build.Package, error) {
	cmd := exec.Command("go", append([]string{"list", "-e", "-json"}, args...)...)
	cmd.Stdout = new(bytes.Buffer)
	if err := cmd.Run(); err != nil {
		return nil, err
	}

	dec := json.NewDecoder(cmd.Stdout.(io.Reader))
	var pkgs []*build.Package
	for {
		var p build.Package
		if err := dec.Decode(&p); err == io.EOF {
			break
		} else if err != nil {
			return nil, err
		}
		pkgs = append(pkgs, &p)
	}
	return pkgs, nil
}
