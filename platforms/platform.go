package platforms

const (
	BbPlatform = "bestbuy"
	NePlatform = "newegg"
	BhPlatform = "b&h"
)

var (
	SupportedPlatforms = []string{
		BbPlatform,
		NePlatform,
		BhPlatform,
	}
)

type SearchResult struct {
	Price	float64
	Title	string
	Model	string
	Url		string
	Available	bool
}

type Platform interface {
	Search() []SearchResult
	Close()
}
