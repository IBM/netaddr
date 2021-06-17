package netaddr

import (
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInsertOrUpdateWrongFamily(t *testing.T) {
	m := NewIPv6Map()
	m.Insert(ParseIP("2000::1"), nil)
	err := m.InsertOrUpdate(ParseIP("10.224.24.0"), true)
	assert.NotNil(t, err)
	assert.Equal(t, 1, m.Size())
}

func TestInsertOrUpdate(t *testing.T) {
	m := NewIPv6Map()
	m.Insert(ParseIP("2000::1"), nil)
	err := m.InsertOrUpdate(ParseIP("2000::1"), 3)
	assert.Nil(t, err)
	assert.Equal(t, 1, m.Size())

	data, ok := m.Get(ParseIP("2000::1"))
	assert.True(t, ok)
	assert.Equal(t, 3, data)
}

func TestInsertOrUpdateDuplicate(t *testing.T) {
	m := NewIPv6Map()
	err := m.InsertOrUpdate(ParseIP("2000::1"), 3)
	assert.Nil(t, err)
	assert.Equal(t, 1, m.Size())
	data, ok := m.Get(ParseIP("2000::1"))
	assert.True(t, ok)
	assert.Equal(t, 3, data)

	err = m.InsertOrUpdate(ParseIP("2000::1"), 4)
	assert.Nil(t, err)
	assert.Equal(t, 1, m.Size())
	data, ok = m.Get(ParseIP("2000::1"))
	assert.True(t, ok)
	assert.Equal(t, 4, data)
}

func TestGetWrongFamily(t *testing.T) {
	m := NewIPv6Map()
	err := m.InsertPrefix(unsafeParseNet("ae0:1801::/32"), 3)
	assert.Nil(t, err)
	assert.Equal(t, 1, m.Size())

	_, ok := m.Get(IPv4(0xa, 0xe0, 0x18, 0x1))
	assert.False(t, ok)
}

func TestGetOnlyExactMatch(t *testing.T) {
	m := NewIPv6Map()
	m.InsertPrefix(unsafeParseNet("2000::/64"), 3)
	assert.Equal(t, 1, m.Size())

	_, ok := m.Get(ParseIP("2000::1"))
	assert.False(t, ok)
}

func TestGetNotFound(t *testing.T) {
	m := NewIPv6Map()
	err := m.Insert(ParseIP("2000::1"), 3)
	assert.Nil(t, err)
	assert.Equal(t, 1, m.Size())

	_, ok := m.Get(ParseIP("3000::1"))
	assert.False(t, ok)
}

func TestGetOrInsertWrongFamily(t *testing.T) {
	m := NewIPv6Map()
	err := m.InsertPrefix(unsafeParseNet("ae0:1801::/32"), 3)
	assert.Nil(t, err)
	assert.Equal(t, 1, m.Size())

	value, err := m.GetOrInsert(IPv4(0xa, 0xe0, 0x18, 0x1), 5)
	assert.NotNil(t, err)
	assert.Nil(t, value)
	assert.Equal(t, 1, m.Size())
}

func TestGetOrInsertOnlyExactMatch(t *testing.T) {
	m := NewIPv6Map()
	m.InsertPrefix(unsafeParseNet("2000::/64"), 3)
	assert.Equal(t, 1, m.Size())

	value, err := m.GetOrInsert(ParseIP("2000::1"), 5)
	assert.Nil(t, err)
	assert.Equal(t, 5, value)
	assert.Equal(t, 2, m.Size())
}

func TestGetOrInsertNotFound(t *testing.T) {
	m := NewIPv6Map()
	err := m.Insert(ParseIP("2000::1"), 3)
	assert.Nil(t, err)

	value, err := m.GetOrInsert(ParseIP("3000::1"), 5)
	assert.Nil(t, err)
	assert.Equal(t, 5, value)
	assert.Equal(t, 2, m.Size())
}

func TestGetOrInsertPrefixWrongFamily(t *testing.T) {
	m := NewIPv6Map()
	err := m.InsertPrefix(unsafeParseNet("ae0:1801::/32"), 3)
	assert.Nil(t, err)
	assert.Equal(t, 1, m.Size())

	value, err := m.GetOrInsertPrefix(unsafeParseNet("10.0.0.0/24"), 5)
	assert.NotNil(t, err)
	assert.Nil(t, value)
	assert.Equal(t, 1, m.Size())
}

func TestGetOrInsertPrefixOnlyExactMatch(t *testing.T) {
	m := NewIPv6Map()
	m.InsertPrefix(unsafeParseNet("2000::/64"), 3)
	assert.Equal(t, 1, m.Size())

	value, err := m.GetOrInsertPrefix(unsafeParseNet("2000::2/127"), 5)
	assert.Nil(t, err)
	assert.Equal(t, 5, value)
	assert.Equal(t, 2, m.Size())
}

func TestGetOrInsertPrefixNotFound(t *testing.T) {
	m := NewIPv6Map()
	err := m.Insert(ParseIP("2000::1"), 3)
	assert.Nil(t, err)

	value, err := m.GetOrInsertPrefix(unsafeParseNet("3000::2/127"), 5)
	assert.Nil(t, err)
	assert.Equal(t, 5, value)
	assert.Equal(t, 2, m.Size())
}

func TestMatchWrongFamily(t *testing.T) {
	m := NewIPv6Map()
	err := m.InsertPrefix(unsafeParseNet("ae0:1801::/32"), 3)
	assert.Nil(t, err)
	assert.Equal(t, 1, m.Size())

	n, _ := m.Match(IPv4(0xa, 0xe0, 0x18, 0x1))
	assert.Nil(t, n)
}

func TestMatchLongestPrefixMatch(t *testing.T) {
	m := NewIPv6Map()
	m.InsertPrefix(unsafeParseNet("2000::/64"), 3)
	assert.Equal(t, 1, m.Size())
	m.InsertPrefix(unsafeParseNet("2000::/32"), 4)
	assert.Equal(t, 2, m.Size())

	n, data := m.Match(ParseIP("2000::1"))
	assert.Equal(t, unsafeParseNet("2000::/64"), n)
	assert.Equal(t, 3, data)
}

func TestMatchNotFound(t *testing.T) {
	m := NewIPv6Map()
	err := m.Insert(ParseIP("2000::1"), 3)
	assert.Nil(t, err)
	assert.Equal(t, 1, m.Size())

	n, _ := m.Match(ParseIP("3000::1"))
	assert.Nil(t, n)
}

func TestRemove(t *testing.T) {
	m := NewIPv6Map()
	err := m.Insert(ParseIP("2000::1"), 3)
	assert.Nil(t, err)
	assert.Equal(t, 1, m.Size())

	m.Remove(ParseIP("2000::1"))
	assert.Equal(t, 0, m.Size())
}

func TestRemoveWrongFamily(t *testing.T) {
	m := NewIPv6Map()
	err := m.InsertPrefix(unsafeParseNet("ae0:1801::/32"), 3)
	assert.Nil(t, err)
	assert.Equal(t, 1, m.Size())

	m.Remove(IPv4(0xa, 0xe0, 0x18, 0x1))
	assert.Equal(t, 1, m.Size())
}

func TestRemoveNotFound(t *testing.T) {
	m := NewIPv6Map()
	err := m.Insert(ParseIP("2000::1"), 3)
	assert.Nil(t, err)
	assert.Equal(t, 1, m.Size())

	m.Remove(ParseIP("3000::1"))
	assert.Equal(t, 1, m.Size())
}

// unsafeParseNet is for testing only. It ignores error so that it can be
// easily inlined.
func unsafeParseNet(str string) *net.IPNet {
	n, _ := ParseNet(str)
	return n
}

func TestInsertPrefixWrongFamily(t *testing.T) {
	m := NewIPv6Map()
	err := m.InsertPrefix(unsafeParseNet("10.224.24.0/24"), nil)
	assert.NotNil(t, err)
	assert.Equal(t, 0, m.Size())
}

func TestInsertPrefix(t *testing.T) {
	m := NewIPv6Map()
	err := m.InsertPrefix(unsafeParseNet("2000::/64"), 3)
	assert.Nil(t, err)
	assert.Equal(t, 1, m.Size())

	data, ok := m.GetPrefix(unsafeParseNet("2000::/64"))
	assert.True(t, ok)
	assert.Equal(t, 3, data)

	data, ok = m.GetPrefix(unsafeParseNet("3000::/64"))
	assert.False(t, ok)
}

func TestInsertOrUpdatePrefixWrongFamily(t *testing.T) {
	m := NewIPv6Map()
	m.InsertPrefix(unsafeParseNet("10.224.24.0/24"), nil)
	err := m.InsertOrUpdatePrefix(unsafeParseNet("10.224.24.0/24"), true)
	assert.NotNil(t, err)
	assert.Equal(t, 0, m.Size())
}

func TestInsertOrUpdatePrefix(t *testing.T) {
	m := NewIPv6Map()
	m.InsertPrefix(unsafeParseNet("2000::/64"), nil)
	err := m.InsertOrUpdatePrefix(unsafeParseNet("2000::/64"), 3)
	assert.Nil(t, err)
	assert.Equal(t, 1, m.Size())

	data, ok := m.GetPrefix(unsafeParseNet("2000::/64"))
	assert.True(t, ok)
	assert.Equal(t, 3, data)

	data, ok = m.GetPrefix(unsafeParseNet("3000::/64"))
	assert.False(t, ok)
}

func TestRemovePrefix(t *testing.T) {
	m := NewIPv6Map()
	err := m.InsertPrefix(unsafeParseNet("2000::/64"), 3)
	assert.Nil(t, err)
	assert.Equal(t, 1, m.Size())

	m.RemovePrefix(unsafeParseNet("2000::/64"))
	assert.Equal(t, 0, m.Size())
}

func TestRemovePrefixWrongFamily(t *testing.T) {
	m := NewIPv6Map()
	err := m.InsertPrefix(unsafeParseNet("ae0:1801::/32"), 3)
	assert.Nil(t, err)
	assert.Equal(t, 1, m.Size())

	// The 32 bits of this IPv4 address match the first 32 bits of the IPv6 address
	m.RemovePrefix(unsafeParseNet("10.224.24.1/32"))
	assert.Equal(t, 1, m.Size())
}

func TestRemovePrefixNotFound(t *testing.T) {
	m := NewIPv6Map()
	err := m.InsertPrefix(unsafeParseNet("2000::/64"), 3)
	assert.Nil(t, err)
	assert.Equal(t, 1, m.Size())

	m.RemovePrefix(unsafeParseNet("3000::/64"))
	assert.Equal(t, 1, m.Size())
}

func TestMatchPrefixWrongFamily(t *testing.T) {
	m := NewIPv6Map()
	m.InsertPrefix(unsafeParseNet("ae0:1801::/32"), 3)
	assert.Equal(t, 1, m.Size())

	n, _ := m.MatchPrefix(unsafeParseNet("10.224.24.0/32"))
	assert.Nil(t, n)
}

func TestMatchPrefixLongestPrefixMatch(t *testing.T) {
	m := NewIPv6Map()
	m.InsertPrefix(unsafeParseNet("2000::/64"), 3)
	assert.Equal(t, 1, m.Size())
	m.InsertPrefix(unsafeParseNet("2000::/32"), 4)
	assert.Equal(t, 2, m.Size())

	n, data := m.MatchPrefix(unsafeParseNet("2000::/124"))
	assert.Equal(t, 3, data)
	assert.Equal(t, unsafeParseNet("2000::/64"), n)
}

func TestMatchPrefixNotFound(t *testing.T) {
	m := NewIPv6Map()
	m.InsertPrefix(unsafeParseNet("2000::/64"), 3)
	assert.Equal(t, 1, m.Size())

	n, _ := m.MatchPrefix(unsafeParseNet("3000::/64"))
	assert.Nil(t, n)
}

func TestExample1(t *testing.T) {
	m := NewIPv4Map()
	m.InsertPrefix(unsafeParseNet("10.224.24.2/31"), true)
	m.InsertPrefix(unsafeParseNet("10.224.24.1/32"), true)
	m.InsertPrefix(unsafeParseNet("10.224.24.0/32"), true)

	var result []string
	m.Iterate(func(net *net.IPNet, value interface{}) bool {
		result = append(result, net.String())
		return true
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
	m.Aggregate(func(net *net.IPNet, value interface{}) bool {
		result = append(result, net.String())
		return true
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
	m := NewIPv4Map()
	m.InsertPrefix(unsafeParseNet("10.224.24.0/30"), true)
	m.InsertPrefix(unsafeParseNet("10.224.24.0/31"), false)
	m.InsertPrefix(unsafeParseNet("10.224.24.1/32"), true)
	m.InsertPrefix(unsafeParseNet("10.224.24.0/32"), false)

	var result []pair
	m.Iterate(func(net *net.IPNet, value interface{}) bool {
		result = append(
			result,
			pair{
				net:   net.String(),
				value: value,
			},
		)
		return true
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
	m.Aggregate(func(net *net.IPNet, value interface{}) bool {
		result = append(
			result,
			pair{
				net:   net.String(),
				value: value,
			},
		)
		return true
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
	m := NewIPv4Map()
	m.InsertPrefix(unsafeParseNet("172.21.0.0/20"), nil)
	m.InsertPrefix(unsafeParseNet("192.68.27.0/25"), nil)
	m.InsertPrefix(unsafeParseNet("192.168.26.128/25"), nil)
	m.InsertPrefix(unsafeParseNet("10.224.24.0/32"), nil)
	m.InsertPrefix(unsafeParseNet("192.68.24.0/24"), nil)
	m.InsertPrefix(unsafeParseNet("172.16.0.0/12"), nil)
	m.InsertPrefix(unsafeParseNet("192.68.26.0/24"), nil)
	m.InsertPrefix(unsafeParseNet("10.224.24.0/30"), nil)
	m.InsertPrefix(unsafeParseNet("192.168.24.0/24"), nil)
	m.InsertPrefix(unsafeParseNet("192.168.25.0/24"), nil)
	m.InsertPrefix(unsafeParseNet("192.168.26.0/25"), nil)
	m.InsertPrefix(unsafeParseNet("192.68.25.0/24"), nil)
	m.InsertPrefix(unsafeParseNet("192.168.27.0/24"), nil)
	m.InsertPrefix(unsafeParseNet("172.20.128.0/19"), nil)
	m.InsertPrefix(unsafeParseNet("192.68.27.128/25"), nil)

	var result []string
	m.Iterate(func(net *net.IPNet, value interface{}) bool {
		result = append(result, net.String())
		return true
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
	iterations := 0
	m.Iterate(func(net *net.IPNet, value interface{}) bool {
		iterations++
		return false
	})
	assert.Equal(t, 1, iterations)

	result = []string{}
	m.Aggregate(func(net *net.IPNet, value interface{}) bool {
		result = append(result, net.String())
		return true
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
	iterations = 0
	m.Aggregate(func(net *net.IPNet, value interface{}) bool {
		iterations++
		return false
	})
	assert.Equal(t, 1, iterations)
}
