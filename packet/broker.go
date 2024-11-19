package packet

import (
	"fmt"

	"github.com/Lei2050/lei-net/tcp"

	"github.com/Lei2050/lei-net/api"
)

var _ tcp.Protocoler = &Broker{}

type Broker struct {
	c chan *Packet
}

func NewBroker() *Broker {
	return &Broker{
		c: make(chan *Packet, 1024),
	}
}

func (b *Broker) C() <-chan *Packet {
	return b.c
}

func (b *Broker) UnpackMsg(conn api.TcpConnectioner, data []byte) (int, error) {
	if len(data) < payloadLengthSize {
		//数据不够，等待下一次数据
		return 0, nil
	}
	packetLen := packetEndian.Uint32(data[0:payloadLengthSize])
	if len(data) < int(packetLen) {
		//数据不够，等待下一次数据
		return 0, nil
	}

	if packetLen > MaxPayloadLength {
		return 0, ErrPayloadTooLarge
	}

	packet := NewPacket()
	payload := packet.extendPayload(int(packetLen))
	// TODO ...
	//这种涉及有点不好，因为copy了两次
	copy(payload, data[payloadLengthSize:payloadLengthSize+packetLen])

	b.c <- packet

	return int(payloadLengthSize + packetLen), nil
}

func (b *Broker) PackMsg(conn api.TcpConnectioner, msg any) ([]byte, error) {
	packet, ok := msg.(*Packet)
	if !ok {
		return nil, fmt.Errorf("data is not a packet")
	}

	payloadLen := packet.GetPayloadLen()
	if payloadLen > MaxPayloadLength {
		return nil, ErrPayloadTooLarge
	}

	// TODO ...
	//这里设计不妥，data复制多次，并且强制用数据池
	data := make([]byte, payloadLengthSize+payloadLen)
	copy(data, packet.Data())

	packet.Release()

	return data, nil
}

func (b *Broker) OnConnect(api.TcpConnectioner) {}

func (b *Broker) HeartBeatMsg() any {
	return nil
}

func (b *Broker) SetOption(tcp.Options) {}
