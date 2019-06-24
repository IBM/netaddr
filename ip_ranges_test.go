package netaddr

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIPRangeFromCIDR(t *testing.T) {
	net, err := ParseNet("10.0.0.0/8")
	assert.Nil(t, err)

	ipRange := IPRangeFromCIDR(net)

	assert.Equal(t, "10.0.0.0", ipRange.First.String())
	assert.Equal(t, "10.255.255.255", ipRange.Last.String())
	assert.Equal(t, "[10.0.0.0,10.255.255.255]", ipRange.String())
}
func TestIPRangeFromCIDRIPv6(t *testing.T) {
	net, err := ParseNet("2001::/16")
	assert.Nil(t, err)

	ipRange := IPRangeFromCIDR(net)

	assert.Equal(t, "2001::", ipRange.First.String())
	assert.Equal(t, "2001:ffff:ffff:ffff:ffff:ffff:ffff:ffff", ipRange.Last.String())
	assert.Equal(t, "[2001::,2001:ffff:ffff:ffff:ffff:ffff:ffff:ffff]", ipRange.String())
}

func TestIPRangeDifference(t *testing.T) {
	for _, tc := range []struct {
		a, b   *IPRange
		result string
	}{
		// Give ranges [10.0.0.0-10.0.0.25], try every overlap case {9.0.0.0, 10.0.0.0, 10.0.0.24, 10.0.0.255, 11.0.0.0}
		{&IPRange{ParseIP("10.0.0.0"), ParseIP("10.0.0.255")}, &IPRange{ParseIP("9.0.0.0"), ParseIP("9.0.0.0")}, "[[10.0.0.0,10.0.0.255]]"},
		{&IPRange{ParseIP("10.0.0.0"), ParseIP("10.0.0.255")}, &IPRange{ParseIP("9.0.0.0"), ParseIP("10.0.0.0")}, "[[10.0.0.1,10.0.0.255]]"},
		{&IPRange{ParseIP("10.0.0.0"), ParseIP("10.0.0.255")}, &IPRange{ParseIP("9.0.0.0"), ParseIP("10.0.0.24")}, "[[10.0.0.25,10.0.0.255]]"},
		{&IPRange{ParseIP("10.0.0.0"), ParseIP("10.0.0.255")}, &IPRange{ParseIP("9.0.0.0"), ParseIP("10.0.0.255")}, "[]"},
		{&IPRange{ParseIP("10.0.0.0"), ParseIP("10.0.0.255")}, &IPRange{ParseIP("9.0.0.0"), ParseIP("11.0.0.0")}, "[]"},
		{&IPRange{ParseIP("10.0.0.0"), ParseIP("10.0.0.255")}, &IPRange{ParseIP("10.0.0.0"), ParseIP("10.0.0.0")}, "[[10.0.0.1,10.0.0.255]]"},
		{&IPRange{ParseIP("10.0.0.0"), ParseIP("10.0.0.255")}, &IPRange{ParseIP("10.0.0.0"), ParseIP("10.0.0.24")}, "[[10.0.0.25,10.0.0.255]]"},
		{&IPRange{ParseIP("10.0.0.0"), ParseIP("10.0.0.255")}, &IPRange{ParseIP("10.0.0.0"), ParseIP("10.0.0.255")}, "[]"},
		{&IPRange{ParseIP("10.0.0.0"), ParseIP("10.0.0.255")}, &IPRange{ParseIP("10.0.0.0"), ParseIP("11.0.0.0")}, "[]"},
		{&IPRange{ParseIP("10.0.0.0"), ParseIP("10.0.0.255")}, &IPRange{ParseIP("10.0.0.24"), ParseIP("10.0.0.24")}, "[[10.0.0.0,10.0.0.23] [10.0.0.25,10.0.0.255]]"},
		{&IPRange{ParseIP("10.0.0.0"), ParseIP("10.0.0.255")}, &IPRange{ParseIP("10.0.0.24"), ParseIP("10.0.0.255")}, "[[10.0.0.0,10.0.0.23]]"},
		{&IPRange{ParseIP("10.0.0.0"), ParseIP("10.0.0.255")}, &IPRange{ParseIP("10.0.0.24"), ParseIP("11.0.0.0")}, "[[10.0.0.0,10.0.0.23]]"},
		{&IPRange{ParseIP("10.0.0.0"), ParseIP("10.0.0.255")}, &IPRange{ParseIP("10.0.0.255"), ParseIP("10.0.0.255")}, "[[10.0.0.0,10.0.0.254]]"},
		{&IPRange{ParseIP("10.0.0.0"), ParseIP("10.0.0.255")}, &IPRange{ParseIP("10.0.0.255"), ParseIP("11.0.0.0")}, "[[10.0.0.0,10.0.0.254]]"},
		{&IPRange{ParseIP("10.0.0.0"), ParseIP("10.0.0.255")}, &IPRange{ParseIP("11.0.0.0"), ParseIP("11.0.0.0")}, "[[10.0.0.0,10.0.0.255]]"},

		{&IPRange{ParseIP("10.0.0.0"), ParseIP("10.0.0.0")}, &IPRange{ParseIP("10.0.0.0"), ParseIP("10.0.0.0")}, "[]"},
	} {
		diff := tc.a.Minus(tc.b)
		assert.Equal(t, tc.result, fmt.Sprintf("%s", diff))
	}
}

