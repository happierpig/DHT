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
const tryTimes int = 4
const localAddress string = "127.0.0.1"
const WaitTime time.Duration = 250 * time.Millisecond // with use of select
const SleepTime time.Duration = 20 * time.Millisecond // avoiding endless for loop
const refreshTimeInterval time.Duration = 30 * time.Second
const expireTimeInterval2 time.Duration = 6 * time.Hour
const expireTimeInterval3 time.Duration = 20 * time.Minute
const republicTimeInterval time.Duration = 5 * time.Hour
const duplicateTimeInterval time.Duration = 15 * time.Minute
const backgroundInterval1 time.Duration = 5 * time.Second
const backgroundInterval2 time.Duration = 10 * time.Minute

type Contact struct {
	Address string
	NodeID  ID
}

type ContactRecord struct {
	SortKey     ID
	ContactInfo Contact
}

type RoutingTable struct {
	nodeAddr       Contact
	rwLock         deadlock.RWMutex
	buckets        [IDlength * 8]*list.List
	refreshIndex   int
	refreshTimeSet [IDlength * 8]time.Time
}

type Node struct {
	station   *network
	isRunning bool
	data      database
	table     RoutingTable
	addr      Contact
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
	Requester Contact
	Target    ID
}

type FindNodeReply struct {
	Requester Contact
	Replier   Contact
	Content   []ContactRecord
}

type database struct {
	rwLock        deadlock.RWMutex
	dataset       map[string]string
	expireTime    map[string]time.Time
	duplicateTime map[string]time.Time
	republicTime  map[string]time.Time
	privilege     map[string]int
}

type StoreRequest struct {
	Key          string
	Value        string
	RequesterPri int
	Requester    Contact
}

type FindValueRequest struct {
	Key       string
	Requester Contact
}

type FindValueReply struct {
	Requester Contact
	Replier   Contact
	Content   []ContactRecord
	IsFind    bool
	Value     string
}
