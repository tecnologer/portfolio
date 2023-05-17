package main

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"os"
	"path"
	"regexp"
	"runtime"
)

var (
	root, _        = getDir()
	mePath         = root + "/about/me.go"
	experiencePath = root + "/about/experience.go"
	templatePath   = root + "/template/template.html"
	pageMePath     = root + "/page/index.html"
	pageExpPath    = root + "/page/experience.html"

	format = []struct {
		symbols []string
		format  string
	}{
		{
			symbols: []string{
				`"([^"]+)"`,
				"`([^`]+)`",
			},
			format: `<span class="string">$1</span>`,
		},
		{
			symbols: []string{
				`[^"']string`,
			},
			format: `<span class="primitivetype">${1}</span>`,
		},
		{
			symbols: []string{
				" nil",
				"package ",
				"type ",
				" struct ",
				"func ",
				"return ",
				"\\[",
				"\\]",
			},
			format: `<span class="keyword">$1</span>`,
		},
		{
			symbols: []string{
				"ListExperience()",
			},
			format: `<a href="experience.html" title="Expand experience list">$1</a>`,
		},
		{
			symbols: []string{
				"\\(",
				"\\)",
			},
			format: `<span class="parenthesis">$1</span>`,
		},
		{
			symbols: []string{
				"\\{",
				"\\}",
			},
			format: `<span class="curlybrace">$1</span>`,
		},
		{
			symbols: []string{
				"NewMe",
				"ListExperience",
			},
			format: `<span class="function">$1</span>`,
		},
		{
			symbols: []string{
				"//.*",
				`/\*.*\*/`,
			},
			format: `<span class="comment">$1</span>`,
		},
	}
)

func main() {
	fileMe, err := os.ReadFile(mePath)
	if err != nil {
		log.Fatal(err)
	}

	fileExp, err := os.ReadFile(experiencePath)
	if err != nil {
		log.Fatal(err)
	}

	template, err := os.ReadFile(templatePath)
	if err != nil {
		log.Fatal(err)
	}

	rules := map[string][]byte{
		"{{code}}":    formatText(fileMe),
		"{{file}}":    []byte("me"),
		"{{go_back}}": make([]byte, 0),
	}

	writeFile(pageMePath, insertText(rules, template))

	rules = map[string][]byte{
		"{{code}}":    formatText(fileExp),
		"{{file}}":    []byte("experience"),
		"{{go_back}}": []byte(`<a href="index.html" title="return to home"><</a>`),
	}

	writeFile(pageExpPath, insertText(rules, template))

}

func formatText(fileContent []byte) []byte {
	var reg *regexp.Regexp
	var err error

	for _, f := range format {
		for _, word := range f.symbols {
			reg, err = regexp.Compile(fmt.Sprintf("(?mi)(%s)", word))
			if err != nil {
				log.Println(err)
				continue
			}
			fileContent = reg.ReplaceAll(fileContent, []byte(f.format))
			// fileContent = bytes.ReplaceAll(fileContent, []byte(word), []byte(fmt.Sprintf(f.format, word)))
		}
	}

	return fileContent
}

func writeFile(path string, content []byte) error {
	return os.WriteFile(path, content, 0777)
}

func insertText(rules map[string][]byte, origin []byte) []byte {
	for key, value := range rules {
		origin = bytes.ReplaceAll(origin, []byte(key), value)
	}

	return origin
}

func getDir() (string, error) {
	_, filename, _, ok := runtime.Caller(1)
	if !ok {
		return "", errors.New("get dir for connection file")
	}

	return path.Dir(filename), nil
}
