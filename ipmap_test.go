package netaddr

import (
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInsertWrongFamily(t *testing.T) {
	m := NewIPMap(net.IPv6len)
	err := m.Insert(ParseIP("10.224.24.0"), nil)
	assert.NotNil(t, err)
	assert.Equal(t, 0, m.Size())
}

func TestInsert(t *testing.T) {
	m := NewIPMap(net.IPv6len)
	err := m.Insert(ParseIP("2000::1"), 3)
	assert.Nil(t, err)
	assert.Equal(t, 1, m.Size())

	data, ok := m.Get(ParseIP("2000::1"))
	assert.True(t, ok)
	assert.Equal(t, 3, data)

	data, ok = m.Get(ParseIP("3000::1"))
	assert.False(t, ok)
}

func TestInsertDuplicate(t *testing.T) {
	m := NewIPMap(net.IPv6len)
	err := m.Insert(ParseIP("2000::1"), 3)
	assert.Nil(t, err)
	assert.Equal(t, 1, m.Size())

	err = m.Insert(ParseIP("2000::1"), 3)
	assert.NotNil(t, err)
	assert.Equal(t, 1, m.Size())
}

func TestGetWrongFamily(t *testing.T) {
	m := NewIPMap(net.IPv6len)
	err := m.InsertNet(unsafeParseNet("ae0:1801::/32"), 3)
	assert.Nil(t, err)
	assert.Equal(t, 1, m.Size())

	_, ok := m.Get(IPv4(0xa, 0xe0, 0x18, 0x1))
	assert.False(t, ok)
}

func TestGetOnlyExactMatch(t *testing.T) {
	m := NewIPMap(net.IPv6len)
	m.InsertNet(unsafeParseNet("2000::/64"), 3)
	assert.Equal(t, 1, m.Size())

	_, ok := m.Get(ParseIP("2000::1"))
	assert.False(t, ok)
}

func TestGetNotFound(t *testing.T) {
	m := NewIPMap(net.IPv6len)
	err := m.Insert(ParseIP("2000::1"), 3)
	assert.Nil(t, err)
	assert.Equal(t, 1, m.Size())

	_, ok := m.Get(ParseIP("3000::1"))
	assert.False(t, ok)
}

func TestMatchWrongFamily(t *testing.T) {
	m := NewIPMap(net.IPv6len)
	err := m.InsertNet(unsafeParseNet("ae0:1801::/32"), 3)
	assert.Nil(t, err)
	assert.Equal(t, 1, m.Size())

	_, ok := m.Match(IPv4(0xa, 0xe0, 0x18, 0x1))
	assert.False(t, ok)
}

func TestMatchLongestPrefixMatch(t *testing.T) {
	m := NewIPMap(net.IPv6len)
	m.InsertNet(unsafeParseNet("2000::/64"), 3)
	assert.Equal(t, 1, m.Size())
	m.InsertNet(unsafeParseNet("2000::/32"), 4)
	assert.Equal(t, 2, m.Size())

	data, ok := m.Match(ParseIP("2000::1"))
	assert.Equal(t, 3, data)
	assert.True(t, ok)
}

func TestMatchNotFound(t *testing.T) {
	m := NewIPMap(net.IPv6len)
	err := m.Insert(ParseIP("2000::1"), 3)
	assert.Nil(t, err)
	assert.Equal(t, 1, m.Size())

	_, ok := m.Match(ParseIP("3000::1"))
	assert.False(t, ok)
}

func TestRemove(t *testing.T) {
	m := NewIPMap(net.IPv6len)
	err := m.Insert(ParseIP("2000::1"), 3)
	assert.Nil(t, err)
	assert.Equal(t, 1, m.Size())

	m.Remove(ParseIP("2000::1"))
	assert.Equal(t, 0, m.Size())
}

func TestRemoveWrongFamily(t *testing.T) {
	m := NewIPMap(net.IPv6len)
	err := m.InsertNet(unsafeParseNet("ae0:1801::/32"), 3)
	assert.Nil(t, err)
	assert.Equal(t, 1, m.Size())

	m.Remove(IPv4(0xa, 0xe0, 0x18, 0x1))
	assert.Equal(t, 1, m.Size())
}

func TestRemoveNotFound(t *testing.T) {
	m := NewIPMap(net.IPv6len)
	err := m.Insert(ParseIP("2000::1"), 3)
	assert.Nil(t, err)
	assert.Equal(t, 1, m.Size())

	m.Remove(ParseIP("3000::1"))
	assert.Equal(t, 1, m.Size())
}

// unsafeParseNet is for testing only. It ignores the error to inline it
func unsafeParseNet(str string) *net.IPNet {
	n, _ := ParseNet(str)
	return n
}

func TestInsertNetWrongFamily(t *testing.T) {
	m := NewIPMap(net.IPv6len)
	err := m.InsertNet(unsafeParseNet("10.224.24.0/24"), nil)
	assert.NotNil(t, err)
	assert.Equal(t, 0, m.Size())
}

func TestInsertNet(t *testing.T) {
	m := NewIPMap(net.IPv6len)
	err := m.InsertNet(unsafeParseNet("2000::/64"), 3)
	assert.Nil(t, err)
	assert.Equal(t, 1, m.Size())

	data, ok := m.GetNet(unsafeParseNet("2000::/64"))
	assert.True(t, ok)
	assert.Equal(t, 3, data)

	data, ok = m.GetNet(unsafeParseNet("3000::/64"))
	assert.False(t, ok)
}

func TestRemoveNet(t *testing.T) {
	m := NewIPMap(net.IPv6len)
	err := m.InsertNet(unsafeParseNet("2000::/64"), 3)
	assert.Nil(t, err)
	assert.Equal(t, 1, m.Size())

	m.RemoveNet(unsafeParseNet("2000::/64"))
	assert.Equal(t, 0, m.Size())
}

func TestRemoveNetWrongFamily(t *testing.T) {
	m := NewIPMap(net.IPv6len)
	err := m.InsertNet(unsafeParseNet("ae0:1801::/32"), 3)
	assert.Nil(t, err)
	assert.Equal(t, 1, m.Size())

	// The 32 bits of this IPv4 address match the first 32 bits of the IPv6 address
	m.RemoveNet(unsafeParseNet("10.224.24.1/32"))
	assert.Equal(t, 1, m.Size())
}

func TestRemoveNetNotFound(t *testing.T) {
	m := NewIPMap(net.IPv6len)
	err := m.InsertNet(unsafeParseNet("2000::/64"), 3)
	assert.Nil(t, err)
	assert.Equal(t, 1, m.Size())

	m.RemoveNet(unsafeParseNet("3000::/64"))
	assert.Equal(t, 1, m.Size())
}

func TestMatchNetWrongFamily(t *testing.T) {
	m := NewIPMap(net.IPv6len)
	m.InsertNet(unsafeParseNet("ae0:1801::/32"), 3)
	assert.Equal(t, 1, m.Size())

	_, ok := m.MatchNet(unsafeParseNet("10.224.24.0/32"))
	assert.False(t, ok)
}

func TestMatchNetLongestPrefixMatch(t *testing.T) {
	m := NewIPMap(net.IPv6len)
	m.InsertNet(unsafeParseNet("2000::/64"), 3)
	assert.Equal(t, 1, m.Size())
	m.InsertNet(unsafeParseNet("2000::/32"), 4)
	assert.Equal(t, 2, m.Size())

	data, ok := m.MatchNet(unsafeParseNet("2000::/124"))
	assert.Equal(t, 3, data)
	assert.True(t, ok)
}

func TestMatchNetNotFound(t *testing.T) {
	m := NewIPMap(net.IPv6len)
	m.InsertNet(unsafeParseNet("2000::/64"), 3)
	assert.Equal(t, 1, m.Size())

	_, ok := m.MatchNet(unsafeParseNet("3000::/64"))
	assert.False(t, ok)
}

func TestExample1(t *testing.T) {
	m := NewIPMap(net.IPv4len)
	m.InsertNet(unsafeParseNet("10.224.24.2/31"), true)
	m.InsertNet(unsafeParseNet("10.224.24.1/32"), true)
	m.InsertNet(unsafeParseNet("10.224.24.0/32"), true)

	var result []string
	m.Iterate(func(net *net.IPNet, value interface{}) {
		result = append(result, net.String())
	})
	assert.Equal(
		t,
		[]string{
			"10.224.24.0/32",
			"10.224.24.1/32",
			"10.224.24.2/31",
		},
		result,
	)

	result = []string{}
	m.Aggregate(func(net *net.IPNet, value interface{}) {
		result = append(result, net.String())
	})
	assert.Equal(
		t,
		[]string{
			"10.224.24.0/30",
		},
		result,
	)
}

type pair struct {
	net   string
	value interface{}
}

func TestExample2(t *testing.T) {
	m := NewIPMap(net.IPv4len)
	m.InsertNet(unsafeParseNet("10.224.24.0/30"), true)
	m.InsertNet(unsafeParseNet("10.224.24.0/31"), false)
	m.InsertNet(unsafeParseNet("10.224.24.1/32"), true)
	m.InsertNet(unsafeParseNet("10.224.24.0/32"), false)

	var result []pair
	m.Iterate(func(net *net.IPNet, value interface{}) {
		result = append(
			result,
			pair{
				net:   net.String(),
				value: value,
			},
		)
	})
	assert.Equal(
		t,
		[]pair{
			pair{net: "10.224.24.0/30", value: true},
			pair{net: "10.224.24.0/31", value: false},
			pair{net: "10.224.24.0/32", value: false},
			pair{net: "10.224.24.1/32", value: true},
		},
		result,
	)

	result = []pair{}
	m.Aggregate(func(net *net.IPNet, value interface{}) {
		result = append(
			result,
			pair{
				net:   net.String(),
				value: value,
			},
		)
	})
	assert.Equal(
		t,
		[]pair{
			pair{net: "10.224.24.0/30", value: true},
			pair{net: "10.224.24.0/31", value: false},
			pair{net: "10.224.24.1/32", value: true},
		},
		result,
	)
}

func TestExample3(t *testing.T) {
	m := NewIPMap(net.IPv4len)
	m.InsertNet(unsafeParseNet("172.21.0.0/20"), nil)
	m.InsertNet(unsafeParseNet("192.68.27.0/25"), nil)
	m.InsertNet(unsafeParseNet("192.168.26.128/25"), nil)
	m.InsertNet(unsafeParseNet("10.224.24.0/32"), nil)
	m.InsertNet(unsafeParseNet("192.68.24.0/24"), nil)
	m.InsertNet(unsafeParseNet("172.16.0.0/12"), nil)
	m.InsertNet(unsafeParseNet("192.68.26.0/24"), nil)
	m.InsertNet(unsafeParseNet("10.224.24.0/30"), nil)
	m.InsertNet(unsafeParseNet("192.168.24.0/24"), nil)
	m.InsertNet(unsafeParseNet("192.168.25.0/24"), nil)
	m.InsertNet(unsafeParseNet("192.168.26.0/25"), nil)
	m.InsertNet(unsafeParseNet("192.68.25.0/24"), nil)
	m.InsertNet(unsafeParseNet("192.168.27.0/24"), nil)
	m.InsertNet(unsafeParseNet("172.20.128.0/19"), nil)
	m.InsertNet(unsafeParseNet("192.68.27.128/25"), nil)

	var result []string
	m.Iterate(func(net *net.IPNet, value interface{}) {
		result = append(result, net.String())
	})
	assert.Equal(
		t,
		[]string{
			"10.224.24.0/30",
			"10.224.24.0/32",
			"172.16.0.0/12",
			"172.20.128.0/19",
			"172.21.0.0/20",
			"192.68.24.0/24",
			"192.68.25.0/24",
			"192.68.26.0/24",
			"192.68.27.0/25",
			"192.68.27.128/25",
			"192.168.24.0/24",
			"192.168.25.0/24",
			"192.168.26.0/25",
			"192.168.26.128/25",
			"192.168.27.0/24",
		},
		result,
	)

	result = []string{}
	m.Aggregate(func(net *net.IPNet, value interface{}) {
		result = append(result, net.String())
	})
	assert.Equal(
		t,
		[]string{
			"10.224.24.0/30",
			"172.16.0.0/12",
			"192.68.24.0/22",
			"192.168.24.0/22",
		},
		result,
	)
}
