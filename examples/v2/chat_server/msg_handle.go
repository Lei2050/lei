package main

import (
	"github.com/Lei2050/lei-net/api"
	pkt "github.com/Lei2050/lei-net/packet/v2"
)

var _ pkt.PacketHandler = &MsgHandle{}

type MsgHandle struct{}

func (mh *MsgHandle) Process(conn api.TcpConnectioner, packet *pkt.Packet) {
	//这种写法，packet可能不会返回对象池，
	//packet最终可能不一定被写入到链接中（不一定走得到Broker.Write，并调用Release）；
	//比如broadcast时是10个链接，而到真的写时某些链接断了。
	//这种情况不是高频，理论上可以接受。
	writeN := G_ChatServer.Broadcast(packet)
	//本身有个1，每个连接复用一次
	packet.RetainCount(int64(writeN) - 1)
}
