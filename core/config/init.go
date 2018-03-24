package config

import (
	"io/ioutil"
	"log"

	"github.com/fsnotify/fsnotify"
	"github.com/hjson/hjson-go"
	"github.com/imdario/mergo"
)

var (
	// C stands for config
	C *Config
)

func Bootstrap() {
	C = new(Config)
	C.Reload = make(chan bool)
	C.Merge("./static/resources/config.hjson")
	C.Merge("./config.hjson")

	// Watch config file
	go C.WatchFile("./config.hjson")
}

type Config struct {
	Reload  chan bool
	current *map[string]interface{}
}

func (c *Config) Copy() map[string]interface{} {
	return *c.current
}

func (c *Config) Merge(file string) {
	var config map[string]interface{}

	// Read the file first
	dat, err := ioutil.ReadFile(file)
	if err != nil {
		log.Println("error:", err)
		return
	}

	if err := hjson.Unmarshal(dat, &config); err != nil {
		panic(err)
	}

	if c.current == nil {
		c.current = new(map[string]interface{})
	}

	// Clone the current runtime config map.
	merged := make(map[string]interface{}, len(config))
	for k, v := range config {
		merged[k] = v
	}

	if err := mergo.Merge(&merged, *c.current, mergo.WithOverride); err != nil {
		panic(err)
	}

	c.current = &merged

	// Reload signal if anyone is listening...
	go func() {
		c.Reload <- true
	}()

	log.Println("configured: ", c.current)
	log.Printf("address: %p\n", c.current)
}

func (c *Config) WatchFile(file string) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}

	go func() {
		defer watcher.Close()

		for {
			select {
			case event := <-watcher.Events:
				log.Println("event:", event)
				if event.Op&fsnotify.Write == fsnotify.Write {
					log.Println("modified file:", event.Name)
					c.Merge(event.Name)
				}
			case err := <-watcher.Errors:
				log.Println("error:", err)
			}
		}
	}()

	err = watcher.Add(file)
	if err != nil {
		log.Println("error:", err)
	}
}
