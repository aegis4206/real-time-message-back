package main

import (
	"crypto/tls"
	"log"
	"net"
	"net/http"
	"real-time-message/handler"

	"github.com/pion/stun"
)

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/", handler.InitWs)
	certFile := "cert.pem"
	keyFile := "key.pem"
	server := &http.Server{
		Addr:    ":8005",
		Handler: mux,
		TLSConfig: &tls.Config{
			MinVersion: tls.VersionTLS12,
		},
	}
	err := server.ListenAndServeTLS(certFile, keyFile)
	if err != nil {
		log.Fatalln(err)
	}
	// go stunServer()
}

func stunServer() {
	addr := net.UDPAddr{
		IP:   net.ParseIP("0.0.0.0"),
		Port: 3478,
	}

	conn, err := net.ListenUDP("udp", &addr)
	if err != nil {
		log.Fatalf("Failed to start STUN server: %v", err)
	}
	defer conn.Close()

	log.Printf("Starting STUN server on %s", addr.String())

	for {
		buf := make([]byte, 1500)
		n, clientAddr, err := conn.ReadFrom(buf)
		if err != nil {
			log.Printf("讀取失敗: %v", err)
			continue
		}

		message := new(stun.Message)
		message.Raw = buf[:n]
		if err := message.Decode(); err != nil {
			log.Printf("失敗: %v", err)
			continue
		}

		go handler.StunHandler(conn, clientAddr, message)
	}
}
