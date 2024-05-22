package identity

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGoogleBotIsRecognised(t *testing.T) {
	ua := "Mozilla/5.0 (Linux; Android 6.0.1; Nexus 5X Build/MMB29P) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/124.0.6367.201 Mobile Safari/537.36 (compatible; Googlebot/2.1; +http://www.google.com/bot.html)"
	isCrawler := NewMileusna()
	assert.NotNil(t, isCrawler)
	assert.True(t, isCrawler(ua))
}
