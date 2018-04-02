package socket

import (
	"bytes"
	"sync"
)

type HandlerChainManager interface {

	// 添加一条接收处理链
	AddChainRecv(recv *HandlerChain) int64

	// 移除接收处理链, 根据添加时的id
	RemoveChainRecv(id int64)

	// 接收处理链是否存在
	ChainRecvExists(id int64) bool

	// 获取当前的处理链(乱序)
	ChainListRecv() HandlerChainList

	// 设置发送处理链
	SetChainSend(chain *HandlerChain)

	// 获取当前发送处理链
	ChainSend() *HandlerChain

	// 读写链
	CreateChainWrite() *HandlerChain
	CreateChainRead() *HandlerChain

	// 设置读写链
	SetReadWriteChain(read, write func() *HandlerChain)
}

type HandlerChainManagerImplement struct {
	recvChainByID      map[int64]*HandlerChain
	recvChainGuard     sync.Mutex
	chainIDAcc         int64
	recvChainListDirty bool
	recvChainList      HandlerChainList

	sendChain      *HandlerChain
	sendChainGuard sync.RWMutex

	readChainCreator  func() *HandlerChain
	writeChainCreator func() *HandlerChain
	rwChainGuard      sync.RWMutex
}

func (self *HandlerChainManagerImplement) ChainRecvExists(id int64) bool {
	self.recvChainGuard.Lock()

	_, ok := self.recvChainByID[id]

	self.recvChainGuard.Unlock()

	return ok
}

func (self *HandlerChainManagerImplement) AddChainRecv(recv *HandlerChain) (autoID int64) {
	self.recvChainGuard.Lock()

	self.chainIDAcc++
	// autoID这里是流水生成，每次添加要变化
	// HandlerChain.id是固定id，用于调试用
	autoID = self.chainIDAcc
	self.recvChainByID[autoID] = recv
	self.recvChainListDirty = true

	self.recvChainGuard.Unlock()

	return
}

func (self *HandlerChainManagerImplement) RemoveChainRecv(id int64) {
	self.recvChainGuard.Lock()

	delete(self.recvChainByID, id)
	self.recvChainListDirty = true

	self.recvChainGuard.Unlock()
}

func (self *HandlerChainManagerImplement) SetChainSend(chain *HandlerChain) {
	self.sendChainGuard.Lock()
	self.sendChain = chain
	self.sendChainGuard.Unlock()
}

func (self *HandlerChainManagerImplement) CreateChainWrite() *HandlerChain {
	self.rwChainGuard.Lock()
	defer self.rwChainGuard.Unlock()
	return self.writeChainCreator()
}

func (self *HandlerChainManagerImplement) CreateChainRead() *HandlerChain {
	self.rwChainGuard.Lock()
	defer self.rwChainGuard.Unlock()
	return self.readChainCreator()
}

func (self *HandlerChainManagerImplement) SetReadWriteChain(read, write func() *HandlerChain) {
	self.rwChainGuard.Lock()

	if read != nil {
		self.readChainCreator = read
	}

	if write != nil {
		self.writeChainCreator = write
	}

	self.rwChainGuard.Unlock()
}

func (self *HandlerChainManagerImplement) ChainSend() *HandlerChain {
	self.sendChainGuard.Lock()
	defer self.sendChainGuard.Unlock()
	return self.sendChain
}

func (self *HandlerChainManagerImplement) ChainListRecv() HandlerChainList {
	self.recvChainGuard.Lock()
	defer self.recvChainGuard.Unlock()

	if self.recvChainListDirty {
		self.recvChainList = make(HandlerChainList, len(self.recvChainByID))
		index := 0
		for _, chain := range self.recvChainByID {

			self.recvChainList[index] = chain
			index++
		}

		self.recvChainListDirty = false
	}

	return self.recvChainList
}

func (self *HandlerChainManagerImplement) ChainString() string {
	var buff bytes.Buffer

	buff.WriteString("ChainRecv:\n")
	buff.WriteString(self.ChainListRecv().String())

	buff.WriteString("ChainSend:\n")
	buff.WriteString(self.ChainSend().String())

	return buff.String()
}

func NewHandlerChainManager() *HandlerChainManagerImplement {
	return &HandlerChainManagerImplement{
		recvChainByID: make(map[int64]*HandlerChain),
	}
}
