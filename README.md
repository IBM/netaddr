# netaddr package for go

This repo contains a library to complement the [go net library][net] and
provides containers and utilities like in python's [netaddr].

Please see the [api documentation] for details. [The authoritative source for
this library is found on github][source]. We encourage importing this code
using the stable, versioned URL provided by [gopkg.in][gopkg]. Once imported,
refer to it as `netaddr` in your code (without the version).

    import "gopkg.in/netaddr.v1"

## IP Maps

This is a data structure that maps IP addresses to arbitrary `interface{}`
values. It supports the constant-time basic map operations: insert, get, and
remove. It also supports O(n) iteration over all prefix/value pairs in
lexigraphical order of prefixes.

When a map is created, you choose whether the prefixes will be IPv4 (4-byte
representation only) or IPv6 (16-byte) addresses. The two families cannot be
mixed in the same map instance. This is consistent with this library's stance on
not conflating IPv4 with 16-byte IPv4 in IPv6 representation.

Since this data structure was specifically designed to use IP addresses as keys,
it supports a couple of other cool operations.

First, it can efficiently perform a longest prefix match when an exact match is
not available. This operation has the same O(1) efficiency as an exact match.

Second, it supports aggregation of key/values while iterating on the fly. This
has nearly the same O(n) efficiency \*\* as iterating without aggregating. The
rules of aggregation are as follows:

1. The values stored must be comparable. Prefixes get aggregated only where
   their values compare equal.
2. The set of key/value pairs visited is the minimal-size set such that any
   longest prefix match against the aggregated set will always return the same
   value as the same match against the non-aggregated set.
3. The aggregated and non-aggregated sets of prefixes may be disjoint.

Aggregation can be useful, for example, to minimize the number of prefixes
needed to install into a router's datapath to guarantee that all of the next
hops are correct. In general, though, routing protocols should be careful when
passing aggregated routes to neighbors as this will likely lead to poor
comparisions by neighboring routers who receive routes aggregated differently
from different peers.

A future enhancement could efficiently compute the difference in the aggregated
set when inserting or removing elements so that the entire set doesn't need to
be iterated after each mutation. Since the aggregated set of prefixes is
disjoint from the original, either operation could result in both adding and
removing key/value pairs. This makes it tricky but it should be possible.

As a simple example, consider the following key/value pairs inserted into a map.

- 10.224.24.2/31 / true
- 10.224.24.0/32 / true
- 10.224.24.1/32 / true

When iterating over the aggregated set, only the following key/value pair will
be visited.

- 10.224.24.0/30 / true

A slightly more complex example shows how value comparison comes into play.

- 10.224.24.0/30 / true
- 10.224.24.0/31 / false
- 10.224.24.1/32 / true
- 10.224.24.0/32 / false

Iterating the aggregated set:

- 10.224.24.0/30 / true
- 10.224.24.0/31 / false
- 10.224.24.1/32 / true

A more complex example where all values are the same (so they aren't shown)

- 172.21.0.0/20
- 192.68.27.0/25
- 192.168.26.128/25
- 10.224.24.0/32
- 192.68.24.0/24
- 172.16.0.0/12
- 192.68.26.0/24
- 10.224.24.0/30
- 192.168.24.0/24
- 192.168.25.0/24
- 192.168.26.0/25
- 192.68.25.0/24
- 192.168.27.0/24
- 172.20.128.0/19
- 192.68.27.128/25

The aggregrated set is as follows:

- 10.224.24.0/30
- 172.16.0.0/12
- 192.68.24.0/22
- 192.168.24.0/22

\*\* There is one complication that may throw its efficiency slightly off of
     O(n) but I haven't analyzed it yet to be sure. It should be pretty close.

## comparison with python's netaddr

This netaddr library was written to complement the existing [net] package in go
just filling in a few gaps that existed. See the table below for a side-by-side
comparison of python netaddr features and the corresponding features in this
library or elsewhere in go packages.

| Python netaddr | Go                                |
|----------------|-----------------------------------|
| EUI            | ???                               |
| IPAddress      | Use [IP] from [net]\*             |
| IPNetwork      | Use [IPNet] from [net]\*\*        |
| IPSet          | Use [IPSet]                       |
| IPRange        | Use [IPRange]                     |
| IPGlob         | Not yet implemented               |

\* The [net] package in golang parses IPv4 address as IPv4 encoded IPv6
addresses. I found this design choice frustrating. Hence, there is a [ParseIP]
in this package that always parses IPv4 as 4 byte addresses.

\*\* This package provides a few extra convenience utilities for [IPNet]. See
[ParseNet], [NetSize], [BroadcastAddr], and [NetworkAddr].

## help

This needs a lot of work. Help if you can!

- More test coverage

[netaddr]: https://netaddr.readthedocs.io/en/latest/installation.html
[net]: https://golang.org/pkg/net/
[api documentation]: https://godoc.org/gopkg.in/netaddr.v1
[source]: https://github.com/IBM/netaddr/
[gopkg]: https://gopkg.in/netaddr.v1
[IP]: https://golang.org/pkg/net/#IP
[IPNet]: https://golang.org/pkg/net/#IPNet
[IPSet]: https://godoc.org/gopkg.in/netaddr.v1#IPSet
[ParseIP]: https://godoc.org/gopkg.in/netaddr.v1#ParseIP
[ParseNet]: https://godoc.org/gopkg.in/netaddr.v1#ParseNet
[NetSize]: https://godoc.org/gopkg.in/netaddr.v1#NetSize
[BroadcastAddr]: https://godoc.org/gopkg.in/netaddr.v1#BroadcastAddr
[NetworkAddr]: https://godoc.org/gopkg.in/netaddr.v1#NetworkAddr
