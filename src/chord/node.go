package chord

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"math/big"
	"sync"
)

const successorListSize int = 5
const hashBitsSize int = 160 // fingerTable: 0 - 160

type Node struct {
	address string   // NewNode()-> Init()
	ID      *big.Int // NewNode()-> Init()

	conRoutineFlag bool // Init() / Run()

	successorList [successorListSize]string
	predecessor   string
	fingerTable   [hashBitsSize + 1]string

	rwLock sync.RWMutex
	data   sync.Map // <string,string>

	station *network
}

func (this *Node) Init(port int) {
	this.address = fmt.Sprintf("%s:%d", GetLocalAddress(), port)
	this.ID = ConsistentHash(this.address)
	this.conRoutineFlag = false
}

func (this *Node) Run() {
	this.station = new(network)
	err := this.station.Init(this.address, this)
	if err != nil {
		log.Errorln("Run failed ", err)
		return
	}
	log.Infoln("Run success in ", this.address)
	this.conRoutineFlag = true
}

func (this *Node) Create() {
	this.predecessor = ""
	this.successorList[0] = this.address
	log.Infoln("Create new ring success in ", this.address)
}

func (this *Node) Join(addr string) bool {
	if isOnline := CheckOnline(addr); !isOnline {
		log.Warningln("Node in ", this.address, " fail to join network in ", addr)
		return false
	}

	return true
}

func (this *Node) Quit() {

}

func (this *Node) ForceQuit() {

}

func (this *Node) Ping(addr string) bool {
	return true
}

func (this *Node) Put(key string, value string) bool {
	return true
}

func (this *Node) Get(key string) (bool, string) {
	return true, ""
}

func (this *Node) Delete(key string) bool {
	return true
}
