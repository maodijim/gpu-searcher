package main

import (
	"flag"
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/tebeka/selenium"
	"github.com/tebeka/selenium/firefox"
	"gpu-searcher/config"
	"gpu-searcher/platforms"
	"gpu-searcher/utils"
)

const (
	port = 8088
)

func init() {
	log.SetReportCaller(true)
	log.SetFormatter(&log.TextFormatter{
		ForceColors:   true,
		FullTimestamp: true,
	})
}

func main() {
	confPath := flag.String("configPath", "config.template", "path to the config.conf file")
	flag.Parse()
	configs := config.ParseConfigs(*confPath)
	if !utils.IsDriverExists(configs.FireFoxDriverPath) {
		log.Infof("firefox driver not found download latest firefox dirver")
		savePath := downloadDriver("")
		utils.Unzip(savePath, "")
	}

	opts := []selenium.ServiceOption{
		//selenium.StartFrameBuffer(),
		selenium.GeckoDriver(configs.FireFoxDriverPath),
		//selenium.Output(os.Stderr),
	}
	service, err := selenium.NewSeleniumService(configs.SeleniumJarPath, port, opts...)
	if err != nil {
		panic(err)
	}

	defer func(service *selenium.Service) {
		err := service.Stop()
		if err != nil {
			return
		}
	}(service)

	caps := selenium.Capabilities{"browserName": "firefox"}
	firefoxCaps := firefox.Capabilities{}
	x, y := configs.GetWindowSize()
	firefoxCaps.Args = []string{
		fmt.Sprintf("--window-size=%dx%d", x, y),
	}
	if configs.Headless {
		firefoxCaps.Args = append(firefoxCaps.Args, "--headless")
		caps.AddFirefox(firefoxCaps)
	}

	var platformDrivers []platforms.Platform

	// Create web driver for each platform
	for _, p := range configs.Platforms {
		switch p {
		case platforms.BbPlatformName:
			bb := platforms.CreateBestBuySearch(caps, port, configs.GPUs)
			platformDrivers = append(platformDrivers, bb)
		case platforms.NePlatformName:
			ne := platforms.CreateNewEggSearch(caps, port, configs.GPUs)
			platformDrivers = append(platformDrivers, ne)
		case platforms.BhPlatformName:
			bh := platforms.CreateBhSearch(caps, port, configs.GPUs)
			platformDrivers = append(platformDrivers, bh)
		default:
			log.Warnf("unsupport platform %s, skipped", p)
		}
	}

	// Close all web drivers
	defer func() {
		for _, d := range platformDrivers {
			d.Close()
		}
	}()

	for _, d := range platformDrivers {
		results := d.Search()
		log.Infof("results: %v", results)
	}

	// TODO Implement email notification
}
