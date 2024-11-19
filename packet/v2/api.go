package packet

import "github.com/Lei2050/lei-net/api"

type PacketHandler interface {
	Process(api.TcpConnectioner, *Packet)
}
