package main

import (
	"fmt"

	"github.com/Lei2050/lei-net/api"
	pkt "github.com/Lei2050/lei-net/packet/v2"
)

var _ pkt.PacketHandler = &MsgHandle{}

type MsgHandle struct{}

func (mh *MsgHandle) Process(conn api.TcpConnectioner, packet *pkt.Packet) {
	str := packet.ReadVarStrH()
	fmt.Printf("from server:%s\n", str)
	packet.Release()
}
