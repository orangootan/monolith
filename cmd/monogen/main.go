package main

import (
	"go/parser"
	"go/token"
	"log"
	"os"
	"path"
	"strings"
)

func main() {
	set := token.NewFileSet()
	var files []File
	var out []string
	for _, name := range os.Args[1:] {
		f, _ := parser.ParseFile(set, name, nil, parser.ParseComments)
		file := parseFile(f).filter()
		files = append(files, file)
		ext := path.Ext(name)
		base := strings.TrimSuffix(name, ext)
		out = append(out, base+".g"+ext)
	}
	sm := createServiceMap(files)
	for i := 0; i < len(files); i++ {
		g := generateFile(files[i], sm)
		f, err := os.Create(out[i])
		if err != nil {
			log.Fatal(err)
		}
		err = g.Render(f)
		if err != nil {
			log.Fatal(err)
		}
	}
}
