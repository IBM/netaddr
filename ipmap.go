package netaddr

import (
	"fmt"
	"net"

	"github.com/ecbaldwin/trie"
)

type IPMap struct {
	length uint
	trie   *trie.TrieNode
}

// NewIPMap returns a new map
// `length` should be either net.IPv4len or net.IPv6len
func NewIPMap(length uint) *IPMap {
	return &IPMap{
		length: length,
	}
}

// Size returns the number of exact prefixes stored in the map
func (m *IPMap) Size() int {
	return m.trie.Size()
}

func netToKey(net *net.IPNet) *trie.TrieKey {
	ones, _ := net.Mask.Size()
	return &trie.TrieKey{
		Length: uint(ones),
		Bits:   net.IP,
	}
}

// GetNet returns the value in the map associated with the given network prefix
// with an exact match: both the IP and the prefix length must match. If an
// exact match is not found, found is false and value is nil and should be
// ignored.
func (m *IPMap) GetNet(net *net.IPNet) (value interface{}, found bool) {
	if net == nil {
		return nil, false
	}
	if uint(len(net.IP)) != m.length {
		return nil, false
	}

	key := netToKey(net)
	node := m.trie.Match(key)

	if node == nil {
		return nil, false
	}

	// Only return an exact match
	if node.TrieKey.Length != key.Length {
		return nil, false
	}
	return node.Data, true
}

// MatchNet returns the value in the map associated with the given network
// prefix with a longest prefix match. If a match is not found, found is false
// and value is nil and should be ignored.
func (m *IPMap) MatchNet(net *net.IPNet) (value interface{}, found bool) {
	if net == nil {
		return nil, false
	}
	if uint(len(net.IP)) != m.length {
		return nil, false
	}

	key := netToKey(net)
	node := m.trie.Match(key)

	if node == nil {
		return nil, false
	}

	return node.Data, true
}

// InsertNet inserts the given IPNet with the given value into the map
func (m *IPMap) InsertNet(net *net.IPNet, value interface{}) error {
	if net == nil {
		return fmt.Errorf("cannot insert nil IPNet")
	}
	if uint(len(net.IP)) != m.length {
		return fmt.Errorf("cannot insert IPNet with length %d in map with length %d", len(net.IP), m.length)
	}

	newHead, err := m.trie.Insert(&trie.TrieNode{
		TrieKey: *netToKey(net),
		Data:    value,
	})

	if err != nil {
		return err
	}

	m.trie = newHead
	return nil
}

func (m *IPMap) RemoveNet(net *net.IPNet) {
	if net == nil {
		return
	}
	if uint(len(net.IP)) != m.length {
		return
	}

	newHead, err := m.trie.Delete(netToKey(net))
	if err == nil {
		m.trie = newHead
	}
}

func ipToKey(ip net.IP) *trie.TrieKey {
	return &trie.TrieKey{
		Length: uint(8 * len(ip)),
		Bits:   ip,
	}
}

// Get is like GetNet to get an exact match for the given IP address
// interpreted as a host prefix (i.e. /32 for IPv4 and /128 for IPv6)
func (m *IPMap) Get(ip net.IP) (value interface{}, found bool) {
	if uint(len(ip)) != m.length {
		return nil, false
	}

	key := ipToKey(ip)
	node := m.trie.Match(key)

	if node == nil {
		return nil, false
	}

	// Only return an exact match
	if node.TrieKey.Length != key.Length {
		return nil, false
	}
	return node.Data, true
}

// Match is like MatchNet to match the given IP address as a host prefix
func (m *IPMap) Match(ip net.IP) (value interface{}, found bool) {
	if uint(len(ip)) != m.length {
		return nil, false
	}

	key := ipToKey(ip)
	node := m.trie.Match(key)

	if node == nil {
		return nil, false
	}

	return node.Data, true
}

// Insert is like InsertNet for the given ip as a host prefix
func (m *IPMap) Insert(ip net.IP, value interface{}) error {
	if uint(len(ip)) != m.length {
		return fmt.Errorf("cannot insert IP with length %d in map with length %d", len(ip), m.length)
	}

	newHead, err := m.trie.Insert(&trie.TrieNode{
		TrieKey: *ipToKey(ip),
		Data:    value,
	})

	if err != nil {
		return err
	}

	m.trie = newHead
	return nil
}

// Remove is like RemoveNet for the given ip as a host prefix
func (m *IPMap) Remove(ip net.IP) {
	if uint(len(ip)) != m.length {
		return
	}

	newHead, err := m.trie.Delete(ipToKey(ip))
	if err == nil {
		m.trie = newHead
	}
}

type Callback func(net *net.IPNet, value interface{})

func trieCallback(m *IPMap, callback Callback) trie.Callback {
	return func(key *trie.TrieKey, data interface{}) {
		// The trie implementation may not store a full 4 or 16 bytes if the prefix
		// length is shorter. But, we want the full size when creating a net.IP.
		ip := NewIP(int(m.length))
		copy(ip, key.Bits)

		callback(
			&net.IPNet{
				IP: ip,
				Mask: net.CIDRMask(
					int(key.Length),
					8*int(m.length),
				),
			},
			data,
		)
	}
}

// Iterate calls `callback` for each key/value pair in the map in lexigraphical order.
func (m *IPMap) Iterate(callback Callback) {
	m.trie.Iterate(trieCallback(m, callback))
}

// Aggregate calls `callback` for each key/value pair in the map aggregated by
// value in lexigraphical order.
//
// 1. The values stored must be comparable. Prefixes get aggregated only where
//    their values compare equal.
// 2. The set of key/value pairs visited is the minimal-size set such that any
//    longest prefix match against the aggregated set will always return the same
//    value as the same match against the non-aggregated set.
// 3. The aggregated and non-aggregated sets of keys may be disjoint.
func (m *IPMap) Aggregate(callback Callback) {
	m.trie.Aggregate(trieCallback(m, callback))
}
