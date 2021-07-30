package kademlia

import (
	"container/list"
	"github.com/sasha-s/go-deadlock"
	"net"
	"net/rpc"
	"time"
)

const IDlength int = 20 // bytes
type ID [IDlength]byte

const K int = 20 // buckets size
const alpha int32 = 3
const WaitTime time.Duration = 200 * time.Millisecond
const tryTimes int = 3
const localAddress string = "127.0.0.1"

type Contact struct {
	address string
	nodeID  ID
}

type ContactRecord struct {
	sortKey ID
	contact Contact
}

type RoutingTable struct {
	nodeID  ID
	rwLock  deadlock.RWMutex
	buckets [IDlength * 8]*list.List
}

type Node struct {
	station   *network
	isRunning bool

	table RoutingTable
	addr  Contact
}

type WrapNode struct {
	node *Node
}

type network struct {
	serv       *rpc.Server
	lis        net.Listener
	nodePtr    *WrapNode
	QuitSignal chan bool
}

type FindNodeRequest struct {
	requester Contact
	target    ID
}

type FindNodeReply struct {
	requester Contact
	replier   Contact
	content   []ContactRecord
}
