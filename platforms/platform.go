package platforms

const (
	BbPlatformName = "bestbuy"
	NePlatformName = "newegg"
	BhPlatformName = "b&h"
)

var (
	SupportedPlatforms = []string{
		BbPlatformName,
		NePlatformName,
		BhPlatformName,
	}
)

type SearchResult struct {
	Price     float64
	Title     string
	Model     string
	Url       string
	Available bool
}

type Platform interface {
	Search() []SearchResult
	Close()
}
