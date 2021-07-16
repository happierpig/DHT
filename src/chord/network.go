package chord

import (
	"errors"
	log "github.com/sirupsen/logrus"
	"net"
	"net/rpc"
)

var (
	localAddress string
)

func init() {
	//localAddress = GetLocalAddress()
	localAddress = "127.0.0.1"
}

type network struct {
	serv    *rpc.Server
	lis     net.Listener
	nodePtr *WrapNode
}

func (this *network) Init(address string, ptr *Node) error {
	//初始化一个server对象
	this.serv = rpc.NewServer()
	this.nodePtr = new(WrapNode)
	this.nodePtr.node = ptr
	//注册rpc服务
	err1 := this.serv.Register(this.nodePtr)
	if err1 != nil {
		log.Errorf("fail to register in address : %s\n", address)
		return err1
	}
	// 指定rpc模式为TCP模式，地址为address，开始监听
	this.lis, err1 = net.Listen("tcp", address)
	if err1 != nil {
		log.Errorf("fail to listen in address : %s\n", address)
		return err1
	}
	log.Infof("tcp rpc service start success in %s\n", address)
	go this.serv.Accept(this.lis)
	return nil
}

func CheckOnline(address string) bool {
	if address == "" {
		log.Warningln("IP address is nil")
		return false
	}
	client, err := rpc.Dial("tcp", address)
	if err != nil {
		log.Warningln("Fail to dial in ", address, " and error is ", err)
		return false
	}
	if client != nil {
		defer client.Close()
	} else {
		return false
	}
	log.Infoln("Ping Online in ", address)
	return true
}

func RemoteCall(targetNode string, funcClass string, input interface{}, result interface{}) error {
	if targetNode == "" {
		log.Warningln("RemoteCall : IP address is nil")
		return errors.New("Null address for RemoteCall")
	}
	client, err := rpc.Dial("tcp", targetNode)
	if err != nil {
		log.Warningln("RemoteCall : Fail to dial in ", targetNode, " and error is ", err)
		return err
	}
	if client != nil {
		defer client.Close()
	}
	err2 := client.Call(funcClass, input, result)
	if err2 == nil {
		log.Infoln("RemoteCall in ", targetNode, " with ", funcClass, " success!")
		return nil
	} else {
		log.Errorln("RemoteCall in ", targetNode, " with ", funcClass, " fail!")
		return err2
	}
}
