package cmd

import (
	"errors"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	//"gopkg.in/russross/blackfriday.v2"
)

func NewPost(filename string) Post {
	post, err := ExtractMetaFromFilename(filename)
	if err != nil {
		log.Fatalln(err)
	}

	mdfile, err := os.Open("./src/markdown/" + filename)
	if err != nil {
		log.Fatal(err)
	}
	defer mdfile.Close()

	rawBytes, err := ioutil.ReadAll(mdfile)

	// Get title from first line of file
	lines := strings.Split(string(rawBytes), "\n")
	post.Title = strings.Replace(lines[0], "# ", "", -1)

	// Convert Markdown to HTML
	// body := blackfriday.Run(rawBytes)
	// post.Body = template.HTML(body)
	post.Body = strings.Replace(string(rawBytes), "# " + post.Title, "", -1)

	//fmt.Println(post.Body)

	// Save file
	post.createFile()

	return post
}

type Post struct {
	Title string
	// Body  template.HTML
	Body  string
	Date  time.Time
	Slug  string
}

func (post *Post) getURL() string {
	return fmt.Sprintf("/posts/%s", post.Slug)
}

func (post *Post) createFile() {
	// Create folder for HTML
	newPath := filepath.Join("dist/posts", post.Slug)
	_ = os.MkdirAll(newPath, os.ModePerm)

	// Create HTML file
	f, err := os.Create("dist/posts/" + post.Slug + "/" + "index.html")
	if err != nil {
		log.Fatal(err)
	}

	// Generate final HTML file from template
	t, _ := template.ParseFiles("./src/post.html")
	err = t.Execute(f, post)
	if err != nil {
		log.Fatalf("can't execute template: %v", err)
	}
	_ = f.Close()
}

func fNameWithoutExtension(fn string) string {
	return strings.TrimSuffix(fn, path.Ext(fn))
}

func ExtractMetaFromFilename(filename string) (Post, error) {
	errorMessage := fmt.Sprintf("can't parse filename '%s', it should be in format 'YYYY-MM-DD-slug.md'", filename)
	dateFormat := "2006-01-02"
	slug := fNameWithoutExtension(filename)[len(dateFormat)+1:]
	if len(slug) == 0 {
		return Post{}, errors.New(errorMessage)
	}
	date, err := time.Parse(dateFormat, filename[:len(dateFormat)])
	if err != nil {
		return Post{}, errors.New(errorMessage)
	}

	return Post{Slug: slug, Date: date}, nil
}
