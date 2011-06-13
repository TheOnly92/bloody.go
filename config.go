package main

import(
	"os"
	"path/filepath"
	"io/ioutil"
	"json"
)

type Config map[string]string

func (c *Config) Set(key string, val string) {
	(*c)[key] = val
}

func (c *Config) Is(key string) string {
	if val, exists := (*c)[key]; exists {
		return val
	}
	return ""
}

func (c *Config) Get(key string) string {
	if val, exists := (*c)[key]; exists {
		return val
	}
	return ""
}

func loadConfig() (config *Config) {
	root, _ := filepath.Split(filepath.Clean(os.Args[0]))
	b, err := ioutil.ReadFile(filepath.Join(root, "config.json"))
	if err != nil {
		panic(err)
		return &Config{}
	}
	err = json.Unmarshal(b, &config)
	return
}