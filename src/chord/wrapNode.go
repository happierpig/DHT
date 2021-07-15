package chord

import "math/big"

// WrapNode Class is designed to register RPC service
type WrapNode struct {
	node *Node
}

func (this *WrapNode) FindSuccessor(target *big.Int, result *string) error {
	return this.node.find_successor(target, result)
}
