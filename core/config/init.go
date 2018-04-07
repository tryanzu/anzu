package config

import (
	"io/ioutil"
	"log"

	"github.com/divideandconquer/go-merge/merge"
	"github.com/fsnotify/fsnotify"
	"github.com/hjson/hjson-go"
)

var (
	// C stands for config
	C *Config
)

func Bootstrap() {
	C = new(Config)
	C.Reload = make(chan struct{})
	C.Boot()

	// Watch config file
	go C.WatchFile("./config.hjson")
}

type Config struct {
	Reload  chan struct{}
	current *map[string]interface{}
}

func (c *Config) Copy() map[string]interface{} {
	return *c.current
}

func (c *Config) Boot() {
	c.current = nil
	c.Merge("./static/resources/config.hjson", false)
	c.Merge("./config.hjson", true)
	log.Println("Config from filesystem loaded.")
}

func (c *Config) Merge(file string, reload bool) {
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
	merged := merge.Merge(*c.current, config)
	cst := merged.(map[string]interface{})
	c.current = &cst

	// Reload signal if anyone is listening...
	if reload {
		close(c.Reload)
		c.Reload = make(chan struct{})
	}
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
				if event.Op&fsnotify.Write == fsnotify.Write {
					c.Boot()
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
