package daptcp

import (
	"bufio"
	"context"
	"fmt"
	"net"
	"os"
	"time"
)

type TCPConnection struct {
	IpAddr      string
	Port        string
	ServerAddr  string
	Connection  net.Conn
	Writer      *bufio.Writer
	Reader      *bufio.Reader
	DataChannel chan string
	handlers    []func(string)
}

func (tcp *TCPConnection) Write(m string) error {
	message := m + "\r\n"
	_, err := tcp.Writer.WriteString(message)
	if err != nil {
		return err
	}

	tcp.Writer.Flush()

	return nil
}

func (tcp *TCPConnection) Listen() {
	go tcp.listen()
	go tcp.dispatchMessages()
}

func (tcp *TCPConnection) OnMessage(handler func(string)) {
	tcp.handlers = append(tcp.handlers, handler)
}

func (tcp *TCPConnection) listen() {
	for {
		message, err := tcp.Reader.ReadString('\n')
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to read from server: %v\n", err)
			close(tcp.DataChannel) //? Close the channel on error to signal the message handler to stop			return // Exit if we encounter an error
		}
		tcp.DataChannel <- message
	}
}

func (tcp *TCPConnection) dispatchMessages() {
	for message := range tcp.DataChannel {
		for _, handler := range tcp.handlers {
			handler(message) //? Dispatch message to each handler
		}
	}
}

func NewTCPConn(ip string, port string) (TCPConnection, error) {
	tcp := TCPConnection{
		IpAddr:      ip,
		Port:        port,
		ServerAddr:  fmt.Sprintf("%s:%s", ip, port),
		DataChannel: make(chan string),
		handlers:    make([]func(string), 0),
	}

	dialer := net.Dialer{}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn, err := dialer.DialContext(ctx, "tcp", tcp.ServerAddr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to connect to server at %s: %v\n", tcp.ServerAddr, err)
		return tcp, err
	}

	tcp.Connection = conn
	tcp.Writer = bufio.NewWriter(tcp.Connection)
	tcp.Reader = bufio.NewReader(tcp.Connection)

	return tcp, nil
}
