package platforms

import (
	log "github.com/sirupsen/logrus"
	"github.com/tebeka/selenium"
	"gpu-searcher/config"
	"gpu-searcher/utils"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const (
	neUrl      = "https://www.newegg.com"
	neUrlRegex = `^.*newegg.com`
)

type NewEggPlatform struct {
	Wd           selenium.WebDriver
	PlatformName string
	GpuTarget    []config.GPUTarget
}

func (n *NewEggPlatform) Search() (results []SearchResult) {
	for _, gpu := range n.GpuTarget {
		newResults := n.searchNewEgg(gpu)
		results = append(results, newResults...)
	}
	return results
}

func (n *NewEggPlatform) Close() {
	err := n.Wd.Close()
	if err != nil {
		return
	}
	err = n.Wd.Quit()
	if err != nil {
		return
	}
}

func (n *NewEggPlatform) searchNewEgg(gpu config.GPUTarget) []SearchResult {
	var results []SearchResult
	curUrl, _ := n.Wd.CurrentURL()
	if mat, _ := regexp.Match(neUrlRegex, []byte(curUrl)); !mat {
		log.Infof("current url %s is not newegg.com redirecting to newegg.com", curUrl)
		err := n.Wd.Get(neUrl)
		if err != nil {
			log.Errorf("failed to open %s: %s", neUrl, err)
			return results
		}
	}

	searchBar, err := n.Wd.FindElement(selenium.ByCSSSelector, "div.header2020-search-box input")
	if err != nil {
		log.Errorf("failed to find search bar")
		return results
	}

	searchBtn, err := n.Wd.FindElement(selenium.ByCSSSelector, "button.fas")
	if err != nil {
		log.Errorf("failed to find search button")
		return results
	}

	_ = searchBar.Clear()
	_ = searchBar.SendKeys(gpu.Name)
	time.Sleep(time.Second)
	_ = searchBtn.Click()

	items, err := n.Wd.FindElements(selenium.ByCSSSelector, "div.item-container")
	if err != nil {
		log.Errorf("failed to find search items: %s", err)
	}

	for _, item := range items {
		result := SearchResult{}
		title, err := item.FindElement(selenium.ByCSSSelector, "a.item-title")
		if err != nil {
			log.Errorf("failed to find item title: %s", err)
		} else {
			result.Title, _ = title.Text()
			result.Url, _ = title.GetAttribute("href")
		}

		// Get the price
		priceBox, err := item.FindElement(selenium.ByCSSSelector, "li.price-current")
		if err != nil {
			log.Errorf("failed to find current price : %s", err)
		} else {
			priceMain, err := priceBox.FindElement(selenium.ByCSSSelector, "strong")
			if err != nil {
				log.Errorf("failed to find main price: %s", err)
			}
			priceSub, err := priceBox.FindElement(selenium.ByCSSSelector, "sup")
			if err != nil {
				log.Errorf("failed to find sub price: %s", err)
			}
			mText, _ := priceMain.Text()
			sText, _ := priceSub.Text()
			mFloat, _ := strconv.ParseFloat(strings.Replace(mText, ",", "", -1), 64)
			sFloat, _ := strconv.ParseFloat(sText, 64)
			result.Price = mFloat + sFloat
		}

		// Get current item state
		itemBtn, err := item.FindElement(selenium.ByCSSSelector, "div.item-button-area button")
		if err != nil {
			log.Errorf("failed to find button state area: %s", err)
		} else {
			btnText, _ := itemBtn.Text()
			promoElem, _ := item.FindElement(selenium.ByCSSSelector, "p.item-promo")
			promoTxt := ""
			if promoElem != nil {
				promoTxt, _ = promoElem.Text()
			}
			if promoTxt == "OUT OF STOCK" {
				result.Available = false
			} else if strings.ToLower(btnText) == "view details" || strings.ToLower(btnText) == "add to cart" {
				result.Available = true
			}
		}
		if result.Title != "" && result.Price > 0 {
			result.Model = gpu.Name
			results = append(results, result)
		}
	}
	return results
}

// TODO multi-page scanning

func CreateNewEggSearch(caps map[string]interface{}, serverPort int, gpus []config.GPUTarget) Platform {
	p := NewEggPlatform{
		PlatformName: NePlatformName,
		Wd:           utils.CreateWebDriver(caps, serverPort),
		GpuTarget:    gpus,
	}
	return &p
}
