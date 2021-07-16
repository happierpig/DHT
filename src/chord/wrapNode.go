package chord

import "math/big"

// WrapNode Class is designed to register RPC service
type WrapNode struct {
	node *Node
}

func (this *WrapNode) FindSuccessor(target *big.Int, result *string) error {
	return this.node.find_successor(target, result)
}

func (this *WrapNode) SetSuccessorList(_ int, result *[successorListSize]string) error {
	return this.node.set_successor_list(result)
}

func (this *WrapNode) GetPredecessor(_ int, result *string) error {
	return this.node.get_predecessor(result)
}

func (this *WrapNode) Notify(instructor string, _ *string) error {
	return this.node.notify(instructor)
}
