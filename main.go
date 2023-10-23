package main

import (
	"fmt"
	"github.com/h2non/filetype"
	"io/fs"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func formatDuration(duration time.Duration) string {
	minutes := int(duration.Minutes())
	return fmt.Sprintf("%d:%02d", minutes/60, minutes%60)
}

func formatSize(size int64) string {
	const (
		_ = 1 << (10 * iota)
		KB
		MB
		GB
	)
	switch {
	case size >= GB:
		return fmt.Sprintf("%.2f GB", float64(size)/GB)
	case size >= MB:
		return fmt.Sprintf("%.2f MB", float64(size)/MB)
	case size >= KB:
		return fmt.Sprintf("%.2f KB", float64(size)/KB)
	default:
		return fmt.Sprintf("%d B", size)
	}
}

func listFiles(dirPath string, w http.ResponseWriter, r *http.Request, indent string) {
	fays := os.DirFS(dirPath)
	files, err := fs.ReadDir(fays, ".")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	for _, file := range files {
		relPath := filepath.Join(dirPath, file.Name())

		if file.IsDir() {
			fmt.Fprintf(w, "<tr>")
			fmt.Fprintf(w, "<td colspan=\"4\" style=\"text-decoration: none; color:#747474; font-size: 24px; font-family: Arial;\">📁 %s%s/</td>", indent, file.Name())
			fmt.Fprintf(w, "</tr>")
			listFiles(relPath, w, r, indent+"  ")
		} else if filepath.Base(relPath) != ".DS_Store" {
			buf, err := os.ReadFile(relPath)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			kind, err := filetype.Match(buf)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			// Get the relative path of the file
			relativePath, err := filepath.Rel(filepath.Join(os.Getenv("PWD"), "files"), relPath)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			// Replace backslashes with forward slashes for URL compatibility
			relativePath = strings.ReplaceAll(relativePath, `\`, `/`)

			// Get duration for video files
			duration := ""
			if kind.MIME.Type == "video" {
				duration = formatDuration(time.Minute * 5) // Replace with actual duration calculation
			}

			// Get file size
			fileInfo, err := os.Stat(relPath)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			size := formatSize(fileInfo.Size())

			extension := filepath.Ext(file.Name())

			fmt.Fprintf(w, "<tr>")
			fmt.Fprintf(w, "<td>%s<a href=\"%s\" style=\"text-decoration: none; color:#F7F7F7; font-size: 30px; font-family: Arial;\">📀️ %s</a></td>", indent, relativePath, strings.TrimSuffix(file.Name(), extension))
			fmt.Fprintf(w, "<td align=\"center\" style=\"text-decoration: none; color:#F7F7F7; font-size: 22px; font-family: Arial;\">%s</a></td>", extension)
			fmt.Fprintf(w, "<td align=\"center\" style=\"text-decoration: none; color:#F7F7F7; font-size: 22px; font-family: Arial;\">%s</td>", duration)
			fmt.Fprintf(w, "<td align=\"center\" style=\"text-decoration: none; color:#F7F7F7; font-size: 22px; font-family: Arial;\">%s</td>", size)
			fmt.Fprintf(w, "</tr>")
		}
	}
}

func main() {
	basePath, _ := os.Getwd()
	fileServer := http.FileServer(http.Dir(filepath.Join(basePath, "files")))
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			username, password, authOK := r.BasicAuth()
			if !authOK || username != "admin" || password != "4815162342" {
				w.Header().Set("WWW-Authenticate", `Basic realm="Restricted"`)
				w.WriteHeader(http.StatusUnauthorized)
				fmt.Fprintf(w, "Unauthorized access\n")
				return
			}

			fmt.Fprintf(w, "<html><head><title>Home Media Server</title></head><body bgcolor=\"#494949\">")
			fmt.Fprintf(w, "<h1 align=\"center\" style=\"color:#ABABAB;\">Home Media Server</h1>")
			fmt.Fprintf(w, "<table width=\"100%%\"><tr><th style=\"text-decoration: none; color:#9A9A9A; font-size: 18px; font-family: Arial;\">Name</th><th style=\"text-decoration: none; color:#9A9A9A; font-size:18px; font-family: Arial;\">Extension</th><th style=\"text-decoration: none; color:#9A9A9A; font-size: 18px; font-family: Arial;\">Duration</th><th style=\"text-decoration: none; color:#9A9A9A; font-size: 18px; font-family: Arial;\">Size</th></tr>")
			listFiles(filepath.Join(basePath, "files"), w, r, "")
			fmt.Fprintf(w, "</table></body></html>")
		} else {
			fileServer.ServeHTTP(w, r)
		}
	})
	log.Println("tiny Web Server Successfully started at port 8081")
	log.Fatal(http.ListenAndServe("0.0.0.0:8081", nil))
}
