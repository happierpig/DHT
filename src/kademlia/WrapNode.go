package kademlia

import log "github.com/sirupsen/logrus"

func (this *WrapNode) GetClose(input FindNodeRequest, result *FindNodeReply) error {
	result.Content = this.node.table.FindClosest(input.Target, K)
	result.Requester = input.Requester
	result.Replier = this.node.addr
	this.node.table.Update(&input.Requester)
	return nil
}

func (this *WrapNode) Ping(requester Contact, result *string) error {
	this.node.Ping(requester.Address)
	this.node.table.Update(&requester)
	return nil
}

func (this *WrapNode) Store(input StoreRequest, result *string) error {
	this.node.data.store(input)
	log.Infoln("<store> Try to put <", input.Key, "> into ", this.node.addr.Address)
	this.node.table.Update(&input.Requester)
	return nil
}

func (this *WrapNode) FindValue(input FindValueRequest, result *FindValueReply) error {
	result.Content = this.node.table.FindClosest(Hash(input.Key), K)
	result.Requester = input.Requester
	result.Replier = this.node.addr
	log.Infoln("<FindValue> Try to get <", input.Key, "> in ", this.node.addr.Address)
	result.IsFind, result.Value = this.node.data.get(input.Key)
	this.node.table.Update(&input.Requester)
	return nil
}
