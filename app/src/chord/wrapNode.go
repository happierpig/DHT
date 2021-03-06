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

func (this *WrapNode) CheckPredecessor(_ int, _ *string) error {
	return this.node.check_predecessor()
}

func (this *WrapNode) Stabilize(_ int, _ *string) error {
	this.node.stabilize()
	return nil
}

func (this *WrapNode) StoreData(dataPair Pair, _ *string) error {
	return this.node.store_data(dataPair)
}

func (this *WrapNode) GetData(key string, value *string) error {
	return this.node.get_data(key, value)
}

func (this *WrapNode) DeleteData(key string, _ *string) error {
	return this.node.delete_data(key)
}

func (this *WrapNode) HereditaryData(predeAddr string, dataSet *map[string]string) error {
	return this.node.hereditary_data(predeAddr, dataSet)
}

func (this *WrapNode) InheritData(dataSet *map[string]string, _ *string) error { // bug !!: input is input isn't output no-use
	return this.node.inherit_data(dataSet)
}

func (this *WrapNode) StoreBackup(dataPair Pair, _ *string) error {
	return this.node.store_backup(dataPair)
}

func (this *WrapNode) DeleteBackup(key string, _ *string) error {
	return this.node.delete_backup(key)
}

func (this *WrapNode) AddBackup(dataSet map[string]string, _ *string) error {
	return this.node.add_backup(dataSet)
}

func (this *WrapNode) SubBackup(dataSet map[string]string, _ *string) error {
	return this.node.sub_backup(dataSet)
}

func (this *WrapNode) GenerateBackup(_ *string, dataSet *map[string]string) error {
	return this.node.generate_backup(dataSet)
}
