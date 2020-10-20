package internal

import (
	"fmt"
	"underblog/internal/watcher"
	"log"
	"net/http"
	"time"
)

func WatchForChangedFiles(rebuildBlog func()) {
	w := watcher.New()

	go func() {
		for {
			select {
			case <-w.Event:
				rebuildBlog()
			case err := <-w.Error:
				log.Fatalln(err)
			case <-w.Closed:
				return
			}
		}
	}()

	// Watch these files for changes.
	files := []string{"./src/index.html", "./src/post.html"}
	for _, file := range files {
		if err := w.Add(file); err != nil {
			log.Fatalln(err)
		}
	}

	directories := []string{"./src/markdown", "./src/static"}
	for _, directory := range directories {
		if err := w.AddRecursive(directory); err != nil {
			log.Fatalln(err)
		}
	}

	if err := w.Start(time.Millisecond * 200); err != nil {
		log.Fatalln(err)
	}
}

func RunDevelopmentWebserver() {
	// todo: extract ./public to constant
	http.Handle("/", http.FileServer(http.Dir("./dist")))
	// todo: check if port is not occupied
	// todo: change port via cli argument
	port := 8080
	fmt.Printf("Development webserver is available at http://%s:%v/\n", "localhost", port)
	addr := fmt.Sprintf(":%v", port)
	if err := http.ListenAndServe(addr, nil); err != nil {
		panic(err)
	}
}
