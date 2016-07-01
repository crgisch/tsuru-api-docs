// Copyright 2016 tsuru authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"bytes"
	"flag"
	"fmt"
	"go/ast"
	"go/parser"
	"go/printer"
	"regexp"
	"strings"

	"golang.org/x/tools/go/loader"
	"gopkg.in/yaml.v2"
)

var (
	searchFlag   = flag.String("search", "", "return handlers matching search regexp")
	noSearchFlag = flag.String("no-search", "", "return handlers NOT matching search regexp")
	methodFlag   = flag.String("method", "", "return handlers with method")
	noMethodFlag = flag.String("no-method", "", "return handlers EXCEPT with method")
)

func apiLoader() (*loader.Program, error) {
	var ldr loader.Config
	ldr.ParserMode = parser.ParseComments
	ldr.Import("github.com/tsuru/tsuru/api")
	ldr.Import("github.com/tsuru/tsuru/provision/docker")
	return ldr.Load()
}

func isHandler(object *ast.Object) bool {
	params := object.Decl.(*ast.FuncDecl).Type.Params.List
	if len(params) < 2 {
		return false
	}
	for _, param := range params {
		t, ok := param.Type.(*ast.SelectorExpr)
		if ok {
			if t.Sel.Name == "ResponseWriter" {
				return true
			}
		}
	}
	return false
}

func shouldBeIgnored(objectName string) bool {
	ignoreList := []string{
		"writeEnvVars", "fullHealthcheck", "setVersionHeadersMiddleware",
		"authTokenMiddleware", "runDelayedHandler", "errorHandlingMiddleware",
		"contextClearerMiddleware", "flushingWriterMiddleware", "setRequestIDHeaderMiddleware",
		"bsEnvSetHandler", "bsConfigGetHandler", "bsUpgradeHandler",
	}
	for _, name := range ignoreList {
		if name == objectName {
			return true
		}
	}
	return false
}

func parse(prog *loader.Program) error {
	files := []*ast.File{}
	for _, f := range prog.Imported["github.com/tsuru/tsuru/api"].Files {
		files = append(files, f)
	}
	for _, f := range prog.Imported["github.com/tsuru/tsuru/provision/docker"].Files {
		files = append(files, f)
	}
	if *searchFlag == "" && *methodFlag == "" && *noSearchFlag == "" && *noMethodFlag == "" {
		fmt.Println("handlers:")
	}
	for _, f := range files {
		for _, object := range f.Scope.Objects {
			if object.Kind == ast.Fun {
				ok := isHandler(object)
				if !ok {
					continue
				}
				if shouldBeIgnored(object.Name) {
					continue
				}
				commentGroup := object.Decl.(*ast.FuncDecl).Doc
				if commentGroup == nil {
					fmt.Printf("missing docs for %s\n", object.Name)
					continue
				}
				err := handleComments(prog, object, commentGroup)
				if err != nil {
					fmt.Printf("error handling comments for %s: %s\n", object.Name, err)
				}
			}
		}
	}
	return nil
}

func handleComments(prog *loader.Program, object *ast.Object, commentGroup *ast.CommentGroup) error {
	if *searchFlag == "" && *methodFlag == "" && *noSearchFlag == "" && *noMethodFlag == "" {
		for _, comment := range commentGroup.List {
			if strings.Contains(comment.Text, "title:") {
				fmt.Println(strings.Replace(comment.Text, "// ", "  - ", -1))
			} else {
				fmt.Println(strings.Replace(comment.Text, "// ", "    ", -1))
			}
		}
		return nil
	}
	var buf bytes.Buffer
	for _, comment := range commentGroup.List {
		buf.WriteString(strings.Replace(comment.Text, "// ", "", -1))
		buf.WriteString("\n")
	}
	var parsed map[string]interface{}
	err := yaml.Unmarshal(buf.Bytes(), &parsed)
	if err != nil {
		return err
	}
	method, ok := parsed["method"].(string)
	if !ok {
		return fmt.Errorf("invalid method declaration for %s: %v", object.Name, parsed["method"])
	}
	if *methodFlag != "" {
		if strings.ToLower(*methodFlag) != strings.ToLower(method) {
			return nil
		}
	}
	if *noMethodFlag != "" {
		if strings.ToLower(*noMethodFlag) == strings.ToLower(method) {
			return nil
		}
	}
	if *searchFlag != "" || *noSearchFlag != "" {
		value := *searchFlag
		negate := true
		if value == "" {
			value = *noSearchFlag
			negate = false
		}
		var funcBuf bytes.Buffer
		printer.Fprint(&funcBuf, prog.Fset, object.Decl)
		re, err := regexp.Compile(value)
		if err != nil {
			return err
		}
		isMatch := re.Match(funcBuf.Bytes())
		if negate {
			isMatch = !isMatch
		}
		if isMatch {
			return nil
		}
	}
	fmt.Println(object.Name)
	return nil
}

func main() {
	flag.Parse()
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
