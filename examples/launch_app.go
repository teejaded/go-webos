package main

import (
	"crypto/tls"
	"log"
	"net"
	"time"

	"github.com/gorilla/websocket"

	webos "github.com/teejaded/go-webos"
)

func main() {
	dialer := websocket.Dialer{
		HandshakeTimeout: 10 * time.Second,
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
		NetDial: (&net.Dialer{
			Timeout: time.Second * 5,
		}).Dial,
	}

	tv, err := webos.NewTV(&dialer, "192.168.4.82")
	if err != nil {
		log.Fatalf("could not dial: %v", err)
	}
	defer tv.Close()

	go tv.MessageHandler()

	if err = tv.AuthoriseClientKey("*****"); err != nil {
		log.Fatalf("could not authoise using client key: %v", err)
	}

	// tv.LaunchApp("netflix")
	msg, err := tv.CurrentChannel()
	log.Printf("%v | %v", msg, err)
}
