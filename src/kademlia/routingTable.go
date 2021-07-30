package kademlia

import (
	"container/list"
	log "github.com/sirupsen/logrus"
)

func (this *RoutingTable) InitRoutingTable(nodeID ID) {
	this.nodeID = nodeID
	this.rwLock.Lock()
	for i := 0; i < IDlength*8; i++ {
		this.buckets[i] = list.New()
	}
	this.rwLock.Unlock()
}

// Update  used when replier node is called and requester call successfully
func (this *RoutingTable) Update(contact *Contact) {
	log.Infoln("<Update> Update ", contact.address, " in ", this.nodeID)
	this.rwLock.RLock()
	bucket := this.buckets[PrefixLen(Xor(this.nodeID, contact.nodeID))]
	target := bucket.Front()
	target = nil
	for i := bucket.Front(); ; i = i.Next() {
		if i == nil {
			target = nil
			break
		}
		if i.Value.(*Contact).nodeID.Equals(contact.nodeID) {
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
			if !pureCheckConn(tmp.Value.(*Contact).address) {
				bucket.Remove(tmp)
				bucket.PushBack(contact)
			}
		}
	}
	this.rwLock.Unlock()
}

func (this *RoutingTable) FindClosest(targetID ID, count int) []ContactRecord {
	result := make([]ContactRecord, 0, count)
	index := PrefixLen(Xor(this.nodeID, targetID))
	this.rwLock.RLock()
	for i := this.buckets[index].Front(); i != nil && len(result) < count; i = i.Next() {
		contact := i.Value.(*Contact)
		result = append(result, ContactRecord{Xor(targetID, contact.nodeID), *contact})
	}
	for i := 1; (index-i >= 0 || index+i < IDlength*8) && len(result) < count; i++ {
		if index-i >= 0 {
			for j := this.buckets[index-i].Front(); j != nil && len(result) < count; j = j.Next() {
				contact := j.Value.(*Contact)
				result = append(result, ContactRecord{Xor(targetID, contact.nodeID), *contact})
			}
		}
		if index+i < IDlength*8 {
			for j := this.buckets[index+i].Front(); j != nil && len(result) < count; j = j.Next() {
				contact := j.Value.(*Contact)
				result = append(result, ContactRecord{Xor(targetID, contact.nodeID), *contact})
			}
		}
	}
	this.rwLock.RUnlock()
	SliceSort(&result)
	return result
}
