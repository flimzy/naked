package main

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"io"
	"os"
	"path/filepath"
	"slices"
	"strings"
)

func main() {
	s := bufio.NewScanner(os.Stdin)
	var total, naked, clothed, mixed int64
	for s.Scan() {
		filename := s.Text()
		t, n, c, m, err := countNaked(filename, nil)
		if err != nil {
			fmt.Printf("%s: %s\n", filename, err)
		}
		total += int64(t)
		naked += int64(n)
		clothed += int64(c)
		mixed += int64(m)
	}
	fmt.Printf("Found %d total returns (%d naked, %d clothed) and %d functions with both naked and clothed returns\n",
		total, naked, clothed, mixed)
}

type frame struct {
	hasReturns            bool
	total, clothed, naked bool
}

const (
	skipTrigger = "DO NOT EDIT"
	headerBytes = 1024
)

func shouldSkip(filename string, src any) (bool, error) {
	if slices.Contains(filepath.SplitList(filename), "testdata") {
		return true, nil
	}
	if src != nil {
		switch t := src.(type) {
		case string:
			return strings.Contains(t, skipTrigger), nil
		case []byte:
			return bytes.Contains(t, []byte(skipTrigger)), nil
		default:
			panic(fmt.Sprintf("not sure how to handle src of type %T", src))
		}
	}

	file, err := os.Open(filename)
	if err != nil {
		return false, err
	}
	content, err := io.ReadAll(io.LimitReader(file, headerBytes))
	if err != nil {
		_ = file.Close()
		return false, err
	}
	_ = file.Close()
	return bytes.Contains(content, []byte(skipTrigger)), nil
}

func countNaked(filename string, src any) (total, naked, clothed, mixed int, _ error) {
	skip, err := shouldSkip(filename, src)
	if err != nil {
		return 0, 0, 0, 0, err
	}
	if skip {
		return 0, 0, 0, 0, nil
	}
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, filename, src, parser.AllErrors)
	if err != nil {
		return 0, 0, 0, 0, fmt.Errorf("failed to parse: %w", err)
	}

	stack := make([]*frame, 0, 10)

	ast.Inspect(f, func(node ast.Node) bool {
		if node == nil {
			if p := stack[len(stack)-1]; p != nil && p.clothed && p.naked {
				mixed++
			}
			stack = stack[:len(stack)-1]
			return false
		}
		switch t := node.(type) {
		case *ast.FuncDecl:
			hasReturns := t.Type.Results != nil
			stack = append(stack, &frame{hasReturns: hasReturns})
			return true
		case *ast.FuncLit:
			hasReturns := t.Type.Results != nil
			stack = append(stack, &frame{hasReturns: hasReturns})
			return true
		case *ast.ReturnStmt:
			p, err := parent(stack)
			if err != nil {
				panic(fmt.Sprintf("%s: %s\n", filename, err))
			}
			total++
			if p.hasReturns {
				if t.Results == nil {
					naked++
					p.naked = true
				} else {
					clothed++
					p.clothed = true
				}
			}
		}
		stack = append(stack, nil)
		return true
	})
	return total, naked, clothed, mixed, nil
}

func parent(stack []*frame) (*frame, error) {
	for i := len(stack) - 1; i >= 0; i-- {
		if stack[i] != nil {
			return stack[i], nil
		}
	}
	return nil, errors.New("no parent")
}

var (
	_ = debugAST
	_ = debugCode
)

func debugAST(fset *token.FileSet, node ast.Node) {
	var buf bytes.Buffer
	ast.Fprint(&buf, fset, node, nil)
	fmt.Printf("%s\n", buf.String())
}

func debugCode(fset *token.FileSet, node ast.Node) {
	var buf bytes.Buffer
	printer.Fprint(&buf, fset, node)
	fmt.Printf("%s\n", buf.String())
}
