// Copyright 2016 tsuru authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"go/ast"
	"go/printer"
	"regexp"
	"strings"

	"golang.org/x/tools/go/packages"
	"gopkg.in/yaml.v2"
)

var (
	searchFlag   = flag.String("search", "", "return handlers matching search regexp")
	noSearchFlag = flag.String("no-search", "", "return handlers NOT matching search regexp")
	methodFlag   = flag.String("method", "", "return handlers with method")
	noMethodFlag = flag.String("no-method", "", "return handlers EXCEPT with method")
)

func isListMode() bool {
	return *searchFlag == "" && *methodFlag == "" && *noSearchFlag == "" && *noMethodFlag == ""
}

func apiLoader() ([]*packages.Package, error) {
	cfg := &packages.Config{
		Mode: packages.NeedFiles | packages.NeedSyntax | packages.NeedTypes | packages.NeedImports | packages.NeedDeps | packages.NeedCompiledGoFiles,
	}
	pkgsStr := []string{
		"github.com/tsuru/tsuru/api",
		"github.com/tsuru/tsuru/provision/docker",
	}
	pkgs, err := packages.Load(cfg, pkgsStr...)
	if err != nil {
		return nil, err
	}
	if packages.PrintErrors(pkgs) > 0 {
		return nil, errors.New("errors reading pkgs")
	}
	return pkgs, err
}

func isHandler(object *ast.Object) bool {
	params := object.Decl.(*ast.FuncDecl).Type.Params.List
	if len(params) < 2 {
		return false
	}
	paramType, ok := params[0].Type.(*ast.SelectorExpr)
	if !ok || paramType.Sel.Name != "ResponseWriter" {
		return false
	}
	if len(params) == 3 {
		paramType, ok = params[2].Type.(*ast.SelectorExpr)
		if !ok || paramType.Sel.Name != "Token" {
			return false
		}
	}
	return true
}

func shouldBeIgnored(objectName string) bool {
	ignoreList := []string{
		"writeEnvVars", "fullHealthcheck", "setVersionHeadersMiddleware",
		"authTokenMiddleware", "runDelayedHandler", "errorHandlingMiddleware",
		"contextClearerMiddleware", "flushingWriterMiddleware", "setRequestIDHeaderMiddleware",
		"bsEnvSetHandler", "bsConfigGetHandler", "bsUpgradeHandler", "contentHijacker",
	}
	for _, name := range ignoreList {
		if name == objectName {
			return true
		}
	}
	return false
}

func parse(pkgs []*packages.Package) error {
	if isListMode() {
		fmt.Println("handlers:")
	}
	for _, pkg := range pkgs {
		err := parsePkg(pkg)
		if err != nil {
			return err
		}
	}
	return nil
}

func parsePkg(pkg *packages.Package) error {
	for i, _ := range pkg.CompiledGoFiles {
		fileAst := pkg.Syntax[i]
		for _, object := range fileAst.Scope.Objects {
			if object.Kind != ast.Fun {
				continue
			}
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
			err := handleComments(object, commentGroup, pkg)
			if err != nil {
				fmt.Printf("error handling comments for %s: %s\n", object.Name, err)
			}
		}
	}
	return nil
}

func handleComments(object *ast.Object, commentGroup *ast.CommentGroup, pkg *packages.Package) error {
	if isListMode() {
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
		text := comment.Text
		text = strings.Replace(text, "// ", "", -1)
		text = strings.Replace(text, "//\t", "\t", -1)
		text = strings.Replace(text, "//", "", -1)
		text = strings.Replace(text, "\t", "  ", -1)

		buf.WriteString(text)
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
		printer.Fprint(&funcBuf, pkg.Fset, object.Decl)
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
	fmt.Printf("%s.%s\n", pkg.String(), object.Name)
	return nil
}

func main() {
	flag.Parse()
	pkgs, err := apiLoader()
	if err != nil {
		fmt.Println("error loading code ", err)
		return
	}
	err = parse(pkgs)
	if err != nil {
		fmt.Println("error parsing api ", err)
		return
	}
}
