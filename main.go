// Copyright 2016 tsuru authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"fmt"
	"go/parser"

	"golang.org/x/tools/go/loader"
)

func apiLoader() (*loader.Program, error) {
	var ldr loader.Config
	ldr.ParserMode = parser.ParseComments
	ldr.Import("github.com/tsuru/tsuru/api")
	return ldr.Load()
}

func parse(prog *loader.Program) error {
	fmt.Println(prog)
	return nil
}

func main() {
	prog, err := apiLoader()
	if err != nil {
		fmt.Println("error loading code ", err)
		return
	}
	err = parse(prog)
	if err != nil {
		fmt.Println("error parsing api ", err)
		return
	}
}
