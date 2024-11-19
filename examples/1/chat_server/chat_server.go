package main

import (
	"fmt"

	"github.com/Lei2050/lei-net/tcp"
)

var G_ChatServer *tcp.Server

func main() {
	ser, err := tcp.NewServer(&MsgHandle{},
		tcp.Address("127.0.0.1:18888"),
		tcp.MaxConn(100),
		tcp.IdleTime(120000),
	)
	if err != nil {
		fmt.Println("NewServerTcp failed !")
		return
	}

	fmt.Println("SERVER STARTED !")

	G_ChatServer = ser
	G_ChatServer.Start()
}
