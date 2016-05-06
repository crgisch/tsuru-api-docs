// Copyright 2016 tsuru authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"fmt"
	"go/ast"
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
	files := prog.Imported["github.com/tsuru/tsuru/api"].Files
	for _, f := range files {
		for _, object := range f.Scope.Objects {
			if object.Kind == ast.Fun && object.Name == "serviceList" {
				fmt.Printf("handler: %s - %#v\n", object.Name, object.Decl.(*ast.FuncDecl).Doc)
				commentGroup := object.Decl.(*ast.FuncDecl).Doc
				for _, comment := range commentGroup.List {
					fmt.Println(comment.Text)
				}
			}
		}
	}
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
