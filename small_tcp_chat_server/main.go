package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"log"
	"net"
)

// Message структура сообщения
type Message struct {
	conn net.Conn
	message []byte
}

// selectChannels
// либо отправляет сообщение всем присоединившимся
// либо добавляет новое подключение в список подключений
// либо удаляет подключение из списка подключений
func selectChannels(){
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

// broadcaseMessage отправляет соощение всем подключившимся, кроме того, от кого это сообщение пришло
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

// removeConn удаляет подключение из списка подключений и закрывает его
func removeConn(conn net.Conn){
	var i int
	for i = range connections {
		if connections[i] == conn {
			break;
		}
	}
	connections = append(connections[:i], connections[i+1:]...)
	conn.Close()
}

// handleMessage получает байты из TCP подключения и отправляет их в канал messages
func handleMessage(conn net.Conn){
	reader := bufio.NewReader(conn)
	var buffer  bytes.Buffer
	for {
		for {
			bytesArray, isPrefix, err := reader.ReadLine()
			if err != nil {
				if err == io.EOF {
					removeClient <- conn
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

// connections массив всех подключений
var connections []net.Conn
// message канал сообщений, сюда приходят все сообщения и затем они рассылаются все подключившимся
var messages = make(chan *Message)
// addClient канал для добавления нового подключившегося в массив connections
var addClient = make(chan net.Conn)
// removeClient катал для удаления подключения (для избежания race conditions)
var removeClient = make(chan net.Conn)

func main(){
	address := ":3344"
	server, err := net.Listen("tcp", address)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Server is listening on: ", address)
	defer server.Close()

	go selectChannels()
	for {
		conn, err := server.Accept()
		if err != nil {
			log.Fatal(err)
		}
		addClient <- conn
		go handleMessage(conn)
	}
}