package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
)

func main() {
	count, mixed, err := countNaked(os.Args[1], nil)
	if err != nil {
		panic(err)
	}
	if count > 0 {
		fmt.Printf("%s: %d naked returns (%d mixed functions)\n", os.Args[1], count, mixed)
	}
}

type frame struct {
	hasReturns     bool
	clothed, naked bool
}

func countNaked(filename string, src any) (total, mixed int, _ error) {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, filename, src, parser.AllErrors)
	if err != nil {
		return 0, 0, err
	}

	stack := make([]*frame, 0, 10)

	ast.Inspect(f, func(node ast.Node) bool {
		if node == nil {
			if p := stack[len(stack)-1]; p != nil && p.clothed && p.naked {
				mixed++
			}
			stack = stack[:len(stack)-1]
			fmt.Println("decr", stack)
			return false
		}
		switch t := node.(type) {
		case *ast.FuncDecl:
			hasReturns := t.Type.Results != nil
			stack = append(stack, &frame{hasReturns: hasReturns})
			return true
		case *ast.ReturnStmt:
			p := parent(stack)
			if p.hasReturns {
				if t.Results == nil {
					total++
					p.naked = true
				} else {
					p.clothed = true
				}
			}
		}
		stack = append(stack, nil)
		fmt.Println("incr", stack)
		return true
	})
	return total, mixed, nil
}

func parent(stack []*frame) *frame {
	for i := len(stack) - 1; i >= 0; i-- {
		if stack[i] != nil {
			return stack[i]
		}
	}
	panic("no parent")
}
