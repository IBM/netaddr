package netaddr

import (
	"math/big"
	"math/rand"
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
)

var (
	Eights = net.ParseIP("8.8.8.8").To4()
	Nines  = net.ParseIP("9.9.9.9").To4()

	_, Ten24, _    = net.ParseCIDR("10.0.0.0/24")
	_, Ten24128, _ = net.ParseCIDR("10.0.0.128/25")
	Ten24Router    = net.ParseIP("10.0.0.1").To4()
	Ten24Broadcast = net.ParseIP("10.0.0.255").To4()

	_, V6Net1, _ = net.ParseCIDR("2001:db8:1234:abcd::/64")
	_, V6Net2, _ = net.ParseCIDR("2001:db8:abcd:1234::/64")
	V6Net1Router = net.ParseIP("2001:db8:1234:abcd::1")

	V6NetSize = big.NewInt(0).Lsh(big.NewInt(1), 64) // 2**64 or 18446744073709551616
)

func TestNetDifference(t *testing.T) {
	diff := netDifference(Ten24, Ten24128)

	_, cidr, _ := net.ParseCIDR("10.0.0.0/25")
	assert.Equal(t, []*net.IPNet{cidr}, diff)

	_, cidr, _ = net.ParseCIDR("10.0.0.120/29")
	diff = netDifference(Ten24, cidr)

	_, cidr1, _ := net.ParseCIDR("10.0.0.128/25")
	_, cidr2, _ := net.ParseCIDR("10.0.0.0/26")
	_, cidr3, _ := net.ParseCIDR("10.0.0.64/27")
	_, cidr4, _ := net.ParseCIDR("10.0.0.96/28")
	_, cidr5, _ := net.ParseCIDR("10.0.0.112/29")
	assert.Equal(t, []*net.IPNet{cidr1, cidr2, cidr3, cidr4, cidr5}, diff)
}

func TestIPSetInit(t *testing.T) {
	set := IPSet{}

	assert.Equal(t, big.NewInt(0), set.tree.size())
}

func TestIPSetContains(t *testing.T) {
	set := IPSet{}

	assert.Equal(t, big.NewInt(0), set.tree.size())
	assert.False(t, set.Contains(Eights))
	assert.False(t, set.Contains(Nines))
}

func TestIPSetInsert(t *testing.T) {
	set := IPSet{}

	set.Insert(Nines)
	assert.Equal(t, 1, set.tree.numNodes())
	assert.Equal(t, big.NewInt(1), set.tree.size())
	assert.True(t, set.Contains(Nines))
	assert.False(t, set.Contains(Eights))
	set.Insert(Eights)
	assert.Equal(t, 2, set.tree.numNodes())
	assert.True(t, set.Contains(Eights))
}

func TestIPSetInsertNetwork(t *testing.T) {
	set := IPSet{}

	set.InsertNet(Ten24)
	assert.Equal(t, 1, set.tree.numNodes())
	assert.Equal(t, big.NewInt(256), set.tree.size())
	assert.True(t, set.ContainsNet(Ten24))
	assert.True(t, set.ContainsNet(Ten24128))
	assert.False(t, set.Contains(Nines))
	assert.False(t, set.Contains(Eights))
}

func TestIPSetInsertMixed(t *testing.T) {
	set := IPSet{}

	set.InsertNet(Ten24)
	assert.Equal(t, 1, set.tree.numNodes())
	set.Insert(Eights)
	set.Insert(Nines)
	set.Insert(Ten24Router)
	assert.Equal(t, 3, set.tree.numNodes())
	assert.Equal(t, big.NewInt(258), set.tree.size())
	assert.True(t, set.ContainsNet(Ten24))
	assert.True(t, set.ContainsNet(Ten24128))
	assert.True(t, set.Contains(Ten24Router))
	assert.True(t, set.Contains(Eights))
	assert.True(t, set.Contains(Nines))
}

