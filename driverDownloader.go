package main

import (
	"encoding/json"
	"fmt"
	log "github.com/sirupsen/logrus"
	"gpu-searcher/utils"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
)

const (
	gitUrl = "https://github.com/mozilla/geckodriver/releases"
)

type assetFormat struct {
	Name				string	`json:"name"`
	BrowserDownloadUrl	string	`json:"browser_download_url"`
}

type resFormat struct {
	Assets []assetFormat `json:"assets"`

}

// Retrieve latest driver download url from github it's the same as https://github.com/mozilla/geckodriver/releases
func getLatestDriverUrl() string {
	res, err := http.Get("https://api.github.com/repos/mozilla/geckodriver/releases/latest")
	if err != nil {
		log.Fatalf("failed to get latest firefox driver url: %s", err)
	}
	resData := &resFormat{}
	body, err := ioutil.ReadAll(res.Body)
	defer res.Body.Close()
	if err != nil {
		log.Fatalf("failed to read body: %s", err)
	}

	err = json.Unmarshal(body, resData)
	if err != nil {
		log.Fatalf("failed to parse reponse json body: %s", err)
	}

	regex := fmt.Sprintf(`geckodriver-v[0-9.]*-%s.(tar.gz|zip)`, utils.ConstructOsArchStr())
	for _, item := range resData.Assets {
		if mat, _ := regexp.Match(regex, []byte(item.Name)); mat {
			log.Infof("driver matching current OS %s found: %s", utils.ConstructOsArchStr(), item.Name)
			return item.BrowserDownloadUrl
		}
	}
	return ""
}

func downloadDriver(dest string) string {
	url := getLatestDriverUrl()
	if url != "" {
		savePath := filepath.Join(dest, filepath.Base(url))
		res, err := http.Get(url)
		if err != nil {
			log.Fatalf("failed to download driver please download firefox driver manully from %s: %s", gitUrl, err)
		}
		defer res.Body.Close()
		of, err := os.Create(savePath)
		if err != nil {
			log.Fatalf("failed to create %s: %s", savePath, err)
		}
		defer of.Close()
		_, err = io.Copy(of, res.Body)
		if err != nil {
			log.Fatalf("failed to write %s: %s", savePath, err)
		}
		return savePath
	} else {
		log.Warnf("latest firefox driver not found please download it manully")
	}
	return ""
}
