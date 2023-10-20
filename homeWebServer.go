package main

import (
	"fmt"
	"github.com/h2non/filetype"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

func getIcon(ext string) string {
	switch ext {
	case "jpg", "png", "gif", "jpeg":
		return "/assets/icons/images.png"
	case "mp4", "avi", "mkv":
		return "/assets/icons/videos.png"
	default:
		return "/assets/icons/unknown.png"
	}
}

func listFiles(w http.ResponseWriter, r *http.Request) {
	basePath, _ := os.Getwd() // Get the current working directory
	files, err := filepath.Glob(filepath.Join(basePath, "files", "*"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	for _, file := range files {
		relPath, _ := filepath.Rel(basePath, file)

		// Check if the file is .DS_Store
		if filepath.Base(relPath) == ".DS_Store" {
			continue
		}

		// Check if the path is a directory
		if info, err := os.Stat(file); err == nil && !info.IsDir() {
			buf, err := ioutil.ReadFile(file)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			kind, err := filetype.Match(buf)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			fmt.Fprintf(w, "<html><head><title>mkv73's Media Server</title></head><body>")
			fmt.Fprintf(w, "<a href=\"/%s\" style=\"text-decoration: none;\"><img src=\"%s\"> %s</a><br>", relPath, getIcon(kind.Extension), strings.TrimSuffix(filepath.Base(file), filepath.Ext(file)))
			fmt.Fprintf(w, "</body></html>")
			log.Println(relPath)
		}
	}
}

func main() {
	basePath, _ := os.Getwd() // Get the current working directory
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.FileServer(http.Dir(filepath.Join(basePath, "./"))).ServeHTTP(w, r)
	})
	http.HandleFunc("/index", listFiles)
	log.Fatal(http.ListenAndServe("localhost:8081", nil))
}
