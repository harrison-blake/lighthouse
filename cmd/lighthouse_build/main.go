package main

import (
	"fmt"
	"os"
	"log"
	"strings"
	"net/http"
	"path/filepath"
	"html/template"
	"bytes"

	// "github.com/gomarkdown/markdown/ast"
	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/html"
	"github.com/gomarkdown/markdown/parser"
)
// ************************************************************
// REFERENCES (code directly yoinked or took inspiration from)
// - https://github.com/gomarkdown/markdown?tab=readme-ov-file
// - https://brandur.org/aws-intrinsic-static
// - https://github.com/brandur/singularity
// - https://github.com/brandur/sorg/tree/master
// ************************************************************
// TODO:
//	- write my own MarkDown parser
// 	- add tailwind
// ************************************************************

const (
	contentDir = "./content"
	targetDir = "./public"
)

func main() {
	// create directories where site is built too
	err := BuildTargetDirectory()
	if err != nil {
		log.Fatal(err)
	}

	// create hard link to main layout
	err = os.Link("./content/layouts/style.css", "./public/layouts/style.css")
	if err != nil {
		log.Fatal(err)
	}

	// create markdown parser
	extensions := parser.CommonExtensions | parser.AutoHeadingIDs | parser.NoEmptyLineBeforeBlock
	p := parser.NewWithExtensions(extensions)

	// create html renderer
	htmlFlags := html.CommonFlags | html.HrefTargetBlank
	opts := html.RendererOptions{Flags: htmlFlags}
	renderer := html.NewRenderer(opts)

	// create html template
	// allows for injection of title and body
	tmpl, err := template.ParseFiles("templates/base.html")
	if err != nil {
		log.Fatal(err)
	}

	// generate array of markdown files to be parsed
	files, err := os.ReadDir("./content/pages")
	if err != nil {
		log.Fatal(err)
	}

	// loop that transforms markdown into html files
	// ***************************************************************
	// read each file into raw []bytes
	// parse markdown files into AST using the parser you created above
	// feed parsed markdown into html renderer and output html
	count := 0
	for _, file := range files {
		filePath := filepath.Join(contentDir, "pages", file.Name())
		fileContent, err := os.ReadFile(filePath)

		if err != nil {
			log.Fatal(err)
		}

		doc := p.Parse(fileContent)
		rendered := markdown.Render(doc, renderer)
		fmt.Printf("file name: %q\n", filePath)

		page := Content{
			Title: strings.TrimSuffix(file.Name(), ".md"),
			Body: template.HTML(rendered)}

		var buf bytes.Buffer
		err = tmpl.Execute(&buf, page)
		if err != nil {
			log.Fatal(err)
		}

		newPath := filepath.Join(targetDir, "pages", file.Name())
		newPath = strings.TrimSuffix(newPath, ".md") + ".html"
		fmt.Printf("html file path: %q\n", newPath)

		err = os.WriteFile(newPath, buf.Bytes(), 0660)
		if err != nil {
			log.Fatal(err)
		}
		count++
	}

	fmt.Printf("number of files parsed: %d\n", count)

	http.Handle("/", http.FileServer(http.Dir("./public")))
	fmt.Println("Serving on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

type Content struct {
	Title string
	Body template.HTML
}

func BuildTargetDirectory() error {
	dirs, err := os.ReadDir(contentDir)
	if err != nil {
		return err
	}

	for _, item := range dirs {
		if item.IsDir() {
			filename := filepath.Join(targetDir, item.Name())
			os.MkdirAll(filename, 0755)
		}
	}

	return err
}
