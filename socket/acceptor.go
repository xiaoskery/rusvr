package socket

import (
	"fmt"
	"net"
)

type socketAcceptor struct {
	*socketPeer

	listener net.Listener
}

func (acceptor *socketAcceptor) Start(address string) Peer {
	ln, err := net.Listen("tcp", address)
	acceptor.listener = ln
	if err != nil {
		fmt.Println(err.Error())
		return acceptor
	}

	// 接受线程
	acceptor.accept()

	return acceptor
}

func (acceptor *socketAcceptor) accept() {
	fmt.Println("accept")
	for {
		conn, err := acceptor.listener.Accept()
		if err != nil {
			fmt.Println(err.Error())
			break
		}

		// 处理连接进入独立线程, 防止accept无法响应
		go acceptor.onAccepted(conn)
	}
}

func (acceptor *socketAcceptor) onAccepted(conn net.Conn) {
	fmt.Println("onAccepted")
	ses := newSession(conn, acceptor)

	// 添加到管理器
	acceptor.Add(ses)

	// 断开后从管理器移除
	ses.OnClose = func() {
		acceptor.Remove(ses)
	}

	// 事件处理完成开始处理数据收发
	ses.run()
}

func (acceptor *socketAcceptor) Stop() {
	acceptor.listener.Close()
}

// NewAcceptor 创建acceptor
func NewAcceptor() Peer {
	self := &socketAcceptor{
		socketPeer: newSocketPeer(NewSessionManager()),
	}

	return self
}
