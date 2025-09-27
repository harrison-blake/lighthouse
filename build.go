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
	"strings"
	"sync"
	"time"

	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/html"
	"github.com/gomarkdown/markdown/parser"

	"github.com/harrison-blake/lighthouse/frontmatter"
)

var siteDir = "public"
var partials = "./templates/partials"
var outputDirs = []string{"bits", "about", "stylesheets"}
var stylesheetsSource = "./content/stylesheets"
var styelsheetsDest = "./public/stylesheets"

var mutex = &sync.Mutex{}
var lighthouseFS = os.DirFS(".")

func Build() {
	start := time.Now()

	//********
	// PHASE 1
	//********

	var phase1Jobs []*Job
	createDirsJob := &Job{
		Name: "Create output directories",
		F: func() error {
			return CreateOutputDirectories(siteDir, outputDirs)
		},
	}
	phase1Jobs = append(phase1Jobs, createDirsJob)

	// AGGREGATE TWEET CONTENT
	var tweets []TweetPageData
	files, err := fs.ReadDir(lighthouseFS, "content/tweets")
	if err != nil {
		log.Fatalf("FAILED TO READ DIRECTORY: %v", err)
	}

	for _, file := range files {
		fileName := file.Name()
		if fileName == ".DS_Store" {
			continue
		}

		j := &Job{
			Name: fmt.Sprintf("parse tweet: %v", fileName),
			F: func() error {
				fsFile, err := lighthouseFS.Open(fmt.Sprintf("content/tweets/%v", fileName))
				if err != nil {
					return err
				}
				fm, content, err := frontmatter.Parse(fsFile)

				defer fsFile.Close()
				if err != nil {
					return err
				}

				tweet := TweetPageData{Content: content, Frontmatter: fm}

				mutex.Lock()
				tweets = append(tweets, tweet)
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
				fsFile, err := os.Open("content/bits/" + bitsFileName)
				if err != nil {
					return err
				}
				defer fsFile.Close()

				fm, content, err := frontmatter.Parse(fsFile)        

				p := DefaultParser()
				doc := p.Parse([]byte(content))
				r := DefaultRenderer()
				html := markdown.Render(doc, r)

				mutex.Lock()
				bits = append(bits, Bit{
					Content: template.HTML(html),
					FM:      fm,
					Slug:    strings.ReplaceAll(strings.ToLower(fm["Title"]), " ", "-"),
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

			// Get top 3 tweets (NEEDS A REFACTOR BADLY BUT IM FEELING LAZY)
			var topTweets []TweetPageData
			mutex.Lock()
			topTweets = tweets[:1]
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
				Tweets: topTweets,
				Bits:   topBits,
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

			var buf bytes.Buffer
			err = aboutTempl.Execute(&buf, nil)
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

	// RENDER BITS SHOW PAGES	
	for _, bit := range bits {
		j := &Job{
			Name: fmt.Sprintf("Render Bit: %s", bit.FM["Title"]),
			F: func() error {
				bitShowTempl, err := template.ParseFiles("./templates/bits/show.html", "./templates/partials/nav.html", "./templates/partials/footer.html")
				if err != nil {
					return err
				}

				slug := strings.ReplaceAll(strings.ToLower(bit.FM["Title"]), " ", "-")
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

func DefaultRenderer() *html.Renderer {
	htmlFlags := html.CommonFlags | html.HrefTargetBlank
	opts := html.RendererOptions{Flags: htmlFlags}

	return html.NewRenderer(opts)
}

func DefaultParser() *parser.Parser {
	defaultExtensions := parser.CommonExtensions | parser.AutoHeadingIDs

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
	Tweets []TweetPageData
	Bits   []Bit
}

type Bit struct {
	Content template.HTML
	FM      map[string]string
	Slug    string
}

type TweetPageData struct {
	Content     string
	Frontmatter map[string]string
}

type BitsPageData struct {
	Bits []Bit
}
