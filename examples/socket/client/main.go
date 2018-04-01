package main

import (
	"fmt"
	"sync"

	"github.com/rusvr/socket"
)

func main() {
	var waitGroup sync.WaitGroup

	peer := socket.NewConnector().Start("127.0.0.1:8801")
	peer.SetName("client")

	// conn, err := net.Dial("tcp", "127.0.0.1:8801")
	// // 连不上
	// if err != nil {
	// 	fmt.Println(err.Error())
	// 	return
	// }

	waitGroup.Add(1)
	waitGroup.Wait()

	fmt.Println("over")
}
