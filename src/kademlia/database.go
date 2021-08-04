package kademlia

import (
	log "github.com/sirupsen/logrus"
	"time"
)

const (
	root       int = 0
	publisher  int = 1
	duplicater int = 2
	common     int = 3
)

func (this *database) init() {
	this.dataset = make(map[string]string)
	this.expireTime = make(map[string]time.Time)
	this.duplicateTime = make(map[string]time.Time)
	this.republicTime = make(map[string]time.Time)
	this.privilege = make(map[string]int)
}

func (this *database) store(request StoreRequest) {
	this.rwLock.Lock()
	defer this.rwLock.Unlock()
	if _, ok := this.dataset[request.Key]; !ok {
		requestPri := request.RequesterPri
		this.privilege[request.Key] = requestPri + 1
		this.dataset[request.Key] = request.Value
		if requestPri == root {
			this.republicTime[request.Key] = time.Now().Add(republicTimeInterval)
			return
		}
		if requestPri == publisher {
			this.duplicateTime[request.Key] = time.Now().Add(duplicateTimeInterval)
			this.expireTime[request.Key] = time.Now().Add(expireTimeInterval2)
			return
		}
		if requestPri == duplicater {
			this.expireTime[request.Key] = time.Now().Add(expireTimeInterval3)
			return
		}
	} else {
		originPri := this.privilege[request.Key]
		requestPri := request.RequesterPri
		if originPri == publisher || requestPri >= originPri {
			return
		}
		// duplicater->publisher || common->publisher || common->duplicater
		if requestPri == root {
			this.privilege[request.Key] = publisher
			this.dataset[request.Key] = request.Value
			this.republicTime[request.Key] = time.Now().Add(republicTimeInterval)
			delete(this.expireTime, request.Key)
			delete(this.duplicateTime, request.Key)
			return
		}
		if requestPri == publisher {
			this.privilege[request.Key] = duplicater
			this.dataset[request.Key] = request.Value
			this.expireTime[request.Key] = time.Now().Add(expireTimeInterval2)
			this.duplicateTime[request.Key] = time.Now().Add(duplicateTimeInterval)
			return
		}
		if requestPri == duplicater {
			this.privilege[request.Key] = common
			this.dataset[request.Key] = request.Value
			this.expireTime[request.Key] = time.Now().Add(expireTimeInterval3)
			return
		}
	}
}

func (this *database) expire() {
	tmp := make(map[string]bool)
	this.rwLock.RLock()
	for k, v := range this.expireTime {
		if !v.After(time.Now()) {
			tmp[k] = true
		}
	}
	this.rwLock.RUnlock()
	this.rwLock.Lock()
	for k, _ := range tmp {
		log.Infoln("<Database expire> Throw ", k)
		delete(this.dataset, k)
		delete(this.expireTime, k)
		delete(this.duplicateTime, k)
		delete(this.republicTime, k)
		delete(this.privilege, k)
	}
	this.rwLock.Unlock()
}

func (this *database) duplicate() (result map[string]string) {
	result = make(map[string]string)
	this.rwLock.RLock()
	for k, v := range this.duplicateTime {
		if !v.After(time.Now()) {
			result[k] = this.dataset[k]
		}
	}
	this.rwLock.RUnlock()
	this.rwLock.Lock()
	for k, _ := range result {
		this.duplicateTime[k] = time.Now().Add(duplicateTimeInterval)
	}
	this.rwLock.Unlock()
	return
}

func (this *database) republic() (result map[string]string) {
	result = make(map[string]string)
	this.rwLock.RLock()
	for k, v := range this.republicTime {
		if !v.After(time.Now()) {
			result[k] = this.dataset[k]
		}
	}
	this.rwLock.RUnlock()
	this.rwLock.Lock()
	for k, _ := range result {
		this.republicTime[k] = time.Now().Add(republicTimeInterval)
	}
	this.rwLock.Unlock()
	return
}

func (this *database) get(key string) (bool, string) {
	this.rwLock.Lock()
	defer this.rwLock.Unlock()
	if v, ok := this.dataset[key]; ok {
		if _, ok2 := this.expireTime[key]; ok2 {
			this.expireTime[key] = time.Now().Add(expireTimeInterval2)
		}
		return true, v
	} else {
		return false, ""
	}
}
