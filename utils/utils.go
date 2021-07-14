package utils

import (
	"archive/zip"
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/tebeka/selenium"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"
)

var (
	ArchBitMap = map[string]int {
		"amd64": 64,
		"arm64": 64,
		"arm": 32,
		"386": 32,
		"aarch64": 64,
		"pine64": 64,
	}
	OSMap = map[string]string {
		"darwin": "macos",
		"linux": "linux",
		"windows": "win",
	}
)

func GetOSArch() string {
	return runtime.GOARCH
}

func GetOS() string {
	return runtime.GOOS
}

func ConstructOsArchStr() string {
	if GetOS() == "darwin" && GetOSArch() == "aarch64" {
		return "macos-aarch64"
	}
	return fmt.Sprintf("%s%d", OSMap[GetOS()], ArchBitMap[GetOSArch()])
}

func IsDriverExists(driverPath string) bool {
	if driverPath == "" {
		if GetOS() == "windows" {
			_, err := os.Stat("geckodriver.exe")
			if os.IsNotExist(err) {
				return false
			}
			return true
		} else {
			_, err := os.Stat("geckodriver")
			if os.IsNotExist(err) {
				return false
			}
			return true
		}
	} else {
		_, err := os.Stat(driverPath)
		if os.IsNotExist(err) {
			return false
		}
		return true
	}
}

func Unzip(src, dest string)  {
	reader, err := zip.OpenReader(src)
	if err != nil {
		log.Errorf("failed to unzip %s： %s", src, err)
	}
	defer reader.Close()

	for _, f := range reader.File {
		fPath := filepath.Join(dest, f.Name)
		if f.FileInfo().IsDir() {
			continue
		}
		of, err := os.OpenFile(fPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			log.Fatalf("failed to open file %s: %s", fPath, err)
		}

		r, err := f.Open()
		if err != nil {
			log.Fatalf("failed to read file %s: %s", f.Name, err)
		}

		_, err = io.Copy(of, r)
		of.Close()
		r.Close()
		if err != nil {
			log.Fatal("failed to save unzip file")
		}
	}

}

func PriceTextToFloat(priceText string) float64 {
	priceRegex := `\$[0-9.,]*`
	reg, _ := regexp.Compile(priceRegex)
	mat := reg.Find([]byte(priceText))
	if len(mat) < 1 {
		return 0
	}
	mat = mat[1:]
	price, err := strconv.ParseFloat(strings.Replace(string(mat), ",", "", -1), 64)
	if err != nil {
		return 0
	}
	return price
}

func CreateWebDriver(caps map[string]interface{}, serverPort int) selenium.WebDriver {
	wd, err := selenium.NewRemote(caps, fmt.Sprintf("http://localhost:%d/wd/hub", serverPort))
	if err != nil {
		panic(err)
	}
	return wd
}

