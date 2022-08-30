package main

import (
	"bytes"
	"github.com/PuerkitoBio/goquery"
	"github.com/avelino/awesome-go/pkg/markdown"
	"html/template"
	"io"
	"os"
)

const (
	indexTemplateFile = "tmpl/tmpl.html"
	readmeFile        = "./README.md"
)

type content struct {
	Body template.HTML
}

// GenerateHTML generate site html from markdown file
func GenerateHTML() ([]byte, error) {
	readmeContent := readmeHTML()

	t := template.Must(template.ParseFiles(indexTemplateFile))
	buf := new(bytes.Buffer)
	err := t.Execute(buf, &content{Body: template.HTML(readmeContent)})
	if err != nil {
		return nil, err
	}
	return io.ReadAll(buf)
}

func readmeHTML() []byte {
	readmeContent, err := os.ReadFile(readmeFile)
	if err != nil {
		panic(err)
	}

	html, _ := markdown.ConvertMarkdownFileToHTML(readmeContent)
	return html
}

func startQuery() *goquery.Document {
	buf := bytes.NewReader(readmeHTML())
	query, err := goquery.NewDocumentFromReader(buf)
	if err != nil {
		panic(err)
	}
	return query
}
