package main

import (
	"bytes"
	"fmt"
	"github.com/avelino/awesome-go/pkg/slug"
	"log"
	"os"
	"strings"
	"text/template"

	"github.com/PuerkitoBio/goquery"
	cp "github.com/otiai10/copy"
)

const (
	outputDirectory      = "public"
	indexOutputFile      = outputDirectory + "/index.html"
	sitemapOutputFile    = outputDirectory + "/sitemap.xml"
	categoryPageTemplate = "tmpl/cat-tmpl.html"
	sitemapTemplate      = "tmpl/sitemap-tmpl.xml"
)

var (
	publicFiles = map[string]string{
		"tmpl/assets":     "public/assets",
		"tmpl/robots.txt": "public/robots.txt",
		"tmpl/_redirects": "public/_redirects",
	}
)

type Link struct {
	Title       string
	Url         string
	Description string
}

type Category struct {
	Title       string
	Slug        string
	Description string
	Items       []Link
}

func main() {
	err := prepareOutput()
	if err != nil {
		panic(err)
	}

	html, err := GenerateHTML()
	if err != nil {
		panic(err)
	}

	err = os.WriteFile(indexOutputFile, html, 0644)
	if err != nil {
		panic(html)
	}

	query, err := goquery.NewDocumentFromReader(bytes.NewReader(html))
	if err != nil {
		panic(err)
	}

	categories, err := renderCategoryPages(query)
	if err != nil {
		panic(err)
	}

	err = changeLinksInIndex(html, query, categories)
	if err != nil {
		panic(err)
	}

	err = makeSitemap(categories)
	if err != nil {
		panic(err)
	}

	err = copyPublicFiles()
	if err != nil {
		panic(err)
	}

	log.Println("Successfully generated HTML output 🦄")

	os.Exit(0)
}

func prepareOutput() error {
	err := os.RemoveAll(outputDirectory)
	if err != nil {
		return err
	}

	return os.MkdirAll(outputDirectory, 0755)
}

func copyPublicFiles() (err error) {
	for from, to := range publicFiles {
		err = cp.Copy(from, to)
		if err != nil {
			return err
		}
	}
	return
}

func renderCategoryPages(query *goquery.Document) (map[string]*Category, error) {
	categories := findCategories(query)
	for _, obj := range categories {
		err := renderCategoryPage(obj)
		if err != nil {
			return nil, err
		}
	}

	return categories, nil
}

func renderCategoryPage(obj *Category) error {
	folder := fmt.Sprintf("%s/%s", outputDirectory, obj.Slug)
	err := os.MkdirAll(folder, 0755)
	if err != nil {
		return err
	}

	t := template.Must(template.ParseFiles(categoryPageTemplate))
	f, err := os.Create(fmt.Sprintf("%s/index.html", folder))
	if err != nil {
		return err
	}
	return t.Execute(f, obj)
}

func findCategories(query *goquery.Document) map[string]*Category {
	categories := make(map[string]*Category)
	query.Find("body #contents").NextFiltered("ul").Find("ul").Each(func(_ int, s *goquery.Selection) {
		s.Find("li a").Each(func(_ int, s *goquery.Selection) {
			selector, exists := s.Attr("href")
			if !exists {
				return
			}
			category := makeCategoryByID(selector, query.Find("body"))
			if category == nil {
				return
			}
			categories[selector] = category
		})
	})

	return categories
}

func makeSitemap(objs map[string]*Category) error {
	t := template.Must(template.ParseFiles(sitemapTemplate))
	f, err := os.Create(sitemapOutputFile)
	if err != nil {
		return err
	}
	return t.Execute(f, objs)
}

func makeCategoryByID(selector string, s *goquery.Selection) (obj *Category) {
	s.Find(selector).Each(func(_ int, s *goquery.Selection) {
		desc := s.NextFiltered("p")
		ul := s.NextFilteredUntil("ul", "h2")

		links := []Link{}
		ul.Find("li").Each(func(_ int, s *goquery.Selection) {
			url, _ := s.Find("a").Attr("href")
			link := Link{
				Title:       s.Find("a").Text(),
				Description: s.Text(),
				Url:         url,
			}
			links = append(links, link)
		})
		if len(links) == 0 {
			return
		}
		obj = &Category{
			Slug:        slug.Generate(s.Text()),
			Title:       s.Text(),
			Description: desc.Text(),
			Items:       links,
		}
	})
	return
}

func changeLinksInIndex(html []byte, query *goquery.Document, objs map[string]*Category) error {
	query.Find("body #content ul li ul li a").Each(func(_ int, s *goquery.Selection) {
		href, hrefExists := s.Attr("href")
		if !hrefExists {
			return
		}

		// do not replace links if no page has been created for it
		_, objExists := objs[href]
		if !objExists {
			return
		}

		uri := strings.SplitAfter(href, "#")
		if len(uri) >= 2 && uri[1] != "contents" {
			oldLink := fmt.Sprintf(`href="%s"`, href)
			newLink := fmt.Sprintf(`href="%s"`, uri[1])
			html = bytes.ReplaceAll(html, []byte(oldLink), []byte(newLink))
		}
	})

	return os.WriteFile(indexOutputFile, html, 0644)
}
