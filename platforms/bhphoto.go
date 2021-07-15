package platforms

import (
	log "github.com/sirupsen/logrus"
	"github.com/tebeka/selenium"
	"gpu-searcher/config"
	"gpu-searcher/utils"
	"regexp"
	"strconv"
	"strings"
)

const (
	bhUrl      = "https://www.bhphotovideo.com/"
	bhUrlRegex = "^.*bhphotovideo.com"
)

// BhPlatform TODO implement b&h search
type BhPlatform struct {
	Wd           selenium.WebDriver
	PlatformName string
	GpuTarget    []config.GPUTarget
}

func (bh *BhPlatform) Close() {
	err := bh.Wd.Close()
	if err != nil {
		return
	}
	err = bh.Wd.Quit()
	if err != nil {
		return
	}
}

func (bh *BhPlatform) Search() (results []SearchResult) {
	for _, gpu := range bh.GpuTarget {
		r := bh.searchBh(gpu)
		results = append(results, r...)
	}
	return results
}

func (bh BhPlatform) searchBh(gpu config.GPUTarget) []SearchResult {
	var results []SearchResult
	curUrl, _ := bh.Wd.CurrentURL()
	if mat, _ := regexp.Match(bhUrlRegex, []byte(curUrl)); !mat {
		log.Warnf("current url is %s not %s redirecting", curUrl, bhUrl)
		err := bh.Wd.Get(bhUrl)
		if err != nil {
			log.Errorf("failed to go to %s: %s", bhUrl, err)
			return results
		}
	}
	searchBar, err := bh.Wd.FindElement(selenium.ByCSSSelector, "#top-search-input")
	if err != nil {
		log.Errorf("failed to find search bar: %s", err)
		return results
	}

	searchBtn, err := bh.Wd.FindElement(selenium.ByCSSSelector, "p.searchContainer button")
	if err != nil {
		log.Errorf("failed to find search button: %s", err)
		return results
	}

	_ = searchBar.Clear()
	_ = searchBar.Click()
	_ = searchBar.SendKeys(gpu.Name)
	_ = searchBtn.Click()

	items, err := bh.Wd.FindElements(selenium.ByCSSSelector, "div[data-selenium='miniProductPageProduct']")

	for _, item := range items {
		var result SearchResult

		// get product title
		title, err := item.FindElement(selenium.ByCSSSelector, "span[data-selenium='miniProductPageProductName']")
		if err != nil {
			log.Errorf("failed to get product title: %s", err)
		} else {
			result.Title, _ = title.Text()
		}

		// get product url
		productUrl, err := item.FindElement(selenium.ByCSSSelector, "a[data-selenium='miniProductPageProductNameLink']")
		if err != nil {
			log.Errorf("failed to find product url: %s", err)
		} else {
			result.Url, _ = productUrl.GetAttribute("href")
		}

		// get price
		upPriceFloat := float64(0)
		supPriceFloat := float64(0)
		upPrice, uPriceErr := item.FindElement(selenium.ByCSSSelector, "span[data-selenium='uppedDecimalPriceFirst']")
		if uPriceErr != nil {
			log.Errorf("failed to get product upper price: %s", err)
		}
		supPrice, supPriceErr := item.FindElement(selenium.ByCSSSelector, "sup[data-selenium='uppedDecimalPriceSecond']")
		if supPriceErr != nil {
			log.Errorf("failed to get product decimal price: %s", err)
		}
		if supPriceErr == nil && uPriceErr == nil {
			upPriceTxt, _ := upPrice.Text()
			supPriceTxt, _ := supPrice.Text()
			supPriceFloat, err = strconv.ParseFloat(supPriceTxt, 64)
			upPriceFloat = utils.PriceTextToFloat(upPriceTxt)
			if upPriceFloat > 0 && supPriceFloat > 0 {
				result.Price = upPriceFloat + supPriceFloat/100
			}
		}

		// get state
		stateContainer, err := item.FindElement(selenium.ByCSSSelector, "div[data-selenium='miniProductPageQuantityContainer']")
		if err != nil {
			log.Errorf("failed to get product state: %s", err)
		} else {
			stateTxt, _ := stateContainer.Text()
			if strings.ToLower(stateTxt) == "add to cart" {
				result.Available = true
			} else {
				result.Available = false
			}
		}

		if result.Price > 0 && result.Title != "" {
			result.Model = gpu.Name
			results = append(results, result)
		}
	}
	return results
}

// TODO Handle bot detection

// TODO multi-page scanning

func CreateBhSearch(caps map[string]interface{}, serverPort int, gpus []config.GPUTarget) Platform {
	p := BhPlatform{
		Wd:           utils.CreateWebDriver(caps, serverPort),
		PlatformName: BhPlatformName,
		GpuTarget:    gpus,
	}
	return &p
}
