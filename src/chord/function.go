package chord

import (
	"crypto/sha1"
	"math/big"
	"net"
)

func ConsistentHash(raw string) *big.Int {
	hash := sha1.New()
	hash.Write([]byte(raw))
	return (&big.Int{}).SetBytes(hash.Sum(nil))
}

// GetLocalAddress function to get local address(ip address)
func GetLocalAddress() string {
	var localaddress string
	ifaces, err := net.Interfaces()
	if err != nil {
		panic("init: failed to find network interfaces")
	}
	// find the first non-loopback interface with an IP address
	for _, elt := range ifaces {
		if elt.Flags&net.FlagLoopback == 0 && elt.Flags&net.FlagUp != 0 {
			addrs, err := elt.Addrs()
			if err != nil {
				panic("init: failed to get addresses for network interface")
			}

			for _, addr := range addrs {
				ipnet, ok := addr.(*net.IPNet)
				if ok {
					if ip4 := ipnet.IP.To4(); len(ip4) == net.IPv4len {
						localaddress = ip4.String()
						break
					}
				}
			}
		}
	}
	if localaddress == "" {
		panic("init: failed to find non-loopback interface with valid address on this node")
	}
	return localaddress
}

//contain:mode true -- ( ]   mode false -- ( )
func contain(target, start, end *big.Int, mode bool) bool {
	if mode {
		if end.Cmp(start) == 0 {
			return true
		}
		if end.Cmp(start) > 0 {
			return (target.Cmp(start) > 0) && (end.Cmp(target) >= 0)
		} else if end.Cmp(start) < 0 {
			return (target.Cmp(start) > 0) || (end.Cmp(target) >= 0)
		}
	} else {
		if end.Cmp(start) == 0 {
			return false
		}
		if end.Cmp(start) > 0 {
			return (target.Cmp(start) > 0) && (end.Cmp(target) > 0)
		} else if end.Cmp(start) < 0 {
			return (target.Cmp(start) > 0) || (end.Cmp(target) > 0)
		}
	}
	return false
}
