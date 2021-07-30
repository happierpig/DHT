package kademlia

func (this *WrapNode) GetClose(input FindNodeRequest, result *FindNodeReply) error {
	result.content = this.node.table.FindClosest(input.target, K)
	result.requester = input.requester
	result.replier = this.node.addr
	this.node.table.Update(&input.requester)
	return nil
}

func (this *WrapNode) Ping(requester Contact, result *string) error {
	this.node.Ping(requester)
	this.node.table.Update(&requester)
	return nil
}
