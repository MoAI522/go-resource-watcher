package main

import (
	"io"
	"io/ioutil"
	"log"
	"os"
	"path"

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
	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				switch event.Op {
				case fsnotify.Write:
					src_path := event.Name
					dst_path := path.Join(config.Destination, path.Base(event.Name))
					src, err := os.Open(src_path)
					if err != nil {
						log.Fatal(err)
						return
					}
					defer src.Close()

					dst, err := os.Create(dst_path)
					if err != nil {
						log.Fatal(err)
						return
					}
					defer dst.Close()

					_, err = io.Copy(dst, src)
					if err != nil {
						log.Fatal(err)
						return
					}
					log.Printf("Copied: %s -> %s", src_path, dst_path)
					break
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				log.Println("error:", err)
			}
		}
	}()

	err = watcher.Add(config.Target)
	if err != nil {
		log.Fatal(err)
	}
	<-done
}
