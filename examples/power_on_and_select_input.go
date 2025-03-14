package main

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"time"

	wol "github.com/ghthor/gowol"
	"github.com/gorilla/websocket"

	webos "github.com/teejaded/go-webos"
)

var (
	gw      = "192.168.4.1"
	mac     = "e4:75:dc:36:82:62"
	ip      = "192.168.4.82"
	key     = "*****"
	inputID = "HDMI_2"
)

func canConnectToPort(ctx context.Context, host string, port int) bool {
	d := net.Dialer{Timeout: 100 * time.Millisecond}
	conn, err := d.DialContext(ctx, "tcp", net.JoinHostPort(host, fmt.Sprintf("%d", port)))
	if err == nil {
		defer conn.Close()
		return true
	}
	return false
}

func waitForPort(ctx context.Context, host string, port int) error {
	for !canConnectToPort(ctx, host, port) {
		select {
		case <-ctx.Done():
			return fmt.Errorf("timeout reached")
		case <-time.After(100 * time.Millisecond):
		}
	}
	return nil
}

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(60)*time.Second)
	defer cancel()

	log.Printf("Waiting for network\n")
	waitForPort(ctx, gw, 3001)

	log.Printf("Sending wake on lan to %s\n", mac)
	err := wol.MagicWake(mac, "255.255.255.255")
	if err != nil {
		log.Fatalf("wake on lan error: %v", err)
	}

	log.Printf("Waiting for %s:%d\n", ip, webos.Port)
	err = waitForPort(ctx, ip, webos.Port)
	if err != nil {
		log.Fatalf("error waiting for port: %v", err)
	}
	log.Printf("Connected to %s:%d\n", ip, webos.Port)

	dialer := websocket.Dialer{
		HandshakeTimeout: 10 * time.Second,
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
		NetDial: (&net.Dialer{
			Timeout: time.Second * 5,
		}).Dial,
	}

	tv, err := webos.NewTV(&dialer, ip)
	if err != nil {
		log.Fatalf("could not dial: %v", err)
	}
	defer tv.Close()

	go tv.MessageHandler()

	if err = tv.AuthoriseClientKey(key); err != nil {
		log.Fatalf("could not authorize using client key: %v", err)
	}

	time.Sleep(5 * time.Second)

	msg, err := tv.SwitchInput(inputID)
	if err != nil {
		log.Fatalf("error: %v", err)
	}
	j, _ := json.Marshal(msg)
	log.Printf("%v\n", string(j))
}
