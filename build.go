package main

import (
	"bytes"
	"fmt"
	"html/template"
	"io/fs"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/html"
	"github.com/gomarkdown/markdown/parser"

	"github.com/harrison-blake/lighthouse/frontmatter"
)

var siteDir = "public"
var outputDirs = []string{"bits", "about", "stylesheets", "forecasts"}
var stylesheetsSource = "./content/stylesheets"
var styelsheetsDest = "./public/stylesheets"

var mutex = &sync.Mutex{}
var lighthouseFS = os.DirFS(".")

func sanitizeSlug(slug string) (string, error) {
	slug = strings.ToLower(slug)
	slug = strings.ReplaceAll(slug, " ", "-")
	reg, err := regexp.Compile("[^a-zA-Z0-9-]+")
	if err != nil {
		return "", err
	}
	slug = reg.ReplaceAllString(slug, "")
	return slug, nil
}

func Build() {
	start := time.Now()

	//********
	// PHASE 1
	//********
	r := DefaultRenderer()

	var phase1Jobs []*Job
	createDirsJob := &Job{
		Name: "Create output directories",
		F: func() error {
			return CreateOutputDirectories(siteDir, outputDirs)
		},
	}
	phase1Jobs = append(phase1Jobs, createDirsJob)

	// AGGREGATE FORECAST CONTENT
	var forecasts []Forecast
	files, err := fs.ReadDir(lighthouseFS, "content/forecasts")
	if err != nil {
		log.Fatalf("FAILED TO READ DIRECTORY: %v", err)
	}

	for _, file := range files {
		fileName := file.Name()
		if fileName == ".DS_Store" {
			continue
		}

		j := &Job{
			Name: fmt.Sprintf("parse forecast: %v", fileName),
			F: func() error {
				fsFile, err := lighthouseFS.Open(fmt.Sprintf("content/forecasts/%v", fileName))
				if err != nil {
					return err
				}
				defer fsFile.Close()

				fm, content, err := frontmatter.Parse(fsFile)
				if err != nil {
					return err
				}

				p := DefaultParser()
				doc := p.Parse([]byte(content))
				rendered := markdown.Render(doc, r)

				slug, err := sanitizeSlug(fm["Title"])
				if err != nil {
					return err
				}

				mutex.Lock()
				forecasts = append(forecasts, Forecast{
					Content: template.HTML(rendered),
					FM:      fm,
					Slug:    slug,
				})
				mutex.Unlock()
				return nil
			},
		}
		phase1Jobs = append(phase1Jobs, j)
	}

	// AGGREGATE BITS
	var bits []Bit
	bitFiles, err := fs.ReadDir(lighthouseFS, "content/bits")
	if err != nil {
		log.Fatalf("FAILED TO READ DIRECTORY: %v", err)
	}

	for _, file := range bitFiles {
		bitsFileName := file.Name()
		if bitsFileName == ".DS_Store" {
			continue
		}

		j := &Job{
			Name: fmt.Sprintf("get %v", bitsFileName),
			F: func() error {
				fsFile, err := lighthouseFS.Open("content/bits/" + bitsFileName)
				if err != nil {
					return err
				}
				defer fsFile.Close()

				fm, content, err := frontmatter.Parse(fsFile)
				if err != nil {
					return err
				}

				p := DefaultParser()
				doc := p.Parse([]byte(content))
				rendered := markdown.Render(doc, r)

				slug, err := sanitizeSlug(fm["Title"])
				if err != nil {
					return err
				}

				mutex.Lock()
				bits = append(bits, Bit{
					Content: template.HTML(rendered),
					FM:      fm,
					Slug:    slug,
				})

				mutex.Unlock()
				return nil
			},
		}
		phase1Jobs = append(phase1Jobs, j)
	}

	pool := NewWorkerPool(5)
	pool.Run(phase1Jobs)

	//********
	// PHASE 2
	//********

	var phase2Jobs []*Job
	// RENDER HOME PAGE
	homepageJob := &Job{
		Name: "Render Homepage",
		F: func() error {
			homeTempl, err := template.ParseFiles("./templates/home/index.html", "./templates/partials/nav.html", "./templates/partials/footer.html")
			if err != nil {
				return err
			}

			// Get top 3 forecasts
			var topForecasts []Forecast
			mutex.Lock()
			topForecasts = forecasts[:1]
			mutex.Unlock()

			// Get top 3 bits
			var topBits []Bit
			mutex.Lock()
			if len(bits) > 3 {
				for _, bit := range bits[:3] {
					topBits = append(topBits, bit)
				}
			} else {
				for _, bit := range bits {
					topBits = append(topBits, bit)
				}
			}
			mutex.Unlock()

			data := HomepageData{
				Forecasts: topForecasts,
				Bits:      topBits,
			}

			var buf bytes.Buffer
			err = homeTempl.Execute(&buf, data)
			if err != nil {
				return err
			}

			err = os.WriteFile("./public/index.html", buf.Bytes(), 0644)
			if err != nil {
				return err
			}

			return nil
		},
	}
	phase2Jobs = append(phase2Jobs, homepageJob)

	// RENDER ABOUT PAGE
	aboutJob := &Job{
		Name: "Render About Page",
		F: func() error {
			aboutTempl, err := template.ParseFiles("./templates/about/index.html", "./templates/partials/nav.html", "./templates/partials/footer.html")
			if err != nil {
				return err
			}

			fsFile, err := lighthouseFS.Open("content/about/about.md")
				if err != nil {
					return err
				}

			_, content, err := frontmatter.Parse(fsFile)
			if err != nil {
				return err
			}

			p := DefaultParser()
			doc := p.Parse([]byte(content))
			rendered := markdown.Render(doc, r)

			data := map[string]template.HTML{"Content": template.HTML(rendered)}

			var buf bytes.Buffer
			err = aboutTempl.Execute(&buf, data)
			if err != nil {
				return err
			}

			err = os.WriteFile("./public/about/index.html", buf.Bytes(), 0644)
			if err != nil {
				return err
			}

			return nil
		},
	}
	phase2Jobs = append(phase2Jobs, aboutJob)

	// RENDER BITS INDEX PAGE
	bitsPageJob := &Job{
		Name: "Render Bits Page",
		F: func() error {
			bitsTempl, err := template.ParseFiles("./templates/bits/index.html", "./templates/partials/nav.html", "./templates/partials/footer.html")
			if err != nil {
				return err
			}

			data := BitsPageData{
				Bits: bits,
			}

			var buf bytes.Buffer
			err = bitsTempl.Execute(&buf, data)
			if err != nil {
				return err
			}

			err = os.WriteFile("./public/bits/index.html", buf.Bytes(), 0644)
			if err != nil {
				return err
			}

			return nil
		},
	}
	phase2Jobs = append(phase2Jobs, bitsPageJob)

	// RENDER FORECASTS INDEX PAGE
	forecastsPageJob := &Job{
		Name: "Render Forecasts Page",
		F: func() error {
			forecastsTempl, err := template.ParseFiles("./templates/forecasts/index.html", "./templates/partials/nav.html", "./templates/partials/footer.html")
			if err != nil {
				return err
			}

			data := ForecastsPageData{
				Forecasts: forecasts,
			}

			var buf bytes.Buffer
			err = forecastsTempl.Execute(&buf, data)
			if err != nil {
				return err
			}

			err = os.WriteFile("./public/forecasts/index.html", buf.Bytes(), 0644)
			if err != nil {
				return err
			}

			return nil
		},
	}
	phase2Jobs = append(phase2Jobs, forecastsPageJob)

	// RENDER BITS SHOW PAGES
	for _, bit := range bits {
		j := &Job{
			Name: fmt.Sprintf("Render Bit: %s", bit.FM["Title"]),
			F: func() error {
				bitShowTempl, err := template.ParseFiles("./templates/bits/show.html", "./templates/partials/nav.html", "./templates/partials/footer.html")
				if err != nil {
					return err
				}

				slug, err := sanitizeSlug(bit.FM["Title"])
				if err != nil {
					return err
				}
				outputPath := fmt.Sprintf("./public/bits/%s.html", slug)

				var buf bytes.Buffer
				err = bitShowTempl.Execute(&buf, bit)
				if err != nil {
					return err
				}

				err = os.WriteFile(outputPath, buf.Bytes(), 0644)
				if err != nil {
					return err
				}

				return nil
			},
		}
		phase2Jobs = append(phase2Jobs, j)
	}

	// RENDER FORECASTS SHOW PAGES
	for _, forecast := range forecasts {
		j := &Job{
			Name: fmt.Sprintf("Render Forecast: %s", forecast.FM["Title"]),
			F: func() error {
				forecastShowTempl, err := template.ParseFiles("./templates/forecasts/show.html", "./templates/partials/nav.html", "./templates/partials/footer.html")
				if err != nil {
					return err
				}

				slug, err := sanitizeSlug(forecast.FM["Title"])
				if err != nil {
					return err
				}
				outputPath := fmt.Sprintf("./public/forecasts/%s.html", slug)

				var buf bytes.Buffer
				err = forecastShowTempl.Execute(&buf, forecast)
				if err != nil {
					return err
				}

				err = os.WriteFile(outputPath, buf.Bytes(), 0644)
				if err != nil {
					return err
				}

				return nil
			},
		}
		phase2Jobs = append(phase2Jobs, j)
	}

	pool.Run(phase2Jobs)

	var phase3Jobs []*Job
	compileTailwindJob := &Job{
		Name: "Compile Tailwind CSS",
		F: func() error {
			cmd := exec.Command("./tailwindcss", "-i", "./content/stylesheets/input.css", "-o", "./public/stylesheets/output.css", "--config", "tailwind.config.js")
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			return cmd.Run()
		},
	}
	phase3Jobs = append(phase3Jobs, compileTailwindJob)

	pool.Run(phase3Jobs)

	p1End := time.Now()
	fmt.Printf("Build took %v\n", p1End.Sub(start))
}

