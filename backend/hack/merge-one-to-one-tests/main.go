// merge-one-to-one-tests concatenates extra *_test.go bodies into the primary test file (1:1 cleanup).
// Usage: go run ./hack/merge-one-to-one-tests <testDir> <primaryFile e.g. service_test.go>
package main

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

var skipFuncs = map[string]bool{
	"codeOf": true, "registerInput": true, "testSupportRegisterInput": true,
	"newInviteService": true, "newInviteServiceNoWebBase": true, "testPassword": true,
	"validPassword": true, "modelsLogin": true,
}

func main() {
	if len(os.Args) != 3 {
		fmt.Fprintf(os.Stderr, "usage: merge-one-to-one-tests <testDir> <primary>\n")
		os.Exit(1)
	}
	testDir := os.Args[1]
	primary := os.Args[2]

	primaryPath := filepath.Join(testDir, primary)
	base, err := os.ReadFile(primaryPath)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	var extras []string
	entries, _ := os.ReadDir(testDir)
	for _, e := range entries {
		name := e.Name()
		if e.IsDir() || !strings.HasSuffix(name, "_test.go") || name == primary {
			continue
		}
		extras = append(extras, name)
	}
	sort.Strings(extras)

	var b strings.Builder
	b.Write(base)
	if !bytes.HasSuffix(base, []byte("\n")) {
		b.WriteByte('\n')
	}

	fset := token.NewFileSet()
	for _, name := range extras {
		path := filepath.Join(testDir, name)
		data, err := os.ReadFile(path)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		f, err := parser.ParseFile(fset, path, data, parser.ParseComments)
		if err != nil {
			fmt.Fprintf(os.Stderr, "parse %s: %v\n", name, err)
			os.Exit(1)
		}
		for _, decl := range f.Decls {
			if gen, ok := decl.(*ast.GenDecl); ok && gen.Tok == token.IMPORT {
				continue
			}
			buf := &bytes.Buffer{}
			if err := format.Node(buf, fset, decl); err != nil {
				continue
			}
			src := buf.String()
			if strings.HasPrefix(src, "func ") {
				fn := strings.TrimPrefix(src, "func ")
				if idx := strings.IndexByte(fn, '('); idx > 0 {
					fnName := fn[:idx]
					if skipFuncs[fnName] {
						continue
					}
				}
			}
			b.WriteString("\n")
			b.WriteString(src)
			if !strings.HasSuffix(src, "\n") {
				b.WriteByte('\n')
			}
		}
		fmt.Fprintf(os.Stderr, "merged %s\n", name)
	}

	out, err := format.Source([]byte(b.String()))
	if err != nil {
		fmt.Fprintf(os.Stderr, "format: %v\n", err)
		out = []byte(b.String())
	}
	if err := os.WriteFile(primaryPath, out, 0o644); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	for _, name := range extras {
		_ = os.Remove(filepath.Join(testDir, name))
	}
}