func TestIPSetInsertSequential(t *testing.T) {
	set := IPSet{}

	set.Insert(net.ParseIP("192.168.1.0").To4())
	assert.Equal(t, 1, set.tree.numNodes())
	set.Insert(net.ParseIP("192.168.1.1").To4())
	assert.Equal(t, 1, set.tree.numNodes())
	set.Insert(net.ParseIP("192.168.1.2").To4())
	assert.Equal(t, 2, set.tree.numNodes())
	set.Insert(net.ParseIP("192.168.1.3").To4())
	assert.Equal(t, 1, set.tree.numNodes())
	assert.Equal(t, big.NewInt(4), set.tree.size())

	_, cidr, _ := net.ParseCIDR("192.168.1.0/30")
	assert.True(t, set.ContainsNet(cidr))

	_, cidr, _ = net.ParseCIDR("192.168.1.4/31")
	set.InsertNet(cidr)
	assert.Equal(t, 2, set.tree.numNodes())
	assert.True(t, set.ContainsNet(cidr))

	_, cidr, _ = net.ParseCIDR("192.168.1.6/31")
	set.InsertNet(cidr)
	assert.Equal(t, 1, set.tree.numNodes())
	assert.True(t, set.ContainsNet(cidr))

	_, cidr, _ = net.ParseCIDR("192.168.1.6/31")
	set.InsertNet(cidr)
	assert.Equal(t, 1, set.tree.numNodes())
	assert.True(t, set.ContainsNet(cidr))

	_, cidr, _ = net.ParseCIDR("192.168.0.240/29")
	set.InsertNet(cidr)
	assert.Equal(t, 2, set.tree.numNodes())
	assert.True(t, set.ContainsNet(cidr))

	_, cidr, _ = net.ParseCIDR("192.168.0.248/29")
	set.InsertNet(cidr)
	assert.Equal(t, 2, set.tree.numNodes())
	assert.True(t, set.ContainsNet(cidr))
}

func TestIPSetRemove(t *testing.T) {
	set := IPSet{}

	set.InsertNet(Ten24)
	assert.Equal(t, 1, set.tree.numNodes())
	set.RemoveNet(Ten24128)
	assert.Equal(t, 1, set.tree.numNodes())
	assert.Equal(t, big.NewInt(128), set.tree.size())
	assert.False(t, set.ContainsNet(Ten24))
	assert.False(t, set.ContainsNet(Ten24128))
	_, cidr, _ := net.ParseCIDR("10.0.0.0/25")
	assert.True(t, set.ContainsNet(cidr))

	set.Remove(Ten24Router)
	assert.Equal(t, big.NewInt(127), set.tree.size())
	assert.Equal(t, 7, set.tree.numNodes())
}

func TestIPSetRemoveNetworkBroadcast(t *testing.T) {
	set := IPSet{}

	set.InsertNet(Ten24)
	assert.Equal(t, 1, set.tree.numNodes())
	set.Remove(Ten24.IP)
	set.Remove(Ten24Broadcast)
	assert.Equal(t, big.NewInt(254), set.tree.size())
	assert.Equal(t, 14, set.tree.numNodes())
	assert.False(t, set.ContainsNet(Ten24))
	assert.False(t, set.ContainsNet(Ten24128))
	assert.False(t, set.Contains(Ten24Broadcast))
	assert.False(t, set.Contains(Ten24.IP))

	_, cidr, _ := net.ParseCIDR("10.0.0.128/26")
	assert.True(t, set.ContainsNet(cidr))
	assert.True(t, set.Contains(Ten24Router))

	set.Remove(Ten24Router)
	assert.Equal(t, big.NewInt(253), set.tree.size())
	assert.Equal(t, 13, set.tree.numNodes())
}

