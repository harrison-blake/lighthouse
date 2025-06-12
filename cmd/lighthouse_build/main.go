package main

import (
	"fmt"
	"os"
	"log"
	"strings"
	// "net/http"
	"path/filepath"

	//"github.com/gomarkdown/markdown/ast"
	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/html"
	"github.com/gomarkdown/markdown/parser"
)

type Content struct {
	Title string
	Body string
}

// most of the code is pulled from from https://github.com/gomarkdown/markdown?tab=readme-ov-file
// TODO:
//	- write my own MarkDown parser
func main() {
	// location for markdown content
	contentDir := "./content/markdown"

	// location of generated html files
	viewsDir := "./views"

	htmlExt := ".html"

	// read all markdown files
	files, err := os.ReadDir(contentDir)
	if err != nil {
		log.Fatal(err)
	}

	// create directory if 
	error := os.MkdirAll(viewsDir, 0750)
	if error != nil {
		log.Fatal(err)
	}
	
	// create markdown parser using common extensions
	extensions := parser.CommonExtensions | parser.AutoHeadingIDs | parser.NoEmptyLineBeforeBlock
	p := parser.NewWithExtensions(extensions)

	// create html renderer using common extensions
	htmlFlags := html.CommonFlags | html.HrefTargetBlank
	opts := html.RendererOptions{Flags: htmlFlags}
	renderer := html.NewRenderer(opts)

	// read each file into raw []bytes
	// parse markdown into AST using the parser you created above
	// feed parsed markdown into html renderer and output html
	count := 0
	for _, file := range files {
		readPath := filepath.Join(contentDir, file.Name())
		content, err := os.ReadFile(readPath)
		if err != nil {
			log.Fatal(err)
		}

		doc := p.Parse(content)
		output := markdown.Render(doc, renderer)
		fmt.Printf("file name: %q\n", readPath)

		path := filepath.Join(viewsDir, file.Name())
		writePath := strings.TrimSuffix(path, ".md") + htmlExt
		fmt.Printf("html file path: %q\n", writePath)

		err = os.WriteFile(writePath, output, 0660)
		if err != nil {
			log.Fatal(err)
		}
		count++
	}

	fmt.Printf("number of files parsed: %d\n", count)

	// serve http
	// http.Handle("/", http.FileServer(http.Dir("./views")))
	// fmt.Println("Serving on http://localhost:8080")
	// log.Fatal(http.ListenAndServe(":8080", nil))
}
