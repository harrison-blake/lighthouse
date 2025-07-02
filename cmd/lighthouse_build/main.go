package main

import (
	"os"
	"log"
	"strings"
	"net/http"
	"path/filepath"
	"html/template"
	"bytes"

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

func main() {
	err := createOutputDirs("./public")
	if err != nil {
		log.Print(err)
	}

	// create hard link to main layout
	err = os.Link("./content/stylesheets/base.css", "./public/stylesheets/base.css")
	if err != nil {
		log.Print(err)
	}
	// markdown parser extensions
	extensions := parser.CommonExtensions | parser.AutoHeadingIDs | parser.NoEmptyLineBeforeBlock

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

	for _, dir := range outputDirs {
		if dir == "stylesheets" {
			continue
		}
		files, err := os.ReadDir(filepath.Join(contentDir, dir))
		if err != nil {
			log.Fatal(err)
		}

		for _, file := range files {
			// can't resuse parser so create 1 for each loop
			p := parser.NewWithExtensions(extensions)

			err := processFiles(file, dir, p, renderer, tmpl)
			if err != nil {
				log.Fatal(err)
			}
		}
	}

	if localhost == true {
		http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			switch r.URL.Path {
			case "/":
				http.ServeFile(w, r, "./public/pages/index.html")
			case "/index.html":
				http.ServeFile(w, r, "./public/pages/index.html")
			case "/now.html":
				http.ServeFile(w, r, "./public/pages/now.html")
			case "/about.html":
				http.ServeFile(w, r, "./public/pages/about.html")
			case "/bits/":
				http.ServeFile(w, r, "./public/bits/index.html")
			case "/bits/index.html":
				http.ServeFile(w, r, "./public/bits/index.html")
			case "/stylesheets/base.css":
				w.Header().Set("Content-Type", "text/css")
				http.ServeFile(w, r, "./public/stylesheets/base.css")
			default:
				http.NotFound(w, r)
			}
		})

		log.Println("Serving on http://localhost:8080")
		log.Fatal(http.ListenAndServe(":8080", nil))
	}
}


// ************************************************************
// global variables go brrrrrrrrrr
// ************************************************************
var localhost = true
var contentDir = "./content"
var outputDir = "./public"
var outputDirs = []string{
	"bits",
	"pages",
	"stylesheets",
}

// ************************************************************
// types
// ************************************************************
type Content struct {
	Title string
	Body template.HTML
}

// ************************************************************
// functions
// ************************************************************

func createOutputDirs(targetDir string) error {
	for _, dir := range outputDirs {
		dir = filepath.Join(targetDir, dir)
		err := os.MkdirAll(dir, 0755)
		if err != nil {
			return err
		}
	}
	return nil
}

// read file
// parse file using the parser you created above
// feed parsed markdown into html renderer and output html
func processFiles(file os.DirEntry, dir string, p *parser.Parser, renderer *html.Renderer, tmpl *template.Template) error {
	filePath := filepath.Join(contentDir, dir, file.Name())
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

	newPath := filepath.Join(outputDir, dir, file.Name())
	newPath = strings.TrimSuffix(newPath, ".md") + ".html"

	err = os.WriteFile(newPath, buf.Bytes(), 0660)
	if err != nil {
		return err
	}

	return err
}
