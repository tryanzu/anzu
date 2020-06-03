package config

import (
	"io/ioutil"
	"os"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/divideandconquer/go-merge/merge"
	"github.com/fsnotify/fsnotify"
	"github.com/hashicorp/hcl"
	"github.com/matcornic/hermes/v2"
	"github.com/op/go-logging"
)

var (
	// C stands for config
	C *Config

	// LoggingBackend packages should use.
	LoggingBackend logging.LeveledBackend
	log            = logging.MustGetLogger("config")
	format         = logging.MustStringFormatter(
		`%{color}%{time:15:04:05.000}  %{pid} %{module}	%{shortfile}	▶ %{level:.4s} %{id:03x}%{color:reset} %{message}`,
	)
)

func Bootstrap() {
	formatted := logging.NewBackendFormatter(logging.NewLogBackend(os.Stdout, "", 0), format)
	LoggingBackend = logging.AddModuleLevel(formatted)
	LoggingBackend.SetLevel(logging.INFO, "")
	log.SetBackend(LoggingBackend)

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

func (c *Config) Hermes() hermes.Hermes {
	if c.current == nil {
		return hermes.Hermes{}
	}
	site := c.current.Site
	return hermes.Hermes{
		Product: hermes.Product{
			Name:      site.Name,
			Link:      site.Url,
			Logo:      site.LogoUrl,
			Copyright: "Copyright © 2019 " + site.Name + ". Some rights reserved.",
		},
	}
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
		log.Info("error:", err)
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
	if level, err := logging.LogLevel(strings.ToUpper(c.current.Runtime.LoggingLevel)); err == nil {
		LoggingBackend.SetLevel(level, "")
		log.Noticef("logging level reloaded	level=%s", c.current.Runtime.LoggingLevel)
	}
	data, err := ioutil.ReadFile("./config.hcl")
	if err != nil {
		log.Error("Cannot load HCL configuration. Skipping")
	}

	var rules Rules
	err = hcl.Unmarshal(data, &rules)
	if err != nil {
		log.Error("Cannot unmarshal HCL configuration. Skipping")
	}

	c.rules = &rules
	log.Notice("config loaded from filesystem")
}

func (c *Config) Merge(file string, reload bool) {
	var config Anzu

	// Read the file first
	dat, err := ioutil.ReadFile(file)
	if err != nil {
		log.Info("error:", err)
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
				log.Info("error:", err)
			}
		}
	}()

	err = watcher.Add(file)
	if err != nil {
		log.Info("error:", err)
	}
}
