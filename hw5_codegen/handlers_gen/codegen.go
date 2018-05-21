package main

import (
	"fmt"
	"go/parser"
	"go/token"
	"log"
	"os"
)

func main() {
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, os.Args[1], nil, parser.ParseComments)
	if err != nil {
		log.Fatal(err)
	}

	out, _ := os.Create(os.Args[2])

	fmt.Fprintln(out, `package `+node.Name.Name)
	fmt.Fprintln(out) // empty line
	//fmt.Fprintln(out, `import "encoding/binary"`)
	//fmt.Fprintln(out, `import "bytes"`)
	//fmt.Fprintln(out) // empty line

}
