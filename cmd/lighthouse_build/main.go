package main

import (
	"fmt"
	"os"
	"log"
	"strings"
	"net/http"

	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/html"
	"github.com/gomarkdown/markdown/parser"
)
// most of the code is pulled from from https://github.com/gomarkdown/markdown?tab=readme-ov-file
// TODO:
//	- write my own MarkDown parser
func main() {
	// folder name for markdown files
	contentDir := "./content"
	// folder name for html files
	viewsDir := "./views"
	htmlExt := ".html"


	// read all markdown files
	files, err := os.ReadDir(contentDir)
	if err != nil {
		log.Fatal(err)
	}

	// create directory if 
	error := os.MkdirAll(viewsDir, 0750)
	if error != nil && !os.IsExist(err) {
		log.Fatal(err)
	}
	
	// create markdown parser using most common extensions
	extensions := parser.CommonExtensions | parser.AutoHeadingIDs | parser.NoEmptyLineBeforeBlock
	p := parser.NewWithExtensions(extensions)

	// create html renderer using most common extensions
	htmlFlags := html.CommonFlags | html.HrefTargetBlank
	opts := html.RendererOptions{Flags: htmlFlags}
	renderer := html.NewRenderer(opts)

	// read each file into raw []bytes
	// parse markdown into AST using the parser you created above
	// feed parsed markdown into html renderer and output html string
	count := 0
	for _, file := range files {
		readPath := contentDir + "/" + file.Name()

		content, err := os.ReadFile(readPath)
		if err != nil {
			log.Fatal(err)
		}

		doc := p.Parse(content)
		output := markdown.Render(doc, renderer)
		fmt.Printf("file name: %q\n", readPath)
		fmt.Printf("HTML:\n%s\n", output)

		writePath := viewsDir + "/" + strings.Split(file.Name(), ".")[0] + htmlExt
		err = os.WriteFile(writePath, output, 0660)
		if err != nil {
			log.Fatal(err)
		}
		count++
	}

	fmt.Printf("number of files parsed: %d\n", count)

	// serve http
	http.Handle("/", http.FileServer(http.Dir("./views")))
	fmt.Println("Serving on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
