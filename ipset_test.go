package netaddr

import (
	"fmt"
	"math/big"
	"math/rand"
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
)

var (
	Eights = net.ParseIP("8.8.8.8").To4()
	Nines  = net.ParseIP("9.9.9.9").To4()

	Ten24, _       = ParseNet("10.0.0.0/24")
	TenOne24, _    = ParseNet("10.0.1.0/24")
	TenTwo24, _    = ParseNet("10.0.2.0/24")
	Ten24128, _    = ParseNet("10.0.0.128/25")
	Ten24Router    = net.ParseIP("10.0.0.1").To4()
	Ten24Broadcast = net.ParseIP("10.0.0.255").To4()

	V6Net1, _    = ParseNet("2001:db8:1234:abcd::/64")
	V6Net2, _    = ParseNet("2001:db8:abcd:1234::/64")
	V6Net1Router = net.ParseIP("2001:db8:1234:abcd::1")

	V6NetSize = big.NewInt(0).Lsh(big.NewInt(1), 64) // 2**64 or 18446744073709551616
)

func TestNetDifference(t *testing.T) {
	diff := netDifference(Ten24, Ten24128)

	cidr, _ := ParseNet("10.0.0.0/25")
	assert.Equal(t, []*net.IPNet{cidr}, diff)

	cidr, _ = ParseNet("10.0.0.120/29")
	diff = netDifference(Ten24, cidr)

	cidr1, _ := ParseNet("10.0.0.128/25")
	cidr2, _ := ParseNet("10.0.0.0/26")
	cidr3, _ := ParseNet("10.0.0.64/27")
	cidr4, _ := ParseNet("10.0.0.96/28")
	cidr5, _ := ParseNet("10.0.0.112/29")
	assert.Equal(t, []*net.IPNet{cidr1, cidr2, cidr3, cidr4, cidr5}, diff)
}

func TestIPSetInit(t *testing.T) {
	set := IPSet{}

	assert.Equal(t, big.NewInt(0), set.tree.size())
	assert.Equal(t, []error{}, set.tree.validate())
}

func TestIPSetContains(t *testing.T) {
	set := IPSet{}

	assert.Equal(t, big.NewInt(0), set.tree.size())
	assert.False(t, set.Contains(Eights))
	assert.False(t, set.Contains(Nines))
	assert.Equal(t, []error{}, set.tree.validate())
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
	assert.Equal(t, []error{}, set.tree.validate())
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
	assert.Equal(t, []error{}, set.tree.validate())
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
	assert.Equal(t, []error{}, set.tree.validate())
}

func TestIPSetInsertSequential(t *testing.T) {
	set := IPSet{}

	set.Insert(net.ParseIP("192.168.1.0").To4())
	assert.Equal(t, 1, set.tree.numNodes())
	assert.Equal(t, []error{}, set.tree.validate())
	set.Insert(net.ParseIP("192.168.1.1").To4())
	assert.Equal(t, 1, set.tree.numNodes())
	assert.Equal(t, []error{}, set.tree.validate())
	set.Insert(net.ParseIP("192.168.1.2").To4())
	assert.Equal(t, 2, set.tree.numNodes())
	assert.Equal(t, []error{}, set.tree.validate())
	set.Insert(net.ParseIP("192.168.1.3").To4())
	assert.Equal(t, 1, set.tree.numNodes())
	assert.Equal(t, []error{}, set.tree.validate())
	assert.Equal(t, big.NewInt(4), set.tree.size())

	cidr, _ := ParseNet("192.168.1.0/30")
	assert.True(t, set.ContainsNet(cidr))

	cidr, _ = ParseNet("192.168.1.4/31")
	set.InsertNet(cidr)
	assert.Equal(t, 2, set.tree.numNodes())
	assert.True(t, set.ContainsNet(cidr))

	cidr, _ = ParseNet("192.168.1.6/31")
	set.InsertNet(cidr)
	assert.Equal(t, 1, set.tree.numNodes())
	assert.True(t, set.ContainsNet(cidr))

	cidr, _ = ParseNet("192.168.1.6/31")
	set.InsertNet(cidr)
	assert.Equal(t, 1, set.tree.numNodes())
	assert.True(t, set.ContainsNet(cidr))

	cidr, _ = ParseNet("192.168.0.240/29")
	set.InsertNet(cidr)
	assert.Equal(t, 2, set.tree.numNodes())
	assert.True(t, set.ContainsNet(cidr))
	assert.Equal(t, []error{}, set.tree.validate())

	cidr, _ = ParseNet("192.168.0.248/29")
	set.InsertNet(cidr)
	assert.Equal(t, 2, set.tree.numNodes())
	assert.True(t, set.ContainsNet(cidr))
	assert.Equal(t, []error{}, set.tree.validate())
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
	cidr, _ := ParseNet("10.0.0.0/25")
	assert.True(t, set.ContainsNet(cidr))
	assert.Equal(t, []error{}, set.tree.validate())

	set.Remove(Ten24Router)
	assert.Equal(t, big.NewInt(127), set.tree.size())
	assert.Equal(t, 7, set.tree.numNodes())
	assert.Equal(t, []error{}, set.tree.validate())
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
	assert.Equal(t, []error{}, set.tree.validate())

	cidr, _ := ParseNet("10.0.0.128/26")
	assert.True(t, set.ContainsNet(cidr))
	assert.True(t, set.Contains(Ten24Router))

	set.Remove(Ten24Router)
	assert.Equal(t, big.NewInt(253), set.tree.size())
	assert.Equal(t, 13, set.tree.numNodes())
}

