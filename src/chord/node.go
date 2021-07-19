package chord

import (
	"errors"
	"fmt"
	log "github.com/sirupsen/logrus"
	"math/big"
	"sync"
	"time"
)

const successorListSize int = 5
const hashBitsSize int = 160 // fingerTable: 0 - 159

type Node struct {
	address string   // NewNode()-> Init()
	ID      *big.Int // NewNode()-> Init()

	conRoutineFlag bool // Init() / Run()

	successorList [successorListSize]string
	predecessor   string
	fingerTable   [hashBitsSize]string

	rwLock   sync.RWMutex
	dataSet  map[string]string
	dataLock sync.RWMutex

	station *network

	next int // for fix_finger
}

func (this *Node) Init(port int) {
	this.address = fmt.Sprintf("%s:%d", localAddress, port)
	this.ID = ConsistentHash(this.address)
	this.conRoutineFlag = false
	this.dataSet = make(map[string]string)
}

func (this *Node) Run() {
	this.station = new(network)
	err := this.station.Init(this.address, this)
	if err != nil {
		log.Errorln("<Run> failed ", err)
		return
	}
	log.Infoln("<Run> success in ", this.address)
	this.conRoutineFlag = true
	this.next = 1
}

func (this *Node) Create() {
	this.predecessor = ""
	this.successorList[0] = this.address
	this.fingerTable[0] = this.address
	this.background()
	log.Infoln("<Create> new ring success in ", this.address)
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
	err = RemoteCall(succAddr, "WrapNode.HereditaryData", this.address, &this.dataSet)
	if err != nil {
		log.Errorln("<Join> Fail to Join ,Can't share Data: ", err)
		return false
	}
	this.background()
	//fmt.Println("[debug] ",this.address," Join Successfully") // debug
	return true
}

func (this *Node) Quit() {
	err := this.station.ShutDown()
	if err != nil {
		log.Errorln("<Quit> fail to quit in ", this.address)
	}
	this.rwLock.Lock()
	this.conRoutineFlag = false // important
	this.next = 1
	this.rwLock.Unlock()
	var succAddr string
	var occupy string
	this.first_online_successor(&succAddr)
	err = RemoteCall(succAddr, "WrapNode.InheritData", &this.dataSet, &occupy)
	if err != nil {
		log.Errorln("<Quit.InheritData> Error : ", err)
	}
	err = RemoteCall(succAddr, "WrapNode.CheckPredecessor", 2021, &occupy)
	if err != nil {
		log.Errorln("<Quit.CheckPredecessor> Error : ", err)
	}
	err = RemoteCall(this.predecessor, "WrapNode.Stablize", 2021, &occupy)
	if err != nil {
		log.Errorln("<Quit.Stablize> Error : ", err)
	}
	this.dataSet = make(map[string]string)
	log.Infoln("<Quit> ", this.address, " Quit Successfully ;)")
	//fmt.Println("[debug] ",this.address," Quit Successfully") // debug
}

func (this *Node) ForceQuit() {
	err := this.station.ShutDown()
	if err != nil {
		log.Errorln("<Quit> fail to quit in ", this.address)
	}
	this.rwLock.Lock()
	this.conRoutineFlag = false
	this.next = 1
	this.rwLock.Unlock()
	log.Infoln("<ForceQuit> ", this.address, " Quit Successfully ;)")
}

func (this *Node) Ping(addr string) bool {
	return CheckOnline(addr)
}

func (this *Node) Put(key string, value string) bool {
	if !this.conRoutineFlag {
		log.Errorln("<Put> The Node is sleeping.. (", this.address)
		return false
	}
	var targetAddr string
	err := this.find_successor(ConsistentHash(key), &targetAddr)
	if err != nil {
		log.Errorln("<Put> Fail to Find Key's Successor...")
		return false
	}
	var occupy string
	var dataPair Pair = Pair{key, value}
	err = RemoteCall(targetAddr, "WrapNode.StoreData", dataPair, &occupy)
	if err != nil {
		log.Errorln("<Put> Fail to Put Data into Target Node in ", targetAddr, " | error: ", err)
		return false
	}
	log.Infoln("<Put> Put Data into ", targetAddr, " successfully. :)")
	return true
}

func (this *Node) Get(key string) (bool, string) {
	if !this.conRoutineFlag {
		log.Errorln("<Get> The node is sleeping.. (", this.address)
		return false, ""
	}
	var value string
	var targetAddr string
	err := this.find_successor(ConsistentHash(key), &targetAddr)
	if err != nil {
		log.Errorln("<Get> Fail to Find Key's Successor...")
		return false, ""
	}
	err = RemoteCall(targetAddr, "WrapNode.GetData", key, &value)
	if err != nil {
		log.Errorln("<Get> Fail to Get Data from Target Node in ", targetAddr, " | error: ", err)
		return false, ""
	}
	log.Infoln("<Get> Get Data from ", targetAddr, " successfully. :)")
	return true, value
}

