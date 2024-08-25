package main

import (
	"DistributedFileSystems/p2p"
	"fmt"
)

func main() {
	tr := p2p.NewTcpTransport(":9000")
	if err := tr.ListenAndAccept(); err != nil {
		fmt.Println(err)
	}
	select {}
	fmt.Println("hello world")
}
