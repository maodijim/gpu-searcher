package platforms

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/tebeka/selenium"
	"gpu-searcher/config"
	"gpu-searcher/utils"
	"regexp"
	"time"
)

const (
	bbUrl = "https://bestbuy.com"
	bbUrlRegex = `^.*bestbuy.com`
)

type BestBuyPlatform struct {
	Wd	selenium.WebDriver
	PlatformName	string
	GpuTarget		[]config.GPUTarget
}

func (b *BestBuyPlatform) Search() (results []SearchResult) {
	for _, target := range b.GpuTarget {
		r := b.searchBestBuy(target)
		if len(r) > 0 {
			results = append(results, r...)
		}
	}
	return results
}

func (b *BestBuyPlatform) Close()  {
	b.Wd.Close()
}

func (b *BestBuyPlatform) searchBestBuy(gpu config.GPUTarget) []SearchResult {
	var results []SearchResult
	curUrl, _ := b.Wd.CurrentURL()
	if mat, _ := regexp.Match(bbUrlRegex, []byte(curUrl)); ! mat {
		log.Warnf("current url %s is not bestbuy.com going bestbuy", curUrl)
		_ = b.Wd.Get(bbUrl)
	}

	_ = b.Wd.WaitWithTimeout(waitModalElement, time.Second * 10)

	closeModal, err := b.Wd.FindElement(selenium.ByCSSSelector, "button.c-close-icon.c-modal-close-icon")
	if err != nil {
		log.Warnf("close bestbuy modal button not found")
	} else {
		_ = closeModal.Click()
	}

	time.Sleep(time.Second * 1)
	searchBar, err := b.Wd.FindElement(selenium.ByID, "gh-search-input")
	if err != nil {
		log.Errorf("%s failed to find search", BbPlatform)
		return results
	}
	clearBtn, err := b.Wd.FindElement(selenium.ByCSSSelector, "#header-clear-search-icon")
	if err == nil {
		_ = clearBtn.Click()
	}

	// Set search keyword
	searchBtn, err := b.Wd.FindElement(selenium.ByCSSSelector, "button.header-search-button")
	if err != nil {
		log.Errorf("%s failed to find search button", BbPlatform)
		return results
	}
	_ = searchBar.SendKeys(gpu.Name)
	time.Sleep(time.Second * 1)
	_ = searchBtn.Click()

	_ = b.Wd.Wait(waitSkuElement)

	// Get search items
	skuItems, err := b.Wd.FindElements(selenium.ByCSSSelector, "li.sku-item")

	for _, skuItem := range skuItems {
		result := SearchResult{
			Model: gpu.Name,
		}
		// Get price
		price, err := skuItem.FindElement(selenium.ByCSSSelector, "div.priceView-hero-price.priceView-customer-price")
		if err != nil {
			log.Errorf("failed to get price view")
		} else {
			txt, err := price.Text()
			if err != nil {
				log.Errorf("failed to get price from price view")
			} else {
				p := utils.PriceTextToFloat(txt)
				log.Infof("price: %f", p)
				result.Price = p
			}
		}

		// Get title
		titleSpan, err := skuItem.FindElement(selenium.ByCSSSelector, "div.sku-title")
		if err != nil {
			log.Errorf("failed to get titleSpan")
		} else {
			title, err := titleSpan.Text()
			if err != nil {
				log.Errorf("failed to get title from title span")
			}
			log.Infof("title: %s", title)
			result.Title = title
		}

		// Get product url
		href, err := skuItem.FindElement(selenium.ByCSSSelector, "h4.sku-header a")
		if err != nil {
			log.Errorf("failed to get product url: %s", err)
		} else {
			attr, err := href.GetAttribute("href")
			if err != nil {
				log.Errorf("failed to get hyper link for %s", result.Title)
			} else {
				result.Url = fmt.Sprintf("%s%s", bbUrl, attr)
			}
		}

		// Get stock status
		btnState, err := skuItem.FindElement(selenium.ByCSSSelector, "button.add-to-cart-button")
		if err != nil {
			log.Errorf("failed to get product state")
		} else {
			state, _ := btnState.GetAttribute("data-button-state")
			log.Infof("state: %s", state)
			if state == "SOLD_OUT" {
				result.Available = false
			} else if state == "ADD_TO_CART" {
				result.Available = true
			}
		}
		results = append(results, result)
	}
	return results
}

func waitSkuElement(wd selenium.WebDriver) (bool, error)  {
	we, err := wd.FindElement(selenium.ByCSSSelector, "li.sku-item")
	if err != nil {
		return false, err
	}
	display, _ := we.IsDisplayed()
	return display, err
}

func waitModalElement(wd selenium.WebDriver) (bool, error)  {
	we, err := wd.FindElement(selenium.ByCSSSelector, "button.c-close-icon.c-modal-close-icon")
	if err != nil {
		return false, err
	}
	display, _ := we.IsDisplayed()
	return display, err
}

func CreateBestBuySearch(caps map[string]interface{}, serverPort int, gpus []config.GPUTarget) Platform {
	p := BestBuyPlatform{
		PlatformName: BbPlatform,
		Wd: utils.CreateWebDriver(caps, serverPort),
		GpuTarget: gpus,
	}
	return &p
}
