package handler

import (
	"log"
	"net"

	"github.com/pion/stun"
)

func StunHandler(conn net.PacketConn, addr net.Addr, message *stun.Message) {
	response := stun.MustBuild(
		stun.TransactionID,
		stun.BindingSuccess,
		stun.XORMappedAddress{
			IP:   addr.(*net.UDPAddr).IP,
			Port: addr.(*net.UDPAddr).Port,
		},
	)
	log.Println("Stun:", response)

	response.Encode()

	if _, err := conn.WriteTo(response.Raw, addr); err != nil {
		log.Println("失敗", err)
	}
}