func TestIPSetRemoveAll(t *testing.T) {
	set := IPSet{}

	set.InsertNet(Ten24)
	cidr1, _ := ParseNet("192.168.0.0/25")
	set.InsertNet(cidr1)
	assert.Equal(t, 2, set.tree.numNodes())

	cidr2, _ := ParseNet("0.0.0.0/0")
	set.RemoveNet(cidr2)
	assert.Equal(t, 0, set.tree.numNodes())
	assert.False(t, set.ContainsNet(Ten24))
	assert.False(t, set.ContainsNet(Ten24128))
	assert.False(t, set.ContainsNet(cidr1))
	assert.Equal(t, []error{}, set.tree.validate())
}

func TestIPSet_RemoveTop(t *testing.T) {
	testSet := IPSet{}
	ip1 := net.ParseIP("10.0.0.1")
	ip2 := net.ParseIP("10.0.0.2")

	testSet.Insert(ip2) // top
	testSet.Insert(ip1) // inserted at left
	testSet.Remove(ip2) // remove top node

	assert.True(t, testSet.Contains(ip1))
	assert.False(t, testSet.Contains(ip2))
	assert.Nil(t, testSet.tree.next())
	assert.Equal(t, []error{}, testSet.tree.validate())
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
	assert.Equal(t, []error{}, set.tree.validate())
}

func TestIPSetUnion(t *testing.T) {
	set1, set2 := &IPSet{}, &IPSet{}

	set1.InsertNet(Ten24)
	cidr, _ := ParseNet("192.168.0.248/29")
	set2.InsertNet(cidr)

	set := set1.Union(set2)
	assert.True(t, set.ContainsNet(Ten24))
	assert.True(t, set.ContainsNet(cidr))
	assert.Equal(t, []error{}, set.tree.validate())
}

func TestIPSetDifference(t *testing.T) {
	set1, set2 := &IPSet{}, &IPSet{}

	set1.InsertNet(Ten24)
	cidr, _ := ParseNet("192.168.0.248/29")
	set2.InsertNet(cidr)

	set := set1.Difference(set2)
	assert.True(t, set.ContainsNet(Ten24))
	assert.False(t, set.ContainsNet(cidr))
	assert.Equal(t, []error{}, set.tree.validate())
}

func TestIntersectionAinB1(t *testing.T) {
	case1 := []string{"10.0.16.0/20", "10.5.8.0/24", "10.23.224.0/23"}
	case2 := []string{"10.0.20.0/30", "10.5.8.0/29", "10.23.224.0/27"}
	output := []string{"10.23.224.0/27", "10.0.20.0/30", "10.5.8.0/29"}
	testIntersection(t, case1, case2, output)

}

func TestIntersectionAinB2(t *testing.T) {
	case1 := []string{"10.10.0.0/30", "10.5.8.0/29", "10.23.224.0/27"}
	case2 := []string{"10.10.0.0/20", "10.5.8.0/24", "10.23.224.0/23"}
	output := []string{"10.10.0.0/30", "10.5.8.0/29", "10.23.224.0/27"}
	testIntersection(t, case1, case2, output)
}

func TestIntersectionAinB3(t *testing.T) {
	case1 := []string{"10.0.5.0/24", "10.5.8.0/29", "10.23.224.0/27"}
	case2 := []string{"10.6.0.0/24", "10.9.9.0/29", "10.23.6.0/23"}
	output := []string{}
	testIntersection(t, case1, case2, output)
}

func TestIntersectionAinB4(t *testing.T) {
	case1 := []string{"10.23.6.0/24", "10.5.8.0/29", "10.23.224.0/27"}
	case2 := []string{"10.6.0.0/24", "10.9.9.0/29", "10.23.6.0/29"}
	output := []string{"10.23.6.0/29"}
	testIntersection(t, case1, case2, output)
}

func TestIntersectionAinB5(t *testing.T) {
	case1 := []string{"2001:db8:0:23::/96", "2001:db8:0:20::/96", "2001:db8:0:15::/96"}
	case2 := []string{"2001:db8:0:23::/64", "2001:db8:0:20::/64", "2001:db8:0:15::/64"}
	output := []string{"2001:db8:0:23::/96", "2001:db8:0:20::/96", "2001:db8:0:15::/96"}
	testIntersection(t, case1, case2, output)
}

