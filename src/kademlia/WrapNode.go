package kademlia

func (this *WrapNode) GetClose(input FindNodeRequest, result *FindNodeReply) error {
	result.Content = this.node.table.FindClosest(input.Target, K)
	result.Requester = input.Requester
	result.Replier = this.node.addr
	this.node.table.Update(&input.Requester)
	return nil
}

func (this *WrapNode) Ping(requester Contact, result *string) error {
	this.node.Ping(requester)
	this.node.table.Update(&requester)
	return nil
}
