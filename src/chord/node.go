package chord

import (
	"errors"
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
	fingerTable   [hashBitsSize]string

	rwLock sync.RWMutex
	data   sync.Map // <string,string>

	station *network

	next int // for fix_finger
}

func (this *Node) Init(port int) {
	this.address = fmt.Sprintf("%s:%d", localAddress, port)
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
	this.next = 1
}

func (this *Node) Create() {
	this.predecessor = ""
	this.successorList[0] = this.address
	this.fingerTable[0] = this.address
	log.Infoln("Create new ring success in ", this.address)
}

func (this *Node) Join(addr string) bool {
	if isOnline := CheckOnline(addr); !isOnline {
		log.Warningln("<Join> Node in ", this.address, " fail to join network in ", addr, " for the network is failed")
		return false
	}
	var succAddr string
	err := RemoteCall(addr, "WrapNode.FindSuccessor", this.ID, &succAddr)
	if err != nil {
		log.Errorln("<Join> Fail to Join ,Error msg: ", err)
		return false
	}
	log.Infoln("<Join> Get Successor and Join Successfully ")
	var temp [successorListSize]string
	err = RemoteCall(succAddr, "WrapNode.SetSuccessorList", 2021, &temp)
	if err != nil {
		log.Errorln("<Join> Fail to Join ,Can't get successorList: ", err)
		return false
	}
	this.rwLock.Lock()
	this.predecessor = ""
	this.successorList[0] = succAddr
	this.fingerTable[0] = succAddr
	for i := 1; i < successorListSize; i++ {
		this.successorList[i] = temp[i-1]
	}
	this.rwLock.Unlock()
	return true
}

func (this *Node) Quit() {

}

func (this *Node) ForceQuit() {

}

func (this *Node) Ping(addr string) bool {
	return CheckOnline(addr)
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

// below are private functions todo(FirstValidSuccessor)
func (this *Node) find_successor(target *big.Int, result *string) error {
	var succAddr string
	this.first_online_successor(&succAddr)
	if contain(target, this.ID, ConsistentHash(succAddr), true) {
		*result = succAddr
		return nil
	}
	closestPre := this.closest_preceding_node(target)
	return RemoteCall(closestPre, "WrapNode.FindSuccessor", target, result)
}

func (this *Node) closest_preceding_node(target *big.Int) string {
	for i := hashBitsSize - 1; i >= 0; i-- {
		if this.fingerTable[i] == "" {
			continue
		}
		if !CheckOnline(this.fingerTable[i]) {
			continue
		}
		if contain(ConsistentHash(this.fingerTable[i]), this.ID, target, false) { // get ( , )
			log.Infoln("Find closest_preceding_node Successfully in Node ", this.address)
			return this.fingerTable[i]
		}
	}
	// means successor fail
	var preaddr string
	err := this.first_online_successor(&preaddr)
	if err != nil {
		log.Errorln("<closest_preceding_node> List Break")
		return ""
	}
	return preaddr
}

func (this *Node) set_successor_list(result *[successorListSize]string) error {
	this.rwLock.RLock()
	*result = this.successorList
	this.rwLock.RUnlock()
	return nil
}

func (this *Node) get_predecessor(result *string) error {
	this.rwLock.RLock()
	*result = this.predecessor
	this.rwLock.RUnlock()
	return nil
}

func (this *Node) first_online_successor(result *string) error {
	for i := 0; i < successorListSize; i++ {
		if CheckOnline(this.successorList[i]) {
			*result = this.successorList[i]
			return nil
		}
	}
	log.Errorln("<first_online_successor> List Break")
	return errors.New("List Break")
}

func (this *Node) stabilize() {
	var succPredAddr string
	var newSucAddr string
	this.first_online_successor(&newSucAddr)
	err := RemoteCall(newSucAddr, "WrapNode.GetPredecessor", 2021, &succPredAddr)
	if err != nil {
		log.Errorln("<stabilize> fail to get predecessor in ", newSucAddr)
		return
	}
	if succPredAddr != "" && contain(ConsistentHash(succPredAddr), this.ID, ConsistentHash(newSucAddr), false) { // change to new successor
		newSucAddr = succPredAddr
	}
	// modify successorList
	var temp [successorListSize]string
	err = RemoteCall(newSucAddr, "WrapNode.SetSuccessorList", 2021, &temp)
	if err != nil {
		log.Errorln("<stabilize> Fail to stabilize ,Can't get successorList:  ", err)
		return
	}
	this.rwLock.Lock()
	this.successorList[0] = newSucAddr
	this.fingerTable[0] = newSucAddr
	for i := 1; i < successorListSize; i++ {
		this.successorList[i] = temp[i-1]
	}
	this.rwLock.Unlock()
	err = RemoteCall(newSucAddr, "WrapNode.Notify", this.address, &this.address)
	if err != nil {
		log.Errorln("<stabilize> Fail to notify ,Can't get call successor :  ", err)
		return
	}
}

func (this *Node) notify(instructor string) error {
	if this.predecessor == instructor {
		return nil
	}
	if this.predecessor == "" || contain(ConsistentHash(instructor), ConsistentHash(this.predecessor), this.ID, false) {
		this.rwLock.Lock()
		this.predecessor = instructor
		this.rwLock.Unlock()
		log.Infoln("<notify> Change ", this.address, " Predecessor to ", instructor)
	}
	return nil
}

func (this *Node) check_predecessor() {
	if this.predecessor != "" && !CheckOnline(this.predecessor) {
		this.rwLock.Lock()
		this.predecessor = ""
		this.rwLock.Unlock()
		log.Infoln("<check_predecessor> Find failed predecessor :)")
	}
}

func (this *Node) fix_finger() {
	var ans string
	err := this.find_successor(calculateID(this.ID, this.next), &ans)
	if err != nil {
		log.Errorln("<fix_finger> error occurs")
		return
	}
	this.rwLock.Lock()
	this.fingerTable[0] = this.successorList[0]
	this.fingerTable[this.next] = ans
	this.next += 1
	if this.next >= hashBitsSize {
		this.next = 1
	}
	this.rwLock.Unlock()
}
