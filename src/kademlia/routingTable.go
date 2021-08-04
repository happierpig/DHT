package kademlia

import (
	"container/list"
	"time"
)

func (this *RoutingTable) InitRoutingTable(nodeAddr Contact) {
	this.nodeAddr = nodeAddr
	this.rwLock.Lock()
	for i := 0; i < IDlength*8; i++ {
		this.buckets[i] = list.New()
		this.refreshTimeSet[i] = time.Now()
	}
	this.refreshIndex = 0
	this.rwLock.Unlock()
}

// Update  used when replier node is called and requester call successfully
func (this *RoutingTable) Update(contact *Contact) {
	//log.Infoln("<Update> Update ",contact.Address)
	this.rwLock.RLock()
	bucket := this.buckets[PrefixLen(Xor(this.nodeAddr.NodeID, contact.NodeID))]
	target := bucket.Front()
	target = nil
	for i := bucket.Front(); ; i = i.Next() {
		if i == nil {
			target = nil
			break
		}
		if i.Value.(*Contact).NodeID.Equals(contact.NodeID) {
			target = i
			break
		}
	}
	this.rwLock.RUnlock()
	this.rwLock.Lock()
	if target != nil {
		bucket.MoveToBack(target)
	} else {
		if bucket.Len() < K {
			bucket.PushBack(contact)
		} else {
			tmp := bucket.Front()
			if !pureCheckConn(tmp.Value.(*Contact).Address) {
				bucket.Remove(tmp)
				bucket.PushBack(contact)
			} else {
				bucket.MoveToBack(tmp)
			}
		}
	}
	this.rwLock.Unlock()
}

func (this *RoutingTable) FindClosest(targetID ID, count int) []ContactRecord {
	result := make([]ContactRecord, 0, count)
	index := PrefixLen(Xor(this.nodeAddr.NodeID, targetID))
	this.rwLock.RLock()
	if targetID == this.nodeAddr.NodeID {
		result = append(result, ContactRecord{Xor(targetID, targetID), NewContact(this.nodeAddr.Address)})
	}
	for i := this.buckets[index].Front(); i != nil && len(result) < count; i = i.Next() {
		contact := i.Value.(*Contact)
		result = append(result, ContactRecord{Xor(targetID, contact.NodeID), *contact})
	}
	for i := 1; (index-i >= 0 || index+i < IDlength*8) && len(result) < count; i++ {
		if index-i >= 0 {
			for j := this.buckets[index-i].Front(); j != nil && len(result) < count; j = j.Next() {
				contact := j.Value.(*Contact)
				result = append(result, ContactRecord{Xor(targetID, contact.NodeID), *contact})
			}
		}
		if index+i < IDlength*8 {
			for j := this.buckets[index+i].Front(); j != nil && len(result) < count; j = j.Next() {
				contact := j.Value.(*Contact)
				result = append(result, ContactRecord{Xor(targetID, contact.NodeID), *contact})
			}
		}
	}
	this.rwLock.RUnlock()
	return result
}
