package chord

import (
	"fmt"
	"math/big"
	"sync"
)

const successorListSize int = 5
const hashBitsSize int = 160 // fingerTable: 0 - 159

type Node struct {
	address string
	ID      *big.Int

	conRoutineFlag bool

	successorList [successorListSize]string
	predecessor   string
	fingerTable   [hashBitsSize]string

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

}

func (this *Node) Create() {

}

func (this *Node) Join(addr string) bool {
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
