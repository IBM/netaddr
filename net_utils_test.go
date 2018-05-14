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

func TestParseIP(t *testing.T) {
	assert.Equal(t, net.ParseIP("0.0.0.0").To4(), ParseIP("0.0.0.0"))

	// The net package parses ipv4 as an ipv6 embedded v4. They aren't the
	// same so the netaddr package distinguishes between them.
	assert.Equal(t, net.ParseIP("10.0.0.1").To4(), ParseIP("10.0.0.1"))
	assert.Equal(t, net.ParseIP("10.0.0.1"), ParseIP("::ffff:10.0.0.1"))
	assert.NotEqual(t, net.ParseIP("10.0.0.1"), ParseIP("10.0.0.1"))
	assert.NotEqual(t, net.ParseIP("10.0.0.1").To4(), ParseIP("::ffff:10.0.0.1"))

	assert.Equal(t, net.ParseIP("2001:db8::1"), ParseIP("2001:db8::1"))
}

func TestNetIP(t *testing.T) {
	assert.Equal(t, net.ParseIP("0.0.0.0").To4(), NewIP(4))
	assert.Equal(t, net.ParseIP("::"), NewIP(16))
}

// Just a little shortcut for parsing a CIDR
func parse(str string) (n *net.IPNet) {
	_, parsed, err := net.ParseCIDR(str)
	if err == nil {
		n = parsed
	}
	return
}

func TestNetworkAddr(t *testing.T) {
	assert.Equal(t, ParseIP("203.0.113.0"), NetworkAddr(parse("203.0.113.0/24")))
	assert.Equal(t, ParseIP("10.0.0.0"), NetworkAddr(parse("10.0.0.0/16")))
	assert.Equal(t, ParseIP("10.1.64.0"), NetworkAddr(parse("10.1.66.3/18")))

	assert.Equal(t, ParseIP("2001:db8::"), NetworkAddr(parse("2001:db8::/64")))
	assert.Equal(t, ParseIP("2001:d00::"), NetworkAddr(parse("2001:db8::/24")))
}

func TestBroadcastAddr(t *testing.T) {
	assert.Equal(t, ParseIP("203.0.113.255"), BroadcastAddr(parse("203.0.113.0/24")))
	assert.Equal(t, ParseIP("10.0.255.255"), BroadcastAddr(parse("10.0.0.0/16")))
	assert.Equal(t, ParseIP("10.1.127.255"), BroadcastAddr(parse("10.1.66.3/18")))

	// IPv6 doesn't really have a broadcast address but it is still useful to
	// find the last address in a cidr
	assert.Equal(t, ParseIP("2001:db8::ffff:ffff:ffff:ffff"), BroadcastAddr(parse("2001:db8::/64")))
	assert.Equal(t, ParseIP("2001:dff:ffff:ffff:ffff:ffff:ffff:ffff"), BroadcastAddr(parse("2001:db8::/24")))
}
