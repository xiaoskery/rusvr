package main

import (
	"github.com/rusvr/socket"
)

func main() {
	peer := socket.NewAcceptor().Start("127.0.0.1:8801")
	peer.SetName("server")
}