func (this *Node) Delete(key string) bool {
	if !this.conRoutineFlag {
		log.Errorln("<Delete> The node is sleeping.. (", this.address)
		return false
	}
	var targetAddr string
	err := this.find_successor(ConsistentHash(key), &targetAddr)
	if err != nil {
		log.Errorln("<Delete> Fail to Find Key's Successor...")
		return false
	}
	var occupy string
	err = RemoteCall(targetAddr, "WrapNode.DeleteData", key, &occupy)
	if err != nil {
		log.Errorln("<Delete> Fail to Delete Data from Target Node in ", targetAddr, " | error: ", err)
		return false
	}
	log.Infoln("<Delete> Delete Data from ", targetAddr, " successfully. :)")
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
			log.Infoln("<closest_preceding_node>Find closest_preceding_node Successfully in Node ", this.address)
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
	log.Errorln("<first_online_successor> List Break in ", this.address)
	return errors.New("List Break")
}

func (this *Node) stabilize() {
	var succPredAddr string
	var newSucAddr string
	this.first_online_successor(&newSucAddr)
	err := RemoteCall(newSucAddr, "WrapNode.GetPredecessor", 2021, &succPredAddr)

	//todo: here ensure the newSucAddr's predecessor is Online or nil may increase efficiency

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
	var occupy string
	err = RemoteCall(newSucAddr, "WrapNode.Notify", this.address, &occupy)
	if err != nil {
		log.Errorln("<stabilize> Fail to notify ,Can't get call successor :  ", err)
		return
	}
	log.Infoln("<stablize> successfully :) in ", this.address)
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

func (this *Node) check_predecessor() error {
	if this.predecessor != "" && !CheckOnline(this.predecessor) {
		this.rwLock.Lock()
		this.predecessor = ""
		this.rwLock.Unlock()
		log.Infoln("<check_predecessor> Find failed predecessor :)")
	}
	return nil
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
	log.Infoln("<fix_finger> fix successfully :) in ", this.address)
}

func (this *Node) background() {
	go func() {
		for this.conRoutineFlag {
			this.stabilize()
			time.Sleep(timeCut)
		}
	}()
	go func() {
		for this.conRoutineFlag {
			this.check_predecessor()
			time.Sleep(timeCut)
		}
	}()
	go func() {
		for this.conRoutineFlag {
			this.fix_finger()
			time.Sleep(timeCut)
		}
	}()
}

func (this *Node) store_data(dataPair Pair) error {
	this.dataLock.Lock()
	this.dataSet[dataPair.Key] = dataPair.Value
	this.dataLock.Unlock()
	//fmt.Println("[debug] Store ",dataPair," into ",this.address) // debug
	return nil
}

func (this *Node) get_data(key string, value *string) error {
	this.dataLock.RLock()
	tmp, ok := this.dataSet[key]
	this.dataLock.RUnlock()
	if ok {
		*value = tmp
		return nil
	} else {
		*value = ""
		//fmt.Println("[debug] Unsuccessfully Get ",key," from ",this.address) // debug
		return errors.New("<get_data> Unreachable Data")
	}
}

func (this *Node) delete_data(key string) error {
	this.dataLock.Lock()
	_, ok := this.dataSet[key]
	if ok {
		delete(this.dataSet, key)
	}
	this.dataLock.Unlock()
	if ok {
		//fmt.Println("[debug] Successfully Delete ",key," from ",this.address) // debug
		return nil
	} else {
		//fmt.Println("[debug] Unsuccessfully Delete ",key," from ",this.address) // debug
		return errors.New("<delete_data> Unreachable Data")
	}
}

func (this *Node) hereditary_data(predeAddr string, dataSet *map[string]string) error { // join
	this.dataLock.Lock()
	for k, v := range this.dataSet {
		if !contain(ConsistentHash(k), ConsistentHash(predeAddr), this.ID, true) {
			(*dataSet)[k] = v
			delete(this.dataSet, k)
			//fmt.Println("[debug] Successfully Move ",k," from ",this.address," to ",predeAddr) // debug
		}
	}
	this.dataLock.Unlock()
	log.Infoln("<hereditary_data> Successfully pass data from ", this.address)
	return nil
}

func (this *Node) inherit_data(dataSet *map[string]string) error {
	this.dataLock.Lock()
	for k, v := range *dataSet {
		this.dataSet[k] = v
		//fmt.Println("[debug] Successfully inherit ",k," to ",this.address) // debug
	}
	this.dataLock.Unlock()
	log.Infoln("<inherit_data> Successfully pass data to ", this.address)
	return nil
}
