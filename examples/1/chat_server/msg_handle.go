package main

import (
	"fmt"

	"github.com/Lei2050/lei-net/api"
	"github.com/Lei2050/lei-net/tcp"
)

type MsgHandle struct{}

func (mh *MsgHandle) UnpackMsg(conn api.TcpConnectioner, data []byte) (int, error) {
	s := string(data)
	mh.Process(s)
	return len(data), nil
}

func (*MsgHandle) PackMsg(conn api.TcpConnectioner, msg interface{}) ([]byte, error) {
	s, _ := msg.(string)
	data := []byte(s)
	return data, nil
}

func (*MsgHandle) SetOption(opt tcp.Options) {}

func (MsgHandle) Process(msg string) {
	fmt.Printf("from client:%s\n", msg)
	G_ChatServer.Broadcast(msg)
}

func (MsgHandle) OnConnect(conn api.TcpConnectioner) {
	fmt.Printf("client %d connected.\n", conn.Id())
	conn.RegisterCloseCb(func() {
		fmt.Printf("client %d is disconnectd.\n", conn.Id())
	})
}

func (MsgHandle) HeartBeatMsg() interface{} { return nil }
