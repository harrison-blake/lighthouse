package main

import (
	// "fmt"
	"os"
	"log"
	"strings"
	// "net/http"
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
// - https://github.com/brandur/sorg/tree/master
// ************************************************************
// TODO:
//	- [] refactor main loop / create function that looks for differences
//		 and rebuilds 
//  - [] split up spinning up local server and building site
//	- [] change up home page content
//  - [] add cd/ci
// 	- [] add tailwind
//  - [] write my own markdown parser
// ************************************************************

func main() {
	// make sure layouts directory is created so hardlink doesn't fail
	err := os.MkdirAll("public/layouts", 0755)
	if err != nil {
		log.Print(err)
	}

	// create hard link to main layout
	err = os.Link("./content/layouts/style.css", "./public/layouts/style.css")
	if err != nil {
		log.Print(err)
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

	// read each file into raw []bytes
	// parse markdown files using the parser you created above
	// feed parsed markdown into html renderer and output html
	for _, file := range files {
		err := processFiles(file, p, renderer, tmpl)
		if err != nil {
			log.Fatal(err)
		}

	}
	log.Println("Build was successful")

	// http.Handle("/", http.FileServer(http.Dir("./public")))
	// fmt.Println("Serving on http://localhost:8080")
	// log.Fatal(http.ListenAndServe(":8080", nil))
}


// ************************************************************
// lobal variables go brrrrrrrrrr
// ************************************************************
const (
	contentDir = "./content"
	targetDir = "./public"
)


// ************************************************************
// types
// ************************************************************
type Content struct {
	Title string
	Body template.HTML
}

// ************************************************************
// types
// ************************************************************
func processFiles(file os.DirEntry, p *parser.Parser, renderer *html.Renderer, tmpl *template.Template) error {
	filePath := filepath.Join(contentDir, "pages", file.Name())
	fileContent, err := os.ReadFile(filePath)

	if err != nil {
		return err
	}

	doc := p.Parse(fileContent)
	rendered := markdown.Render(doc, renderer)

	page := Content{
		Title: strings.TrimSuffix(file.Name(), ".md"),
		Body: template.HTML(rendered)}

	var buf bytes.Buffer
	err = tmpl.Execute(&buf, page)
	if err != nil {
		return err
	}

	newPath := filepath.Join(targetDir, file.Name())
	newPath = strings.TrimSuffix(newPath, ".md") + ".html"

	err = os.WriteFile(newPath, buf.Bytes(), 0660)
	if err != nil {
		return err
	}

	return err
}
