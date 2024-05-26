package main

import (
	"fmt"
	"log"
	"net"
)

type MessageType int

const (
	ClientConnected MessageType = iota + 1
	ClientDisconnected
	NewMessage
)

type Message struct {
	Type MessageType
	Conn net.Conn
	Text string
}

type Client struct {
	Conn net.Conn
}

func server(messages chan Message) {
	clients := map[string]*Client{}
	for {
		msg := <-messages
		switch msg.Type {
		case ClientConnected:
			addr := msg.Conn.RemoteAddr().(*net.TCPAddr)
			log.Printf("Client %s connected", addr.String())
			clients[msg.Conn.RemoteAddr().String()] = &Client{
				Conn: msg.Conn,
			}
		case ClientDisconnected:
			addr := msg.Conn.RemoteAddr().(*net.TCPAddr)
			log.Printf("Client %s disconnected", addr.String())
			delete(clients, addr.String())
		case NewMessage:
			authorAddr := msg.Conn.RemoteAddr().(*net.TCPAddr)
			author := clients[authorAddr.String()]
			if author != nil {
				log.Printf("Client %s sent message %s", authorAddr.String(), msg.Text)
				for _, client := range clients {
					if client.Conn.RemoteAddr().String() != authorAddr.String() {
						client.Conn.Write([]byte(msg.Text))
					}
				}
			} else {
				msg.Conn.Close()
			}
		}
	}
}

func client(conn net.Conn, messages chan Message) {
	buffer := make([]byte, 64)
	for {
		n, err := conn.Read(buffer)
		if err != nil {
			conn.Close()
			messages <- Message{
				Type: ClientDisconnected,
				Conn: conn,
			}
			return
		}

		text := string(buffer[0:n])
		messages <- Message{
			Type: NewMessage,
			Text: text,
			Conn: conn,
		}
	}
}

func main() {
	ln, err := net.Listen("tcp", ":8080")
	if err != nil {
		panic(fmt.Sprintf("could not open tcp :8080: %s", err))
	}

	messages := make(chan Message)
	go server(messages)
	for {
		conn, err := ln.Accept()
		if err != nil {
			panic(fmt.Sprintf("could not get accept connection: %s", err))
		}

		messages <- Message{
			Type: ClientConnected,
			Conn: conn,
		}

		go client(conn, messages)
	}
}
