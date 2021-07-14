package config

import (
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
	"os"
	"strconv"
	"strings"
)

type GPUTarget struct {
	Name	string	`yaml:"name"`
	Price	string	`yaml:"price"`
}

type Configs struct {
	Platforms         	[]string 	`yaml:"platforms"`
	FireFoxDriverPath 	string   	`yaml:"firefox_driver_path"`
	Headless			bool		`yaml:"headless"`
	GPUs				[]GPUTarget	`yaml:"gpus"`
	SeleniumJarPath		string		`yaml:"selenium_jar_path"`
	WindowSize			string		`yaml:"window_size"`
}

func (c *Configs) GetWindowSize() (x, y uint64) {
	x = 1920
	y = 1080
	xy := strings.Split(strings.ToLower(c.WindowSize), "x")
	if len(xy) != 2 {
		log.Warnf("invalid window size format %s using default size 1920x1080", c.WindowSize)
		return x, y
	}
	x, err := strconv.ParseUint(xy[0], 10, 64)
	if err != nil {
		log.Warnf("invalid window size x: %s using default size 1920x1080", err)
		return x, y
	}
	y, err = strconv.ParseUint(xy[0], 10, 64)
	if err != nil {
		log.Warnf("invalid window size y: %s using default size 1920x1080", err)
		return x, y
	}
	return x, y
}

func ParseConfigs(configPath string) (configs *Configs) {
	configs = &Configs{}
	f, err := os.Open(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			log.Fatalf("%s does not exist please ensure the configuration file path is correct", configPath)
		} else {
			log.Fatalf("Unknow error: %v", err)
		}
	}
	decoder := yaml.NewDecoder(f)
	err = decoder.Decode(configs)
	if err != nil {
		log.Fatalf("failed to parse yaml file: %s", err)
	}
	return configs
}