func TestIntersectionAinB6(t *testing.T) {
	case1 := []string{"2001:db8:0:23::/64", "2001:db8:0:20::/64", "2001:db8:0:15::/64"}
	case2 := []string{"2001:db8:0:23::/96", "2001:db8:0:20::/96", "2001:db8:0:15::/96"}
	output := []string{"2001:db8:0:15::/96", "2001:db8:0:20::/96", "2001:db8:0:23::/96"}
	testIntersection(t, case1, case2, output)
}

func TestIntersectionAinB7(t *testing.T) {
	case1 := []string{"2001:db8:0:23::/64", "2001:db8:0:20::/64", "2001:db8:0:15::/64"}
	case2 := []string{"2001:db8:0:14::/96", "2001:db8:0:10::/96", "2001:db8:0:8::/96"}
	output := []string{}
	testIntersection(t, case1, case2, output)
}

func TestIntersectionAinB8(t *testing.T) {
	case1 := []string{"2001:db8:0:23::/64", "2001:db8:0:20::/64", "172.16.1.0/24"}
	case2 := []string{"2001:db9:0:14::/96", "2001:db9:0:10::/96", "172.16.1.0/28"}
	output := []string{"172.16.1.0/28"}
	testIntersection(t, case1, case2, output)
}

func TestIntersectionAinB9(t *testing.T) {
	case1 := []string{"10.5.8.0/29"}
	case2 := []string{"10.10.0.0/20", "10.5.8.0/24", "10.23.224.0/23"}
	output := []string{"10.5.8.0/29"}
	testIntersection(t, case1, case2, output)
}

func testIntersection(t *testing.T, input1 []string, input2 []string, output []string) {
	set1, set2, interSect := &IPSet{}, &IPSet{}, &IPSet{}
	for i := 0; i < len(input1); i++ {
		cidr, _ := ParseNet(input1[i])
		set1.InsertNet(cidr)
	}
	for j := 0; j < len(input2); j++ {
		cidr, _ := ParseNet(input2[j])
		set2.InsertNet(cidr)
	}
	for k := 0; k < len(output); k++ {
		cidr, _ := ParseNet(output[k])
		interSect.InsertNet(cidr)
	}
	set := set1.Intersection(set2)
	s1 := set.String()
	intSect := interSect.String()
	if !assert.Equal(t, intSect, s1) {
		t.Logf("\nEXPECTED: %s\nACTUAL: %s\n", intSect, s1)
	}
	assert.Equal(t, []error{}, set.tree.validate())
	assert.Equal(t, []error{}, interSect.tree.validate())

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
	assert.Equal(t, []error{}, set.tree.validate())
}

func TestIPSetAllocateDeallocate(t *testing.T) {
	rand.Seed(29)

	set := IPSet{}

	bigNet, _ := ParseNet("15.1.0.0/16")
	set.InsertNet(bigNet)

	ips := set.GetIPs(0)
	assert.Equal(t, 65536, len(ips))
	assert.Equal(t, big.NewInt(65536), set.tree.size())

	allocated := &IPSet{}
	for i := 0; i != 16384; i++ {
		allocated.Insert(ips[rand.Intn(65536)])
	}
	assert.Equal(t, big.NewInt(14500), allocated.tree.size())
	assert.Equal(t, []error{}, allocated.tree.validate())
	ips = allocated.GetIPs(0)
	assert.Equal(t, 14500, len(ips))
	for _, ip := range ips {
		assert.True(t, set.Contains(ip))
	}
	assert.Equal(t, []error{}, set.tree.validate())

	available := set.Difference(allocated)
	assert.Equal(t, big.NewInt(51036), available.tree.size())
	ips = available.GetIPs(0)
	for _, ip := range ips {
		assert.True(t, set.Contains(ip))
		assert.False(t, allocated.Contains(ip))
	}
	assert.Equal(t, 51036, len(ips))
	assert.Equal(t, []error{}, available.tree.validate())
}

func TestGetNetworks(t *testing.T) {
	s := &IPSet{}
	assert.Equal(t, []*net.IPNet{}, s.GetNetworks())
	s.InsertNet(Ten24)
	assert.Equal(t, "[10.0.0.0/24]", fmt.Sprintf("%s", s.GetNetworks()))
	ten25, _ := ParseNet("10.0.0.0/25")
	s.RemoveNet(ten25)
	assert.Equal(t, "[10.0.0.128/25]", fmt.Sprintf("%s", s.GetNetworks()))
	s.Remove(ParseIP("10.0.0.129"))
	assert.Equal(t, "[10.0.0.128/32 10.0.0.130/31 10.0.0.132/30 10.0.0.136/29 10.0.0.144/28 10.0.0.160/27 10.0.0.192/26]", fmt.Sprintf("%s", s.GetNetworks()))
}
