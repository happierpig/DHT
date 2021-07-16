package chord

import (
	"errors"
	log "github.com/sirupsen/logrus"
	"net"
	"net/rpc"
	"time"
)

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
		log.Errorf("<RPC Init>fail to register in address : %s\n", address)
		return err1
	}
	// 指定rpc模式为TCP模式，地址为address，开始监听
	this.lis, err1 = net.Listen("tcp", address)
	if err1 != nil {
		log.Errorf("<RPC Init>fail to listen in address : %s\n", address)
		return err1
	}
	log.Infof("<RPC Init> service start success in %s\n", address)
	go this.serv.Accept(this.lis)
	return nil
}

func GetClient(address string) (*rpc.Client, error) {
	if address == "" {
		log.Warningln("<GetClient> IP address is nil")
		return nil, errors.New("<GetClient> IP address is nil")
	}
	var client *rpc.Client
	var err error
	ch := make(chan error)
	for i := 0; i < 4; i++ {
		go func() {
			client, err = rpc.Dial("tcp", address)
			ch <- err
		}()
		select {
		case <-ch:
			if err == nil {
				return client, nil
			}
			// try many times to avoid that
		case <-time.After(waitTime):
			err = errors.New("Timeout")
			log.Warnln("<GetClient> Timeout to ", address)
		}
	}
	return nil, err
}

func CheckOnline(address string) bool {
	client, err := GetClient(address)
	if err != nil {
		log.Infoln("<CheckOnline> Ping Fail in ", address, "error: ", err)
		return false
	}
	if client != nil {
		defer client.Close()
	} else {
		return false
	}
	log.Infoln("<CheckOnline> Ping Online in ", address)
	return true
}

func RemoteCall(targetNode string, funcClass string, input interface{}, result interface{}) error {
	if targetNode == "" {
		log.Warningln("<RemoteCall> IP address is nil")
		return errors.New("Null address for RemoteCall")
	}
	client, err := GetClient(targetNode)
	if err != nil {
		log.Warningln("<RemoteCall> Fail to dial in ", targetNode, " and error is ", err)
		return err
	}
	if client != nil {
		defer client.Close()
	}
	err2 := client.Call(funcClass, input, result)
	if err2 == nil {
		log.Infoln("<RemoteCall> in ", targetNode, " with ", funcClass, " success!")
		return nil
	} else {
		log.Errorln("<RemoteCall> in ", targetNode, " with ", funcClass, " fail!")
		return err2
	}
}

func (this *network) ShutDown() error {
	err := this.lis.Close()
	if err != nil {
		log.Errorln("<ShutDown> Fail to close the network in ", this.nodePtr.node.address)
		return err
	}
	log.Infoln("<ShutDown> -", this.nodePtr.node.address, "- network close successfully :)")
	return nil
}