func CreateOutputDirectories(siteDir string, paths []string) error {
	err := os.MkdirAll(siteDir, 0750)
	if err != nil {
		return fmt.Errorf("failed to create %s directory\n %w", siteDir, err)
	}

	for _, path := range paths {
		fullPath := filepath.Join(siteDir, path)
		if err := os.MkdirAll(fullPath, 0755); err != nil {
			return fmt.Errorf("failed to create %s directory\n %w", fullPath, err)
		}
	}

	return nil
}

// CAN reuse rendered
func DefaultRenderer() *html.Renderer {
	htmlFlags := html.CommonFlags | html.HrefTargetBlank
	opts := html.RendererOptions{Flags: htmlFlags}

	return html.NewRenderer(opts)
}

// CANNOT reuse parser
func DefaultParser() *parser.Parser {
	defaultExtensions := parser.CommonExtensions | parser.AutoHeadingIDs | parser.NoEmptyLineBeforeBlock

	return parser.NewWithExtensions(defaultExtensions)
}

type Job struct {
	Name string
	F    func() error
}

type WorkerPool struct {
	Jobs        chan *Job
	Concurrency int
	wg          sync.WaitGroup
}

func NewWorkerPool(c int) *WorkerPool {
	workerPool := &WorkerPool{
		Concurrency: c,
	}

	return workerPool
}

func (wp *WorkerPool) Worker() {
	for job := range wp.Jobs {
		err := job.F()
		if err != nil {
			log.Printf("Job %s failed: %v\n", job.Name, err)
		}
		wp.wg.Done()
	}
}

func (wp *WorkerPool) Run(jobBatch []*Job) {
	wp.Jobs = make(chan *Job, len(jobBatch))
	wp.wg.Add(len(jobBatch))
	for i := 0; i < wp.Concurrency; i++ {
		go wp.Worker()
	}

	for _, job := range jobBatch {

		wp.Jobs <- job
	}

	close(wp.Jobs)
	wp.wg.Wait()
}

type HomepageData struct {
	Forecasts []Forecast
	Bits      []Bit
}

type Bit struct {
	Content template.HTML
	FM      map[string]string
	Slug    string
}

type Forecast struct {
	Content template.HTML
	FM      map[string]string
	Slug    string
}

type BitsPageData struct {
	Bits []Bit
}

type ForecastsPageData struct {
	Forecasts []Forecast
}