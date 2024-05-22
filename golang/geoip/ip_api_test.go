package geoip

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFindCloudflareIP(t *testing.T) {
	find := NewIPApiFinder()
	assert.NotNil(t, find)
	l, err := find("1.1.1.1")
	assert.NoError(t, err)
	assert.Contains(t, l.Isp, "Cloudflare")
}

func TestFindLocalhost(t *testing.T) {
	find := NewIPApiFinder()
	assert.NotNil(t, find)
	_, err := find("127.0.0.1")
	assert.Error(t, err)
}
