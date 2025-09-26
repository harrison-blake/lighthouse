package main

import (
	"bufio"
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
)

var siteDir = "public"
var partials = "./templates/partials"
var outputDirs = []string{"bits", "about", "stylesheets"}
var stylesheetsSource = "./content/stylesheets"
var styelsheetsDest = "./public/stylesheets"

var mutex = &sync.Mutex{}


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
	var tweetRes []string
	files, err := fs.ReadDir(os.DirFS("."), "content/tweets")
	if err != nil {
		log.Fatalf("FAILED TO READ DIRECTORY: %v", err)
	}

	for _, file := range files {
		if file.Name() == ".DS_Store" {
			continue
		}

		j := &Job{
			Name: file.Name(),
			F: func() error {
				content, err := os.ReadFile("content/tweets/" + file.Name())
				if err != nil {
					return err
				}

				mutex.Lock()
				tweetRes = append(tweetRes, strings.TrimSpace(string(content)))
				mutex.Unlock()
				return nil
			},
		}
		phase1Jobs = append(phase1Jobs, j)
	}

	// AGGREGATE BITS
	var bitsRes []Bit
	bitFiles, err := fs.ReadDir(os.DirFS("."), "content/bits")
	if err != nil {
		log.Fatalf("FAILED TO READ DIRECTORY: %v", err)
	}

	for _, file := range bitFiles {
		if file.Name() == ".DS_Store" {
			continue
		}
		
		j := &Job{
			Name: fmt.Sprintf("get %v", file.Name()),
			F: func() error {
				file, err := os.Open("content/bits/" + file.Name())
				if err != nil {
					return err
				}
				defer file.Close()

				scanner := bufio.NewScanner(file)
				scanner.Scan()
				title := strings.Trim(scanner.Text(), "# ")

				var content strings.Builder
                for scanner.Scan() {
                    content.WriteString(scanner.Text() + "\n")
                }                

				p := DefaultParser()
				doc := p.Parse([]byte(content.String()))
				r := DefaultRenderer()
				html := markdown.Render(doc, r)

				mutex.Lock()
				bitsRes = append(bitsRes, Bit{
					Title:   title,
					Content: template.HTML(html),
					Slug:    strings.ReplaceAll(strings.ToLower(title), " ", "-"),
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

			// Get top 3 tweets
			var topTweets []string
			mutex.Lock()
			if len(tweetRes) > 3 {
				topTweets = tweetRes[:3]
			} else {
				topTweets = tweetRes
			}
			mutex.Unlock()

			// Get top 3 bits
			var topBits []Bit
			mutex.Lock()
			if len(bitsRes) > 3 {
				for _, bit := range bitsRes[:3] {
					topBits = append(topBits, bit)
				}
			} else {
				for _, bit := range bitsRes {
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
				Bits: bitsRes,
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
	for _, bit := range bitsRes {
		j := &Job{
			Name: fmt.Sprintf("Render Bit: %s", bit.Title),
			F: func() error {
				bitShowTempl, err := template.ParseFiles("./templates/bits/show.html", "./templates/partials/nav.html", "./templates/partials/footer.html")
				if err != nil {
					return err
				}

				slug := strings.ReplaceAll(strings.ToLower(bit.Title), " ", "-")
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
	Tweets []string
	Bits   []Bit
}

type Bit struct {
	Title   string
	Content template.HTML
	Slug    string
}

type TweetsPageData struct {
	Tweets []string
}

type BitsPageData struct {
	Bits []Bit
}
