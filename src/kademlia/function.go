package kademlia

import (
	"crypto/sha1"
	"encoding/hex"
	log "github.com/sirupsen/logrus"
	"net"
	"sort"
)

func NewID(hashValue string) (ID, error) {
	var result ID
	raw, err := hex.DecodeString(hashValue)
	if err != nil {
		log.Errorln("<NewID> Fail to generate ID due to ", err)
		return result, err
	}
	for i := 0; i < IDlength; i++ {
		result[i] = raw[i]
	}
	return result, nil
}

func NewContact(address string) Contact {
	nodeID, _ := NewID(address)
	return Contact{address: address, nodeID: nodeID}
}

func (IDValue ID) String() string {
	return hex.EncodeToString(IDValue[0:IDlength])
}

func Hash(raw string) (result [IDlength]byte) {
	hash := sha1.New()
	hash.Write([]byte(raw))
	tmp := hash.Sum(nil)
	for i := 0; i < IDlength; i++ {
		result[i] = tmp[i]
	}
	return
}

func (this ID) Equals(other ID) bool {
	for i := 0; i < IDlength; i++ {
		if this[i] != other[i] {
			return false
		}
	}
	return true
}

func (this ID) LessThan(other ID) bool {
	for i := 0; i < IDlength; i++ {
		if this[i] == other[i] {
			continue
		}
		if this[i] < other[i] {
			return true
		}
		if this[i] > other[i] {
			return false
		}
	}
	return false
}

func Xor(one ID, other ID) (result ID) {
	for i := 0; i < IDlength; i++ {
		result[i] = one[i] ^ other[i]
	}
	return
}

func PrefixLen(IDvalue ID) int {
	for i := 0; i < IDlength; i++ {
		for j := 0; j <= 7; j++ {
			if IDvalue[i]&(0b1<<(7-j)) == 0 {
				return i*8 + j
			}
		}
	}
	return IDlength*8 - 1
}

func SliceSort(dataSet *[]ContactRecord) {
	sort.Slice(*dataSet, func(i, j int) bool {
		return (*dataSet)[i].sortKey.LessThan((*dataSet)[j].sortKey)
	})
}

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
