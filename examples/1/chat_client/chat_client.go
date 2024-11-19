package main

import (
	"fmt"
	"os"

	"github.com/Lei2050/lei-net/tcp"
)

var G_ChatConn *tcp.Client

func main() {
	ct, err := tcp.Dial("127.0.0.1:18888", &MsgHandle{},
		tcp.IdleTime(180000),
		tcp.ReadMaxSize(10240),
		tcp.WriteMaxSize(10240),
	)
	if err != nil {
		fmt.Println("Dial failed !")
		return
	}

	//注册连接断开时的回调函数
	ct.RegisterCloseCb(func() {
		fmt.Println("disconnected! client will exit.")
		os.Exit(0)
	})

	fmt.Println("CLIENT STARTED !")

	G_ChatConn = ct

	var input string
	for {
		input = ""
		fmt.Println("Please input:")
		fmt.Scanf("%s\n", &input)
		if input == "exit" {
			fmt.Println("done !")
			G_ChatConn.Close()
			os.Exit(0)
		} else if input == "reconnect" {
			G_ChatConn.Reconnect()
			continue
		}

		if len(input) <= 0 {
			continue
		}

		G_ChatConn.Write(input)
	}
}
