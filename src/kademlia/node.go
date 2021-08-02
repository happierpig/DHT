package kademlia

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"sync/atomic"
	"time"
)

func (this *Node) reset() {
	this.isRunning = false
	this.table.InitRoutingTable(this.addr.NodeID)
}

func (this *Node) Init(port int) {
	this.addr = NewContact(fmt.Sprintf("%s:%d", localAddress, port))
	this.reset()
}

func (this *Node) Run() {
	this.station = new(network)
	err := this.station.Init(this.addr.Address, this)
	if err != nil {
		log.Errorln("<Run> Fail to Run ", this.addr.Address)
		return
	}
	log.Infoln("<Run> Successfully Run Node in ", this.addr.Address)
	this.isRunning = true
}

func (this *Node) Join(address string) bool {
	bootstrap := new(Contact)
	*bootstrap = NewContact(address)
	if isOnline := CheckOnline(this, bootstrap); !isOnline {
		log.Warningln("<Join> Node in ", this.addr.Address, " fail to join network in ", address, " for the network is failed")
		return false
	}
	this.table.Update(bootstrap)
	this.FindClosestNode(this.addr.NodeID)
	return true
}

func (this *Node) Quit() {
	this.station.ShutDown()
	this.reset()
}

func (this *Node) Ping(requester Contact) {
	log.Infoln("<Ping>", this.addr.Address, "is Ping by ", requester.Address)
}

func (this *Node) FindClosestNode(target ID) []ContactRecord {
	resultList := make([]ContactRecord, 0, K+2)
	pendingList := this.table.FindClosest(target, K)
	inRun := new(int32)
	*inRun = 0
	visit := make(map[string]bool)
	visit[this.addr.Address] = true
	index := 0
	ch := make(chan FindNodeReply, alpha+3)
	for index < len(pendingList) && *inRun < alpha {
		tmpReplier := pendingList[index].ContactInfo
		if _, ok := visit[tmpReplier.Address]; !ok {
			visit[tmpReplier.Address] = true
			atomic.AddInt32(inRun, 1)
			go func(Replier *Contact, channel chan FindNodeReply) {
				var response FindNodeReply
				err := RemoteCall(this, Replier, "WrapNode.GetClose", FindNodeRequest{this.addr, target}, &response)
				if err != nil {
					atomic.AddInt32(inRun, -1)
					log.Warnln("<FindClosestNode> Fail due to  ", err)
					return
				}
				channel <- response
				return
			}(&tmpReplier, ch)
		}
		index++
	}
	for index < len(pendingList) || *inRun > 0 {
		if *inRun > 0 {
			select {
			case response := <-ch:
				atomic.AddInt32(inRun, -1)
				resultList = append(resultList, ContactRecord{Xor(response.Replier.NodeID, target), response.Replier})
				for _, v := range response.Content {
					pendingList = append(pendingList, v)
				}
			case <-time.After(WaitTime):
				log.Infoln("<FindClosestNode> Avoid Blocking...")
			}
		}
		for index < len(pendingList) && *inRun < alpha {
			tmpReplier := pendingList[index].ContactInfo
			if _, ok := visit[tmpReplier.Address]; !ok {
				visit[tmpReplier.Address] = true
				atomic.AddInt32(inRun, 1)
				go func(Replier *Contact, channel chan FindNodeReply) {
					var response FindNodeReply
					err := RemoteCall(this, Replier, "WrapNode.GetClose", FindNodeRequest{this.addr, target}, &response)
					if err != nil {
						atomic.AddInt32(inRun, -1)
						log.Warnln("<FindClosestNode> Fail due to  ", err)
						return
					}
					channel <- response
					return
				}(&tmpReplier, ch)
			}
			index++
		}
	}
	SliceSort(&resultList)
	if len(resultList) > K {
		resultList = resultList[:K]
	}
	return resultList
}

func (this *Node) Refresh() {
	lastRefreshTime := this.table.refreshTimeSet[this.table.refreshIndex]
	if !lastRefreshTime.Add(refreshTimeInterval).After(time.Now()) {
		//use refreshIndex to construct a new ID
		tmpID := GenerateID(this.addr.NodeID, this.table.refreshIndex)
		this.FindClosestNode(tmpID)
		this.table.refreshTimeSet[this.table.refreshIndex] = time.Now()
	}
	this.table.refreshIndex = (this.table.refreshIndex + 1) % (IDlength * 8)
}
