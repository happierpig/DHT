package kademlia

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"sync/atomic"
	"time"
)

func (this *Node) reset() {
	this.isRunning = false
	this.table.InitRoutingTable(this.addr)
	this.data.init()
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
	this.Background()
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

func (this *Node) Ping(requester string) bool {
	//log.Infoln("<Ping>",this.addr.Address,"is Ping by ",requester.Address)
	return true
}

func (this *Node) FindClosestNode(target ID) []ContactRecord {
	resultList := make([]ContactRecord, 0, K*2)
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
				SliceSort(&pendingList)
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

func (this *Node) Put(key string, value string) bool {
	request := StoreRequest{key, value, root, this.addr}
	this.data.store(request)
	request.RequesterPri = publisher
	this.RangePut(request)
	return true
}

func (this *Node) RangePut(request StoreRequest) {
	pendingList := this.FindClosestNode(Hash(request.Key))
	count := new(int32)
	*count = 0
	for index := 0; index < len(pendingList); {
		if *count < alpha {
			target := pendingList[index].ContactInfo
			index++
			atomic.AddInt32(count, 1)
			go func(input StoreRequest, targetNode *Contact) {
				var occupy string
				err := RemoteCall(this, targetNode, "WrapNode.Store", input, &occupy)
				if err != nil {
					log.Warningln("<RangePut> Fail to Put ")
				}
				atomic.AddInt32(count, -1)
			}(request, &target)
		} else {
			time.Sleep(SleepTime)
		}
	}
}

func (this *Node) Get(key string) (bool, string) {
	isFind := false
	reply := ""
	requestInfo := FindValueRequest{key, this.addr}
	resultList := make([]Contact, 0, K*2)
	pendingList := this.table.FindClosest(Hash(key), K)
	inRun := new(int32)
	*inRun = 0
	visit := make(map[string]bool)
	visit[this.addr.Address] = true
	index := 0
	ch := make(chan FindValueReply, alpha+3)
	for index < len(pendingList) && *inRun < alpha {
		tmpReplier := pendingList[index].ContactInfo
		if _, ok := visit[tmpReplier.Address]; !ok {
			visit[tmpReplier.Address] = true
			atomic.AddInt32(inRun, 1)
			go func(Replier *Contact, channel chan FindValueReply) {
				var response FindValueReply
				err := RemoteCall(this, Replier, "WrapNode.FindValue", requestInfo, &response)
				if err != nil {
					atomic.AddInt32(inRun, -1)
					log.Warnln("<Get> Fail due to  ", err)
					return
				}
				channel <- response
				return
			}(&tmpReplier, ch)
		}
		index++
	}
	for (index < len(pendingList) || *inRun > 0) && !isFind {
		if *inRun > 0 {
			select {
			case response := <-ch:
				atomic.AddInt32(inRun, -1)
				if response.IsFind {
					isFind = true
					reply = response.Value
					break
				}
				resultList = append(resultList, response.Replier)
				for _, v := range response.Content {
					pendingList = append(pendingList, v)
				}
				SliceSort(&pendingList) // efficiency
			case <-time.After(WaitTime):
				log.Infoln("<Get> Avoid Blocking...")
			}
			if isFind {
				break
			}
		}
		for index < len(pendingList) && *inRun < alpha && !isFind {
			tmpReplier := pendingList[index].ContactInfo
			if _, ok := visit[tmpReplier.Address]; !ok {
				visit[tmpReplier.Address] = true
				atomic.AddInt32(inRun, 1)
				go func(Replier *Contact, channel chan FindValueReply) {
					var response FindValueReply
					err := RemoteCall(this, Replier, "WrapNode.FindValue", requestInfo, &response)
					if err != nil {
						atomic.AddInt32(inRun, -1)
						log.Warnln("<Get> Fail due to  ", err)
						return
					}
					channel <- response
					return
				}(&tmpReplier, ch)
			}
			index++
		}
	}
	if !isFind {
		return false, ""
	} else {
		StoreInfo := StoreRequest{key, reply, duplicater, this.addr}
		count := new(int32)
		*count = 0
		for i := 0; i < len(resultList); {
			if *count < alpha {
				target := resultList[i]
				i++
				atomic.AddInt32(count, 1)
				go func(input StoreRequest, targetNode *Contact) {
					var occupy string
					err := RemoteCall(this, targetNode, "WrapNode.Store", input, &occupy)
					if err != nil {
						log.Warningln("<RangePut> Fail to Put ")
					}
					atomic.AddInt32(count, -1)
				}(StoreInfo, &target)
			} else {
				time.Sleep(SleepTime)
			}
		}
		return true, reply
	}
}

func (this *Node) Republic() {
	pendingList := this.data.republic()
	for k, v := range pendingList {
		request := StoreRequest{k, v, publisher, this.addr}
		this.RangePut(request)
	}
}

func (this *Node) Duplicate() {
	pendingList := this.data.duplicate()
	for k, v := range pendingList {
		request := StoreRequest{k, v, duplicater, this.addr}
		this.RangePut(request)
	}
}

func (this *Node) Expire() {
	this.data.expire()
}

func (this *Node) Background() {
	go func() {
		for this.isRunning {
			this.Refresh()
			time.Sleep(backgroundInterval1)
		}
	}()
	go func() {
		for this.isRunning {
			this.Duplicate()
			time.Sleep(backgroundInterval2)
		}
	}()
	go func() {
		for this.isRunning {
			this.Expire()
			time.Sleep(backgroundInterval2)
		}
	}()
	go func() {
		for this.isRunning {
			this.Republic()
			time.Sleep(backgroundInterval2)
		}
	}()
}

// unused function
func (this *Node) Create() {

}

func (this *Node) ForceQuit() {
	this.station.ShutDown()
	this.reset()
}

func (this *Node) Delete(key string) bool {
	return true
}
