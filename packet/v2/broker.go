package packet

import (
	"fmt"
	"io"

	leiio "github.com/Lei2050/lei-utils/io"

	api "github.com/Lei2050/lei-net/api"
	tcp "github.com/Lei2050/lei-net/tcp/v2"
)

var _ tcp.Protocoler = &Broker{}

type Broker struct {
	packerHandler PacketHandler
}

func NewBroker(packerHandler PacketHandler) *Broker {
	return &Broker{
		packerHandler: packerHandler,
	}
}

func (b *Broker) Read(conn api.TcpConnectioner, reader io.Reader) error {
	var uint32Buffer [4]byte
	var err error

	// 先读长度
	err = leiio.ReadFull(reader, uint32Buffer[:])
	if err != nil {
		return err
	}

	payloadSize := packetEndian.Uint32(uint32Buffer[:])
	if payloadSize > MaxPayloadLength {
		return ErrPayloadTooLarge
	}

	packet := NewPacket()
	payload := packet.extendPayload(int(payloadSize))
	err = leiio.ReadFull(reader, payload)
	if err != nil {
		return err
	}

	packet.SetPayloadLen(payloadSize)

	b.packerHandler.Process(conn, packet)

	return nil
}

func (b *Broker) Write(writer io.Writer, msg any) error {
	packet, ok := msg.(*Packet)
	if !ok {
		return fmt.Errorf("msg is not a Packet")
	}

	pdata := packet.data()
	err := leiio.WriteFull(writer, pdata)
	packet.Release()
	return err
}

func (b *Broker) OnConnect(api.TcpConnectioner) {}

func (b *Broker) HeartBeatMsg() any {
	return nil
}

func (b *Broker) SetOption(tcp.Options) {}
