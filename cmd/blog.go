package cmd

import (
	"context"
	"errors"
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path"
	"path/filepath"
	"sort"
	"sync"

	"underblog/internal"
)

const DefaultMarkdownPath = "./src/markdown/"

// Create and initialize Blog
func NewBlog(opts internal.Opts) *Blog {
	b := new(Blog)

	b.opts = opts

	b.mux = &sync.Mutex{}
	b.wg = &sync.WaitGroup{}

	b.files = make(chan os.FileInfo)

	return b
}

// Blog options and blog creating methods
type Blog struct {
	opts internal.Opts

	meta      BlogMeta
	files     chan os.FileInfo
	Posts     []Post
	indexPage io.Writer

	mux *sync.Mutex
	wg  *sync.WaitGroup
}

// Render md-files->HTML, generate root index.html
func (b *Blog) Render() error {
	if err := b.verifyMarkdownPresent(); err != nil {
		log.Fatal(errors.New(fmt.Sprintf("Markdown directory is not found: %v", err)))
	}

	b.indexPage = b.getIndexPage(b.opts.Path)
	b.createPosts()
	err := b.renderMd()
	b.copyCssToPublicDir()

	return err
}

func (b *Blog) addPost(post Post) {
	b.mux.Lock()
	b.Posts = append(b.Posts, post)
	b.mux.Unlock()
}

func (b *Blog) getIndexPage(currentPath string) io.Writer {
	rootPath := "."

	if currentPath != "" {
		rootPath = currentPath
	}
	p := filepath.Join(rootPath, "dist")
	err := os.MkdirAll(p, os.ModePerm)
	if err != nil {
		log.Fatal(errors.New(fmt.Sprintf("Can't create public dir: %v", err)))
	}

	f, err := os.Create("dist/index.html")

	if err != nil {
		log.Fatal(errors.New(fmt.Sprintf("Can't create dist/index.html: %v", err)))
	}

	return f
}

func (b *Blog) startWorker(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case file, ok := <-b.files:
			if !ok || !isFileValid(file) {
				// todo: catch it?
				b.wg.Done()
				return
			}
			b.addPost(NewPost(file.Name()))
			b.wg.Done()
		}
	}

}

func (b *Blog) getMdFiles() []os.FileInfo {
	files, err := ioutil.ReadDir(DefaultMarkdownPath)
	if err != nil {
		fmt.Println("Can't get directory of markdown files")
		log.Fatal(err)
	}
	return files
}

func (b *Blog) createPosts() {
	ctx := context.Background()

	filesChan := make(chan os.FileInfo)
	files := b.getMdFiles()

	wLimit := internal.GetWorkersLimit(len(files))
	b.wg.Add(len(files))

	for i := 0; i < wLimit; i++ {
		go b.startWorker(ctx)
	}

	for _, file := range files {
		b.files <- file
	}

	close(filesChan)
}

func (b *Blog) copyCssToPublicDir() {
	copyDir("./src/static", "./dist/static")
}

func IsExist(path string) (bool) {
	if _, err := os.Stat(path); err == nil { // 检查路径是否存在
		return true
	}
	return false
}

func copyDir(path, newpath string)(bool){
	path, _ = filepath.Abs(path)
	newpath, _ = filepath.Abs(newpath)

	if file_names, err := ioutil.ReadDir(path); err == nil {
		if IsExist(newpath) == false {
			os.MkdirAll(newpath, os.ModePerm)
		}
		for _, file := range file_names {
				//fmt.Println(path+"/"+file.Name(), file.IsDir())
				file_path, _ := filepath.Abs(path+"/"+file.Name())
				new_file_path, _ := filepath.Abs(newpath+"/"+file.Name())
				if file.IsDir() == true {
					if IsExist(newpath) == false {
						os.MkdirAll(newpath, os.ModePerm)
					}
					//fmt.Println("file_path:", file_path, "new_dir_path:", new_file_path)
					copyDir(file_path, new_file_path)
				}else{
					copyFile(file_path, new_file_path)
				}
		}
		return true
	}

	return false
}

func copyFile(path, topath string)(bool){
	if data, err := ioutil.ReadFile(path); err == nil {
		if ioutil.WriteFile(topath, data, os.ModePerm) == nil {
			return true
		}
	}
	return false
}

func (b *Blog) renderMd() error {
	t, err := template.New("index.html").
		Funcs(b.getTemplateFuncs()).
		ParseFiles("src/index.html")
	if err != nil {
		log.Fatalf("can't parse template: %v", err)
	}
	b.wg.Wait() // wait until b.Posts is populated
	b.SortPosts()

	err = t.Execute(b.indexPage, b.Posts)
	if err != nil {
		log.Fatalf("can't execute template: %v", err)
	}
	// todo: should i close file interface?
	return nil
}

func (b *Blog) getTemplateFuncs() template.FuncMap {
	b.meta = BlogMeta{}
	return template.FuncMap{
		"BlogLink":        b.meta.BlogLink,
		"BlogTitle":       b.meta.BlogTitle,
		"BlogDescription": b.meta.BlogDescription,
	}
}

func (b *Blog) SortPosts() {
	sort.Slice(b.Posts, func(i, j int) bool {
		return b.Posts[i].Date.Unix() > b.Posts[j].Date.Unix()
	})
}

func (b *Blog) verifyMarkdownPresent() error {
	if _, err := os.Stat(DefaultMarkdownPath); os.IsNotExist(err) {
		return err
	}
	return nil
}

func isFileValid(file os.FileInfo) bool {
	return path.Ext(file.Name()) == ".md" || path.Ext(file.Name()) == ".markdown"
}
