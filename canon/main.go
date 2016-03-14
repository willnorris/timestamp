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
	"go/build"
	"io"
	"log"
	"os"
	"os/exec"
	"strings"
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
		if strings.Contains(pkg.ImportPath, "/vendor/") || strings.Contains(pkg.ImportPath, "/third_party/") {
			// skip vendored packages
			continue
		}
		if pkg.ImportComment != "" {
			if pkg.ImportComment != pkg.ImportPath {
				return fmt.Errorf("package %q does not having matching import comment %q", pkg.ImportPath, pkg.ImportComment)
			}
			// skip packages with canonical import path
			continue
		}
		if err := fixPackage(pkg); err != nil {
			return err
		}
	}

	return nil
}

func fixPackage(pkg *build.Package) error {
	// TODO(willnorris): fix the package
	fmt.Println(pkg.ImportPath)
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