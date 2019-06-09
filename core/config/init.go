package config

import (
	"io/ioutil"
	"log"

	"github.com/BurntSushi/toml"
	"github.com/divideandconquer/go-merge/merge"
	"github.com/fsnotify/fsnotify"
	"github.com/hashicorp/hcl"
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
	go C.WatchFile("./config.toml")
	go C.WatchFile("./config.hcl")
}

type Config struct {
	Reload  chan struct{}
	current *Anzu
	rules   *Rules
}

func (c *Config) Copy() Anzu {
	if c.current == nil {
		return Anzu{}
	}
	return *c.current
}

func (c *Config) Rules() Rules {
	if c.rules == nil {
		return Rules{}
	}
	return *c.rules
}

func (c *Config) UserCopy() (conf map[string]interface{}) {

	// Read the file first
	dat, err := ioutil.ReadFile("./config.toml")
	if err != nil {
		log.Println("error:", err)
		return
	}

	if err := toml.Unmarshal(dat, &conf); err != nil {
		panic(err)
	}

	return
}

func (c *Config) Boot() {
	c.current = nil
	c.Merge("./static/resources/config.toml", false)
	c.Merge("./config.toml", true)
	data, err := ioutil.ReadFile("./config.hcl")
	if err != nil {
		log.Println("Cannot load HCL configuration. Skipping")
	}

	var rules Rules
	err = hcl.Unmarshal(data, &rules)
	if err != nil {
		log.Println("Cannot unmarshal HCL configuration. Skipping")
	}

	c.rules = &rules
	log.Println("Config from filesystem loaded.")
}

func (c *Config) Merge(file string, reload bool) {
	var config Anzu

	// Read the file first
	dat, err := ioutil.ReadFile(file)
	if err != nil {
		log.Println("error:", err)
		return
	}

	if err := toml.Unmarshal(dat, &config); err != nil {
		panic(err)
	}

	if c.current == nil {
		c.current = new(Anzu)
	}

	// Clone the current runtime config map.
	merged := merge.Merge(*c.current, config)
	cst := merged.(*Anzu)
	c.current = cst

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
