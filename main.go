package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/spf13/cobra"
)

var serve bool

var rootCmd = &cobra.Command{
	Use:   "lighthouse",
	Short: "Lighthouse is a static site generator",
	Long:  `A Fast and Flexible Static Site Generator in Go.`,
	Run: func(cmd *cobra.Command, args []string) {
	},
}

var buildCmd = &cobra.Command{
	Use:   "build",
	Short: "Builds the site",
	Long:  `Builds the site and outputs it to the public directory.`,
	Run: func(cmd *cobra.Command, args []string) {
		Build()
		if serve {
			fs := http.FileServer(http.Dir("public"))
			http.Handle("/", fs)
			port := ":8080"
			log.Printf("Serving public directory on http://localhost%s\n", port)
			log.Fatal(http.ListenAndServe(port, nil))
		}
	},
}

func init() {
	rootCmd.AddCommand(buildCmd)
	buildCmd.Flags().BoolVar(&serve, "serve", false, "Serve the public directory on localhost after building")
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func main() {
	Execute()
}
