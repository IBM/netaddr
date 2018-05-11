package netaddr

import (
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExpandNet(t *testing.T) {
	_, n, _ := net.ParseCIDR("203.0.113.0/29")
	ips := expandNet(n, 10)
	assert.Equal(t, 8, len(ips))
	assert.Equal(t, net.ParseIP("203.0.113.0").To4(), ips[0])
	assert.Equal(t, net.ParseIP("203.0.113.7").To4(), ips[7])
}

func TestExpandNetLimit(t *testing.T) {
	_, n, _ := net.ParseCIDR("203.0.113.0/29")
	ips := expandNet(n, 5)
	assert.Equal(t, 5, len(ips))
	assert.Equal(t, net.ParseIP("203.0.113.0").To4(), ips[0])
	assert.Equal(t, net.ParseIP("203.0.113.4").To4(), ips[4])
}

func TestExpandNetLarge(t *testing.T) {
	_, n, _ := net.ParseCIDR("2001:db8::/56")
	ips := expandNet(n, 1000)
	assert.Equal(t, 1000, len(ips))
	assert.Equal(t, net.ParseIP("2001:db8::0"), ips[0])
	assert.Equal(t, net.ParseIP("2001:db8::100"), ips[256])
	assert.Equal(t, net.ParseIP("2001:db8::3e7"), ips[999])
}
