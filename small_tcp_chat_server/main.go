package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"log"
	"net"
)

type Message struct {
	conn net.Conn
	message []byte
}

func startChannels(){
	for {
		select {
			case message := <- messages:
				broadcastMessage(message)
				case newClient := <- addClient:
					connections = append(connections, newClient)
					fmt.Println("New connection. Connections: ", len(connections))
				case deadClient := <- removeClient:
					removeConn(deadClient)
					fmt.Println("Disconnect. Connections: ", len(connections))
		}
	}
}

func broadcastMessage(m *Message){
	for _, conn := range connections {
		if m.conn != conn {
			_, err := conn.Write(m.message)
			if err != nil {
				log.Fatal(err)
			}
		}
	}
}

func removeConn(conn net.Conn){
	var i int
	for i = range connections {
		if connections[i] == conn {
			break;
		}
	}
	connections = append(connections[:i], connections[i+1:]...)
}

func handleRequest(conn net.Conn){
	reader := bufio.NewReader(conn)
	var buffer  bytes.Buffer
	for {
		for {
			bytesArray, isPrefix, err := reader.ReadLine()
			if err != nil {
				if err == io.EOF {
					removeClient <- conn
					conn.Close()
					return
				}
				log.Fatal(err)
			}
			buffer.Write(bytesArray)
			if !isPrefix {
				buffer.Write([]byte("\n"))
				break;
			}
		}
		m := Message {
			conn: conn,
			message: buffer.Bytes(),
		}
		buffer.Reset()
		messages <- &m
	}
}

var connections []net.Conn
var messages = make(chan *Message)
var addClient = make(chan net.Conn)
var removeClient = make(chan net.Conn)

func main(){
	address := ":3344"
	server, err := net.Listen("tcp", address)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Server is listening on: ", address)
	defer server.Close()

	go startChannels()
	for {
		conn, err := server.Accept()
		if err != nil {
			log.Fatal(err)
		}
		addClient <- conn
		go handleRequest(conn)
	}
}