package main

import (
	"os"
	"log"
	"strings"
	"net/http"
	"path/filepath"
	"html/template"
	"bytes"
	"fmt"
	"io"

	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/html"
	"github.com/gomarkdown/markdown/parser"
)

// ************************************************************
// REFERENCES (code directly yoinked or took inspiration from)
// - https://brandur.org/aws-intrinsic-static
// - https://github.com/brandur/sorg
// - https://github.com/gomarkdown/markdown?tab=readme-ov-file
// ************************************************************

func main() {
    serveLocalhost, err := ParseBool(os.Getenv("SERVE_LOCALHOST"))
    if err != nil {
		log.Println("SERVE_LOCALHOST not found")
	}
	log.Printf("serveLocalhost set to: %v", serveLocalhost)

	outputDirs := []string{
		"bits",
		"about",
		"now",
		"stylesheets",
	}

	count, err := CreateOutputDirs("./public", outputDirs)
	if err != nil {
		log.Print(err)
	}
	log.Printf("%v directories created", count)

	src, err := os.Open("./content/stylesheets/base.css")
	if err != nil {
		log.Fatal(err)
	}
	defer src.Close()

	dst, err := os.Create("./public/stylesheets/base.css")
	if err != nil {
		log.Fatal(err)
	}
	defer dst.Close()

	if _, err := io.Copy(dst, src); err != nil {
		log.Fatal(err)
	}
	// markdown parser extensions
	extensions := parser.CommonExtensions | parser.AutoHeadingIDs | parser.NoEmptyLineBeforeBlock

	// create html renderer
	htmlFlags := html.CommonFlags | html.HrefTargetBlank
	opts := html.RendererOptions{Flags: htmlFlags}
	renderer := html.NewRenderer(opts)

	// create html template
	// allows for injection of title and body
	baseTmpl, err := template.ParseFiles("templates/base.html")
	if err != nil {
		log.Fatal(err)
	}

	for _, dir := range outputDirs {
		if dir == "stylesheets" || dir == "home" {
			continue
		}
		files, err := os.ReadDir(filepath.Join(contentDir, dir))
		if err != nil {
			log.Fatal(err)
		}

		for _, file := range files {
			// can't resuse parser so create 1 for each loop
			p := parser.NewWithExtensions(extensions)

			err := ProcessFiles(file, dir, p, renderer, baseTmpl)
			if err != nil {
				log.Fatal(err)
			}
		}
	}

	err = CreateHomePage(extensions, renderer)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("build finished")

	if serveLocalhost == true {
		http.Handle("/", http.FileServer(http.Dir("./public")))
		log.Println("Serving on http://localhost:8080")
		log.Fatal(http.ListenAndServe(":8080", nil))
	}
}

// ************************************************************
// global variables go brrrrrrrrrr
// ************************************************************
var contentDir = "./content"
var outputDir = "./public"

// ************************************************************
// types
// ************************************************************
type Content struct {
	Title string
	Body  template.HTML
}

type HomeContent struct {
	Title    string
	Body     template.HTML
	Projects template.HTML
	Contact  template.HTML
}

// ************************************************************
// functions
// ************************************************************

func CreateHomePage(extensions parser.Extensions, renderer *html.Renderer) error {
	homeTmpl, err := template.ParseFiles("templates/home.html")
	if err != nil {
		return err
	}

	files, err := os.ReadDir("./content/home")
	if err != nil {
		return err
	}

	homeContent := HomeContent{Title: "Home"}

	for _, file := range files {
		p := parser.NewWithExtensions(extensions)
		filePath := filepath.Join(contentDir, "home", file.Name())
		fileContent, err := os.ReadFile(filePath)
		if err != nil {
			return err
		}

		astContent := p.Parse(fileContent)
		mdContent := markdown.Render(astContent, renderer)

		switch file.Name() {
		case "index.md":
			homeContent.Body = template.HTML(mdContent)
		case "projects.md":
			homeContent.Projects = template.HTML(mdContent)
		case "contact.md":
			homeContent.Contact = template.HTML(mdContent)
		}
	}

	var buf bytes.Buffer
	err = homeTmpl.Execute(&buf, homeContent)
	if err != nil {
		return err
	}

	err = os.WriteFile("./public/index.html", buf.Bytes(), 0660)
	if err != nil {
		return err
	}

	return nil
}

func ParseBool(str string) (bool, error) {
	switch str {
	case "true", "True", "TRUE", "t", "T", "1":
		return true, nil
	case "false", "False", "FALSE", "f", "F", "0":
		return false, nil
	}
	// defaults to false
	return false, fmt.Errorf("cannot parse '%s' to bool", str)
}

func CreateOutputDirs(targetDir string, outputDirs []string) (count uint32, err error) {
	count = 0
	for _, dir := range outputDirs {
		dir = filepath.Join(targetDir, dir)
		err = os.MkdirAll(dir, 0755)
		if err != nil {
			return count, err
		}

		count = count + 1
	}
	return count, nil
}

// read file
// parse file using the parser you created above
// feed parsed markdown into html renderer and output html
func ProcessFiles(file os.DirEntry, dir string, p *parser.Parser, renderer *html.Renderer, baseTmpl *template.Template) error {
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
	err = baseTmpl.Execute(&buf, page)
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
