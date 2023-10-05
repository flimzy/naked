package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
)

func main() {
	count, mixed, err := scanFile(os.Args[1])
	if err != nil {
		panic(err)
	}
	if count > 0 {
		fmt.Printf("%s: %d naked returns (%d mixed functions)\n", os.Args[1], count, mixed)
	}
}

func scanFile(filename string) (naked int, mixed int, _ error) {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, filename, nil, parser.AllErrors)
	if err != nil {
		return 0, 0, err
	}
	v := &visitor{fset: fset}
	ast.Walk(v, f)
	return v.found, v.mixed, nil
}

func foo() (err error) {
	return
}

func bar() {}

func baz() error {
	return func() (err error) {
		if false {
			return nil
		}
		return
	}()
}

type frame struct {
	hasReturns     bool
	clothed, naked bool
}

type visitor struct {
	fset  *token.FileSet
	stack []*frame
	found int
	mixed int
}

func (v *visitor) Visit(node ast.Node) ast.Visitor {
	if node == nil {
		if l := len(v.stack); l > 0 {
			var p *frame
			p, v.stack = v.stack[l-1], v.stack[:l-1]
			if p != nil && p.clothed && p.naked {
				v.mixed++
			}
		}
		return nil
	}
	switch t := node.(type) {
	case *ast.FuncDecl:
		hasReturns := t.Type.Results != nil
		v.stack = append(v.stack, &frame{hasReturns: hasReturns})
	case *ast.ReturnStmt:
		p := v.parent()
		if p.hasReturns {
			if t.Results == nil {
				v.found++
				p.naked = true
			} else {
				p.clothed = true
			}
		}
	default:
		v.stack = append(v.stack, nil)
	}
	return v
}

func (v *visitor) parent() *frame {
	for i := len(v.stack) - 1; i >= 0; i-- {
		if v.stack[i] != nil {
			return v.stack[i]
		}
	}
	return nil
}