func TestIPRangeContains(t *testing.T) {
	for i, tc := range []struct {
		a, b   *IPRange
		result bool
	}{
		// Give ranges [10.0.0.0-10.0.0.25], try every overlap case {9.0.0.0, 10.0.0.0, 10.0.0.24, 10.0.0.255, 11.0.0.0}
		{&IPRange{ParseIP("10.0.0.0"), ParseIP("10.0.0.255")}, &IPRange{ParseIP("9.0.0.0"), ParseIP("9.0.0.0")}, false},
		{&IPRange{ParseIP("10.0.0.0"), ParseIP("10.0.0.255")}, &IPRange{ParseIP("9.0.0.0"), ParseIP("10.0.0.0")}, false},
		{&IPRange{ParseIP("10.0.0.0"), ParseIP("10.0.0.255")}, &IPRange{ParseIP("9.0.0.0"), ParseIP("10.0.0.24")}, false},
		{&IPRange{ParseIP("10.0.0.0"), ParseIP("10.0.0.255")}, &IPRange{ParseIP("9.0.0.0"), ParseIP("10.0.0.255")}, false},
		{&IPRange{ParseIP("10.0.0.0"), ParseIP("10.0.0.255")}, &IPRange{ParseIP("9.0.0.0"), ParseIP("11.0.0.0")}, false},
		{&IPRange{ParseIP("10.0.0.0"), ParseIP("10.0.0.255")}, &IPRange{ParseIP("10.0.0.0"), ParseIP("10.0.0.0")}, true},
		{&IPRange{ParseIP("10.0.0.0"), ParseIP("10.0.0.255")}, &IPRange{ParseIP("10.0.0.0"), ParseIP("10.0.0.24")}, true},
		{&IPRange{ParseIP("10.0.0.0"), ParseIP("10.0.0.255")}, &IPRange{ParseIP("10.0.0.0"), ParseIP("10.0.0.255")}, true},
		{&IPRange{ParseIP("10.0.0.0"), ParseIP("10.0.0.255")}, &IPRange{ParseIP("10.0.0.0"), ParseIP("11.0.0.0")}, false},
		{&IPRange{ParseIP("10.0.0.0"), ParseIP("10.0.0.255")}, &IPRange{ParseIP("10.0.0.24"), ParseIP("10.0.0.24")}, true},
		{&IPRange{ParseIP("10.0.0.0"), ParseIP("10.0.0.255")}, &IPRange{ParseIP("10.0.0.24"), ParseIP("10.0.0.255")}, true},
		{&IPRange{ParseIP("10.0.0.0"), ParseIP("10.0.0.255")}, &IPRange{ParseIP("10.0.0.24"), ParseIP("11.0.0.0")}, false},
		{&IPRange{ParseIP("10.0.0.0"), ParseIP("10.0.0.255")}, &IPRange{ParseIP("10.0.0.255"), ParseIP("10.0.0.255")}, true},
		{&IPRange{ParseIP("10.0.0.0"), ParseIP("10.0.0.255")}, &IPRange{ParseIP("10.0.0.255"), ParseIP("11.0.0.0")}, false},
		{&IPRange{ParseIP("10.0.0.0"), ParseIP("10.0.0.255")}, &IPRange{ParseIP("11.0.0.0"), ParseIP("11.0.0.0")}, false},

		{&IPRange{ParseIP("10.0.0.0"), ParseIP("10.0.0.0")}, &IPRange{ParseIP("10.0.0.0"), ParseIP("10.0.0.0")}, true},
	} {
		ret := tc.a.Contains(tc.b)
		if !assert.Equal(t, tc.result, ret) {
			t.Logf("Test %d failed: a: %s b: %s was %v and was supposed to be %v", i, tc.a, tc.b, ret, tc.result)
		}
	}
}
