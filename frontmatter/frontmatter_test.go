package frontmatter_test

import (
	"testing"
	"testing/fstest"
	"github.com/harrison-blake/lighthouse/frontmatter"
)

var fs = fstest.MapFS{
    "sample-tweet.md": {Data: []byte(`---
DatePublished: 08/21/2025
---
body content goes here`)},

    "sample-bit.md": {Data: []byte(`---
Title: Sample Title
DatePublished: 08/25/2025
---
multiple
lines`)},

"sample-tweet-error-format.md": {Data: []byte(`DatePublished: 08/29/2025
---
body content goes here`)},
}

func TestParseOnTweet(t *testing.T) {
	file, _ := fs.Open("sample-tweet.md")
	frontmatter, body, _ := frontmatter.Parse(file)

	expectedTweet := make(map[string]string)
	expectedTweet["DatePublished"] = "08/21/2025"
	expectedBody := "body content goes here"

	if frontmatter["DatePublished"] != expectedTweet["DatePublished"] {
		t.Errorf("got %v, expected %v", frontmatter["DatePublished"], expectedTweet["DatePublished"])
	}
	
	if body != expectedBody {
		t.Errorf("got %v\n expected %v\n", body, expectedBody)
	}
}

func TestParseOnBit(t *testing.T) {
	file, _ := fs.Open("sample-bit.md")
	frontmatter, body, _ := frontmatter.Parse(file)

	expectedTweet := make(map[string]string)
	expectedTweet["DatePublished"] = "08/25/2025"
	expectedTweet["Title"] = "Sample Title"
	expectedBody := "multiple\nlines"

	if frontmatter["DatePublished"] != expectedTweet["DatePublished"] {
		t.Errorf("got %v, expected %v", frontmatter["DatePublished"], expectedTweet["DatePublished"])
	}

	if frontmatter["Title"] != expectedTweet["Title"] {
		t.Errorf("got %v, expected %v", frontmatter["Title"], expectedTweet["Title"])
	}

	if body != expectedBody {
		t.Errorf("got %v\n expected %v\n", body, expectedBody)
	}
}

func TestFrontmatterFormatError(t *testing.T) {
	file, _ := fs.Open("sample-tweet-error-format.md")
	frontmatter, body, err := frontmatter.Parse(file)

	if frontmatter != nil {
		t.Errorf("got %v, expected nil\n", frontmatter)
	}

	if body != "" {
		t.Errorf("got %v, expected empty string\n", body)
	}
	if err.Error() != "ERROR: invalid frontmatter format" {
		t.Error("wrong error dude LMFAO")
	}
}
