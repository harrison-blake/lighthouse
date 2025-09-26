package frontmatter

import (
	"bufio"
	"fmt"
	"io/fs"
	"strings"
)

func Parse(file fs.File) (map[string]string, string, error) {
	scanner := bufio.NewScanner(file)
	frontmatter := make(map[string]string)
	var bodyLines []string

	isFrontmatter := true

	// crude check of frontmatter formatting
	scanner.Scan()
	if scanner.Text() != "---" {
		return nil, "", fmt.Errorf("ERROR: invalid frontmatter format")
	}

	for scanner.Scan() {
		line := scanner.Text()
		if line == "---" {
			isFrontmatter = false
			continue
		}

		if isFrontmatter {
			res := strings.Split(line, ": ")
			frontmatter[res[0]] = res[1]
		} else {
			bodyLines = append(bodyLines, line)
		}
	}

	body := strings.Join(bodyLines, "\n")

	return frontmatter, body, nil
}