func TestIPSetRemoveAll(t *testing.T) {
	set := IPSet{}

	set.InsertNet(Ten24)
	_, cidr1, _ := net.ParseCIDR("192.168.0.0/25")
	set.InsertNet(cidr1)
	assert.Equal(t, 2, set.tree.numNodes())

	_, cidr2, _ := net.ParseCIDR("0.0.0.0/0")
	set.RemoveNet(cidr2)
	assert.Equal(t, 0, set.tree.numNodes())
	assert.False(t, set.ContainsNet(Ten24))
	assert.False(t, set.ContainsNet(Ten24128))
	assert.False(t, set.ContainsNet(cidr1))
}

func TestIPSetInsertOverlapping(t *testing.T) {
	set := IPSet{}

	set.InsertNet(Ten24128)
	assert.False(t, set.ContainsNet(Ten24))
	assert.Equal(t, 1, set.tree.numNodes())
	set.InsertNet(Ten24)
	assert.Equal(t, 1, set.tree.numNodes())
	assert.Equal(t, big.NewInt(256), set.tree.size())
	assert.True(t, set.ContainsNet(Ten24))
	assert.True(t, set.Contains(Ten24Router))
	assert.False(t, set.Contains(Eights))
	assert.False(t, set.Contains(Nines))
}

func TestIPSetUnion(t *testing.T) {
	set1, set2 := &IPSet{}, &IPSet{}

	set1.InsertNet(Ten24)
	_, cidr, _ := net.ParseCIDR("192.168.0.248/29")
	set2.InsertNet(cidr)

	set := set1.Union(set2)
	assert.True(t, set.ContainsNet(Ten24))
	assert.True(t, set.ContainsNet(cidr))
}

func TestIPSetDifference(t *testing.T) {
	set1, set2 := &IPSet{}, &IPSet{}

	set1.InsertNet(Ten24)
	_, cidr, _ := net.ParseCIDR("192.168.0.248/29")
	set2.InsertNet(cidr)

	set := set1.Difference(set2)
	assert.True(t, set.ContainsNet(Ten24))
	assert.False(t, set.ContainsNet(cidr))
}

func TestIPSetInsertV6(t *testing.T) {
	set := IPSet{}

	set.InsertNet(V6Net1)
	assert.Equal(t, 1, set.tree.numNodes())
	set.Insert(V6Net1Router)
	assert.Equal(t, 1, set.tree.numNodes())
	assert.Equal(t, V6NetSize, set.tree.size())
	assert.True(t, set.ContainsNet(V6Net1))
	assert.False(t, set.ContainsNet(V6Net2))
	assert.False(t, set.Contains(Ten24Router))
	assert.True(t, set.Contains(V6Net1Router))
	assert.False(t, set.Contains(Eights))
	assert.False(t, set.Contains(Nines))

	set.InsertNet(V6Net2)
	assert.Equal(t, 2, set.tree.numNodes())
	assert.True(t, set.ContainsNet(V6Net1))
	assert.True(t, set.ContainsNet(V6Net2))
	assert.Equal(t, big.NewInt(0).Mul(big.NewInt(2), V6NetSize), set.tree.size())
}

func TestIPSetAllocateDeallocate(t *testing.T) {
	rand.Seed(29)

	set := IPSet{}

	_, bigNet, _ := net.ParseCIDR("15.1.64.0/16")
	set.InsertNet(bigNet)

	ips := set.GetIPs(0)
	assert.Equal(t, 65536, len(ips))
	assert.Equal(t, big.NewInt(65536), set.tree.size())

	allocated := &IPSet{}
	for i := 0; i != 16384; i++ {
		allocated.Insert(ips[rand.Intn(65536)])
	}
	assert.Equal(t, big.NewInt(14500), allocated.tree.size())
	ips = allocated.GetIPs(0)
	assert.Equal(t, 14500, len(ips))
	for _, ip := range ips {
		assert.True(t, set.Contains(ip))
	}

	available := set.Difference(allocated)
	assert.Equal(t, big.NewInt(51036), available.tree.size())
	ips = available.GetIPs(0)
	for _, ip := range ips {
		assert.True(t, set.Contains(ip))
		assert.False(t, allocated.Contains(ip))
	}
	assert.Equal(t, 51036, len(ips))
}
