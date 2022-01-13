package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path"
	"path/filepath"
	"time"

	"github.com/fsnotify/fsnotify"
	"gopkg.in/yaml.v2"
)

type TConfig struct {
	Target      string `yaml:"target"`
	Destination string `yaml:"destination"`
}

func main() {
	data, err := ioutil.ReadFile("./config.yaml")
	if err != nil {
		log.Fatal(err)
		return
	}

	config := TConfig{}
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		log.Fatal(err)
		return
	}

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Close()

	done := make(chan bool)
	fsEvent := make(chan fsnotify.Event)
	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				fsEvent <- event
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				log.Println("error:", err)
			}
		}
	}()

	go func() {
		for {
			event := <-fsEvent
			if event.Op&fsnotify.Write == fsnotify.Write {
				for ok := false; !ok; ok = CopyFile(event.Name, config) {
					time.Sleep(time.Second * 1)
				}
			}
		}
	}()

	err = watcher.Add(config.Target)
	if err != nil {
		log.Fatal(err)
	}
	<-done
}

func CopyFile(src_path string, config TConfig) bool {
	dst_path := path.Join(config.Destination, filepath.Base(src_path))
	fmt.Printf("%s %s\n", dst_path, filepath.Base(src_path))
	src, err := os.Open(src_path)
	if err != nil {
		log.Fatal(err)
		return false
	}
	defer src.Close()

	dst, err := os.Create(dst_path)
	if err != nil {
		log.Fatal(err)
		return false
	}
	defer dst.Close()

	_, err = io.Copy(dst, src)
	if err != nil {
		log.Fatal(err)
		return false
	}
	log.Printf("Copied: %s -> %s", src_path, dst_path)

	return true
}
