package chord

import "math/big"

type WrapNode struct {
	node *Node
}

func (this *WrapNode) Find_Successor(hashValue *big.Int, succaddr *string) error {
	return nil
}
