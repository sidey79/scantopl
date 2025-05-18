package main

import (
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path"
	"strings"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/jnovack/flag"
	log "github.com/sirupsen/logrus"
)

const (
	Version       = "v1.0.1"
	PauseDuration = 1 * time.Second
)

var httpClient = &http.Client{Timeout: 10 * time.Second}

func init() {
	log.SetLevel(log.InfoLevel)
	log.SetFormatter(&log.TextFormatter{
		FullTimestamp: true,
	})
}

func FilenameWithoutExtension(fn string) string {
	return strings.TrimSuffix(path.Base(fn), path.Ext(fn))
}

func TitleFromFileName(fn string) string {
	return strings.TrimPrefix(FilenameWithoutExtension(fn), "pl_")
}

func createForm(form map[string]string) (string, io.Reader, error) {
	body := new(bytes.Buffer)
	mp := multipart.NewWriter(body)
	defer mp.Close()

	for key, val := range form {
		if strings.HasPrefix(val, "@") {
			val = val[1:]
			file, err := os.Open(val)
			if err != nil {
				return "", nil, err
			}
			defer file.Close()

			part, err := mp.CreateFormFile(key, val)
			if err != nil {
				return "", nil, err
			}
			io.Copy(part, file)
		} else {
			mp.WriteField(key, val)
		}
	}
	return mp.FormDataContentType(), body, nil
}

func uploadFile(document, plurl, pltoken string, removeFile func(string) error) {
	form := map[string]string{"document": "@" + document, "title": TitleFromFileName(document)}
	ct, body, err := createForm(form)
	if err != nil {
		log.Error("Error creating form:", err)
		return
	}
	log.Info("createForm reported error:", err)
	log.Info("createForm reported contentType:", ct)

	req, err := http.NewRequest("POST", plurl+"/api/documents/post_document/", body)
	if err != nil {
		log.Error("Error creating HTTP request:", err)
		return
	}

	req.Header.Set("Content-Type", ct)
	req.Header.Set("Authorization", "Token "+pltoken)

	resp, err := httpClient.Do(req)
	if err != nil {
		log.Error("Error uploading file:", err)
		return
	}

	if resp.StatusCode != http.StatusOK {
		log.Warn("Upload failed with response code:", resp.StatusCode)
	} else {
		log.Info("Upload successful, removing file:", document)
		err := removeFile(document)
		if err != nil {
			log.Warn("Error removing file:", err)
		}
	}
}

func watchDirectory(watcher *fsnotify.Watcher, done chan bool, plurl, pltoken string, removeFile func(string) error) {
	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				done <- true
				return
			}
			handleFileEvent(event, plurl, pltoken, removeFile)
		case err, ok := <-watcher.Errors:
			if !ok {
				done <- true
				return
			}
			log.Warn("Error watching directory:", err)
		}
	}
}

func handleFileEvent(event fsnotify.Event, plurl, pltoken string, removeFile func(string) error) {
	if event.Has(fsnotify.Create) && strings.HasPrefix(path.Base(event.Name), "pl_") {
		log.Info("New file to upload:", event.Name)
		time.Sleep(PauseDuration) // Consider a more robust solution for file readiness
		uploadFile(event.Name, plurl, pltoken, removeFile)
	}
}

func main() {
	flag.String(flag.DefaultConfigFlagname, "", "path to config file")
	scandir := flag.String("scandir", "/home/scanservjs/output", "Scanserjs output directory")
	plurl := flag.String("plurl", "http://localhost:8080", "The paperless instance URL without trailing /")
	pltoken := flag.String("pltoken", "xxxxxxxxxxxxxxxxxx", "Paperless auth token, generated through admin")
	showversion := flag.Bool("version", false, "Show version and exit")

	flag.Parse()

	if *showversion {
		fmt.Println("version", Version)
		os.Exit(0)
	}

	log.Println("Start watching:", *scandir)

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal("Error creating watcher:", err)
	}
	defer watcher.Close()

	done := make(chan bool)
	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					done <- true
					return
				}
				handleFileEvent(event, *plurl, *pltoken, os.Remove)
			case err, ok := <-watcher.Errors:
				if !ok {
					done <- true
					return
				}
				log.Warn("Error watching directory:", err)
			}
		}
	}()

	err = watcher.Add(*scandir)
	if err != nil {
		log.Fatal("Error adding directory to watcher:", err)
	}
	<-done
}

