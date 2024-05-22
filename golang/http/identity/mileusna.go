package identity

import "github.com/mileusna/useragent"

func NewMileusna() IsCrawler {
	return func(ua string) bool {
		return useragent.Parse(ua).Bot
	}
}
