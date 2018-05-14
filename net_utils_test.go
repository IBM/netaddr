package netaddr

import (
	"math/big"
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

func TestNetSize(t *testing.T) {
	_, n, _ := net.ParseCIDR("10.0.0.0/24")
	assert.Equal(t, int64(256), NetSize(n).Int64())
}

func TestNetSizeHost(t *testing.T) {
	_, n, _ := net.ParseCIDR("203.0.113.29/32")
	assert.Equal(t, int64(1), NetSize(n).Int64())
}

func TestNetSizeSlash8(t *testing.T) {
	_, n, _ := net.ParseCIDR("15.0.0.0/8")
	assert.Equal(t, int64(16777216), NetSize(n).Int64())
}

func TestNetSizeV6(t *testing.T) {
	_, n, _ := net.ParseCIDR("2001:db8::/64")
	assert.Equal(t, big.NewInt(0).Lsh(big.NewInt(1), 64), NetSize(n))
}

func TestNetSizeV6Huge(t *testing.T) {
	_, n, _ := net.ParseCIDR("2001:db8::/8")
	assert.Equal(t, big.NewInt(0).Lsh(big.NewInt(1), 120), NetSize(n))
}

func TestNetSizeV6Host(t *testing.T) {
	_, n, _ := net.ParseCIDR("2001:db8::1/128")
	assert.Equal(t, big.NewInt(1), NetSize(n))
}
