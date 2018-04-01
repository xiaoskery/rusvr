package socket

import "net"

// Peer 端, Connector或Acceptor
type Peer interface {
	// 开启
	Start(address string) Peer

	// 关闭
	Stop()

	// 基础信息
	PeerProfile

	// 定制处理链
	//HandlerChainManager

	// 会话管理
	SessionAccessor
}

// Peer间的共享数据
type socketPeer struct {
	// 会话管理器
	SessionManager

	// 共享配置
	*PeerProfileImplement

	// 处理链管理
	//*cellnet.HandlerChainManagerImplement

	// socket配置
	*socketOptions

	// 停止过程同步
	stopping chan bool
}

func (self *socketPeer) waitStopFinished() {
	// 如果正在停止时, 等待停止完成
	if self.stopping != nil {
		<-self.stopping
		self.stopping = nil
	}
}

func (self *socketPeer) isStopping() bool {
	return self.stopping != nil
}

func (self *socketPeer) startStopping() {
	self.stopping = make(chan bool)
}

func (self *socketPeer) endStopping() {
	select {
	case self.stopping <- true:

	default:
		self.stopping = nil
	}
}

func newSocketPeer(sm SessionManager) *socketPeer {
	self := &socketPeer{
		SessionManager:       sm,
		socketOptions:        newSocketOptions(),
		PeerProfileImplement: NewPeerProfile(),
	}

	return self
}

func errToResult(err error) Result {

	if err == nil {
		return Result_OK
	}

	switch n := err.(type) {
	case net.Error:
		if n.Timeout() {
			return Result_SocketTimeout
		}
	}

	return Result_SocketError
}
