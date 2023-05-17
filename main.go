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
	"time"
)

var (
	root, _         = getDir()
	mePath          = root + "/about/me.go"
	experiencePath  = root + "/about/experience.go"
	contactPath     = root + "/about/contact.go"
	templatePath    = root + "/template/template.html"
	pageMePath      = root + "/page/index.html"
	pageExpPath     = root + "/page/experience.html"
	pageContactPath = root + "/page/contact.html"

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
				"ListExperience",
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
				"ListContactOptions",
			},
			format: `<a href="contact.html" title="See contact options">$1</a>`,
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
				"ListContactOptions",
			},
			format: `<span class="function">$1</span>`,
		},
		{
			symbols: []string{
				"^//.*",
				`^/\*.*\*/`,
				"[^https:]//.*",
			},
			format: `<span class="comment">$1</span>`,
		},
		{
			symbols: []string{
				"rdominguez@tecnologer.net",
			},
			format: `<a href="mailto:${1}">${1}</a>`,
		},
		{
			symbols: []string{
				`"(http(s)?:\/\/.+)"`,
			},
			format: `<a href="${2}">${1}</a>`,
		},
	}
)

func main() {
	version := time.Now().Format("2006.0102")

	fileMe, err := os.ReadFile(mePath)
	if err != nil {
		log.Fatal(err)
	}

	fileExp, err := os.ReadFile(experiencePath)
	if err != nil {
		log.Fatal(err)
	}

	fileContact, err := os.ReadFile(contactPath)
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
		"{{version}}": []byte(version),
	}

	writeFile(pageMePath, insertText(rules, template))

	rules = map[string][]byte{
		"{{code}}":    formatText(fileExp),
		"{{file}}":    []byte("experience"),
		"{{go_back}}": []byte(`<a href="/" title="return to home"><</a>`),
		"{{version}}": []byte(version),
	}

	writeFile(pageExpPath, insertText(rules, template))

	rules = map[string][]byte{
		"{{code}}":    formatText(fileContact),
		"{{file}}":    []byte("contact"),
		"{{go_back}}": []byte(`<a href="/" title="return to home"><</a>`),
		"{{version}}": []byte(version),
	}

	writeFile(pageContactPath, insertText(rules, template))

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
