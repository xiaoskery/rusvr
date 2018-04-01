package socket

import (
	"fmt"
	"net"
	"time"
)

// Connector 连接器, 可由Peer转换
type Connector interface {
	// 连接后的Session
	DefaultSession() Session

	// 自动重连间隔, 0表示不重连, 默认不重连
	SetAutoReconnectSec(sec int)
}

type socketConnector struct {
	*socketPeer

	autoReconnectSec int // 重连间隔时间, 0为不重连

	tryConnTimes int // 尝试连接次数

	closeSignal chan bool

	defaultSes Session
}

func (self *socketConnector) SetAutoReconnectSec(sec int) {
	self.autoReconnectSec = sec
}

func (self *socketConnector) Start(address string) Peer {
	self.waitStopFinished()

	if self.IsRunning() {
		return self
	}

	go self.connect(address)

	return self
}

const reportConnectFailedLimitTimes = 3

func (self *socketConnector) connect(address string) {
	defer func() {
		fmt.Println("Dialx")
	}()
	self.SetRunning(true)
	self.SetAddress(address)

	for {
		self.tryConnTimes++

		fmt.Println("Dial")
		// 开始连接
		conn, err := net.Dial("tcp", address)
		fmt.Println("after dial")
		// 连不上
		if err != nil {
			if self.tryConnTimes <= reportConnectFailedLimitTimes {
				fmt.Println(err.Error())
			}

			if self.tryConnTimes == reportConnectFailedLimitTimes {
				fmt.Println(err.Error())
			}

			// 没重连就退出
			if self.autoReconnectSec == 0 {
				break
			}

			// 有重连就等待
			time.Sleep(time.Duration(self.autoReconnectSec) * time.Second)

			// 继续连接
			continue
		}

		// 创建Session
		ses := newSession(conn, self)

		self.defaultSes = ses
		self.tryConnTimes = 0
		self.Add(ses)

		// 内部断开回调
		ses.OnClose = func() {
			self.Remove(ses)
			self.closeSignal <- true
		}

		// 投递连接建立事件
		// extend.PostSystemEvent(ses, cellnet.Event_Connected, self.ChainListRecv(), cellnet.Result_OK)

		// 事件处理完成开始处理数据收发
		ses.run()

		if <-self.closeSignal {
			self.defaultSes = nil

			// 没重连就退出/主动退出
			if self.isStopping() || self.autoReconnectSec == 0 {
				break
			}

			// 有重连就等待
			time.Sleep(time.Duration(self.autoReconnectSec) * time.Second)

			// 继续连接
			continue
		}
	}

	self.SetRunning(false)

	self.endStopping()
}

func (self *socketConnector) Stop() {
	if !self.IsRunning() {
		return
	}

	if self.isStopping() {
		return
	}

	self.startStopping()

	if self.defaultSes != nil {
		self.defaultSes.Close()
	}

	// 等待线程结束
	self.waitStopFinished()
}

func (self *socketConnector) DefaultSession() Session {
	return self.defaultSes
}

func NewConnector() Peer {
	return NewConnectorBySessionManager(NewSessionManager())
}

func NewConnectorBySessionManager(sm SessionManager) Peer {
	self := &socketConnector{
		socketPeer:  newSocketPeer(sm),
		closeSignal: make(chan bool),
	}

	return self
}
