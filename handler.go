package main

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/printer"
	"go/token"
	"io"
	"log"
	"net"
	"os"
	"regexp"
	"sort"
	"strings"
)

type handler struct {
	filename                  string
	sdk, inner, outer, result []*ast.ImportSpec
	fset                      *token.FileSet
	file                      ast.Node

	firstDecl *ast.GenDecl
	all       []ast.Spec
}

func newHandler(filename string) *handler {
	return &handler{filename: filename}
}

func (h *handler) start() {
	h.fset = token.NewFileSet()

	node, err := parser.ParseFile(h.fset, h.filename, nil, parser.ParseComments)
	if err != nil {
		log.Fatalf("Failed to parse file: %v", err)
	}
	ast.Inspect(node, h.travel)

	h.firstDecl.Specs = h.sortimport(h.all)
	h.delDecl()
}

func (h *handler) travel(node ast.Node) (ok bool) {
	if file, ok := node.(*ast.File); ok && file != nil {
		h.file = file
		return true
	}
	if impDecl, ok := node.(*ast.GenDecl); ok && impDecl.Tok == token.IMPORT {
		if len(impDecl.Specs) > 0 {
			h.all = append(h.all, impDecl.Specs...)
			if h.firstDecl == nil {
				h.firstDecl = impDecl
			} else {
				impDecl.Specs = []ast.Spec{getEmptyImport()}
			}
			// impDecl.Specs = h.sortimport(impDecl.Specs)
			// return false
		}
	}
	return true
}

func (h *handler) delDecl() {
	root, ok := h.file.(*ast.File)
	if !ok {
		return
	}
	var decls []ast.Decl
	for _, v := range root.Decls {
		im, ok := v.(*ast.GenDecl)
		if !ok || im.Tok != token.IMPORT || h.firstDecl == v {
			decls = append(decls, v)
		}

	}
	root.Decls = decls
}

func (h *handler) sortimport(list []ast.Spec) (result []ast.Spec) {
	if len(list) <= 1 {
		result = list
		return
	}
	for _, v := range list {
		spec, ok := v.(*ast.ImportSpec)
		if !ok || spec == nil {
			continue
		}
		if isSdkPath(spec.Path.Value) {
			h.sdk = append(h.sdk, spec)
			continue
		}
		if isOuter(spec.Path.Value) {
			h.outer = append(h.outer, spec)
			continue
		}

		h.inner = append(h.inner, spec)
	}
	order(h.sdk)
	order(h.inner)
	order(h.outer)

	result = h.group()

	return
}

func (h *handler) group() (list []ast.Spec) {
	for _, v := range h.sdk {
		list = append(list, v)
	}
	if len(h.inner) > 0 {
		if len(list) > 0 {
			list = append(list, getEmptyImport())
		}

		for _, v := range h.inner {
			list = append(list, v)
		}

	}
	if len(h.outer) > 0 {
		if len(list) > 0 {
			list = append(list, getEmptyImport())
		}
		for _, v := range h.outer {
			list = append(list, v)
		}

	}
	return
}

// writeBack write back to the source file
func (h *handler) writeBack() (err error) {
	var writers []io.Writer
	if *stdOut {
		if *onlyChanged {
			h.printLine()
		} else {
			writers = append(writers, os.Stdout)
		}
	}

	if *writeback {
		f, err := os.OpenFile(h.filename, os.O_RDWR, 0o66)
		if err != nil {
			return err
		}
		defer f.Close()

		writers = append(writers, f)
	}

	if len(writers) == 0 {
		return
	}
	var buf bytes.Buffer
	printer.Fprint(&buf, h.fset, h.file)
	data, err := format.Source(buf.Bytes())
	if err != nil {
		return
	}

	w := io.MultiWriter(writers...)

	_, err = w.Write(data)

	return
}

func (h *handler) printLine() {
	var buf bytes.Buffer
	printer.Fprint(&buf, h.fset, h.firstDecl)
	data, err := format.Source(buf.Bytes())
	if err != nil {
		return
	}

	os.Stdout.Write(data)
}

func (h *handler) print() (err error) {
	var buf bytes.Buffer
	if err = printer.Fprint(&buf, h.fset, h.file); err != nil {
		return
	}
	fmt.Println(string(buf.Bytes()))
	return
}

func isSdkPath(str string) (yes bool) {
	str = strings.Trim(str, "\"")
	base := strings.Split(str, "/")[0]
	if base == "" {
		return
	}
	for _, v := range paths {
		if yes = base == v; yes {
			return
		}
	}
	return
}

func isOuter(path string) (yes bool) {
	path = strings.Trim(path, "\"")
	path = strings.Split(path, "/")[0]
	yes = isValidDomain(path)
	return
}

var domainRegex = regexp.MustCompile(`(?i)^([a-z0-9]+(-[a-z0-9]+)*\.)+[a-z]{2,}$`)

func isValidDomain(domain string) bool {
	return domainRegex.MatchString(domain)
}

func canResolveDomain(domain string) bool {
	ips, err := net.LookupIP(domain)
	if err != nil {
		// 域名解析错误
		return false
	}
	return len(ips) > 0
}

func order(list []*ast.ImportSpec) {
	if len(list) <= 1 {
		return
	}
	sort.Slice(list, func(i, j int) bool {
		return list[i].Path.Value < list[j].Path.Value
	})
}

func getEmptyImport() ast.Spec {
	return &ast.ImportSpec{
		Path: &ast.BasicLit{
			Kind:  token.STRING,
			Value: "\n",
		},
	}
}
