// strip_extra_imports removes duplicate import blocks from a merged test file and runs gofmt.
// Usage: go run strip_extra_imports.go <file.go>
package main

import (
	"bufio"
	"fmt"
	"go/format"
	"os"
	"strings"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintln(os.Stderr, "usage: strip_extra_imports <file.go>")
		os.Exit(1)
	}
	path := os.Args[1]
	f, err := os.Open(path)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	defer f.Close()

	var out []string
	sc := bufio.NewScanner(f)
	firstImportDone := false
	skipImport := false
	paren := 0

	for sc.Scan() {
		line := sc.Text()
		trim := strings.TrimSpace(line)

		if skipImport {
			paren += strings.Count(line, "(") - strings.Count(line, ")")
			if paren <= 0 {
				skipImport = false
			}
			continue
		}

		if strings.HasPrefix(trim, "import ") {
			if firstImportDone {
				skipImport = true
				if strings.Contains(trim, "(") {
					paren = strings.Count(trim, "(") - strings.Count(trim, ")")
					if paren <= 0 {
						skipImport = false
					}
				} else {
					// single-line import
					continue
				}
				continue
			}
			if strings.Contains(trim, "(") {
				paren = strings.Count(trim, "(") - strings.Count(trim, ")")
			}
		}

		out = append(out, line)
		if !firstImportDone && strings.HasPrefix(trim, "import ") {
			if strings.Contains(trim, "(") {
				if paren <= 0 && strings.HasSuffix(trim, ")") {
					firstImportDone = true
				}
			} else {
				firstImportDone = true
			}
		} else if !firstImportDone && paren > 0 {
			paren += strings.Count(line, "(") - strings.Count(line, ")")
			if paren <= 0 {
				firstImportDone = true
			}
		}
	}

	formatted, err := format.Source([]byte(strings.Join(out, "\n")))
	if err != nil {
		fmt.Fprintf(os.Stderr, "format %s: %v\n", path, err)
		formatted = []byte(strings.Join(out, "\n"))
	}
	if err := os.WriteFile(path, formatted, 0o644); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
