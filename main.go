package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
)

type client chan<- string // an outgoing message channel
var (
	messages = make(chan string) // all incoming client messages
	entering = make(chan client) // entering clients
	leaving  = make(chan client) // leaving clients
)

func main() {
	listener, err := net.Listen("tcp", "localhost:8000")
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Starting chat room")
	go broadcaster()
	for {
		conn, err := listener.Accept()

		if err != nil {
			log.Print(err)
			continue
		}

		go handleConn(conn)
	}
}

func handleConn(conn net.Conn) {
	ch := make(chan string)
	go clientWriter(conn, ch)

	fmt.Fprintln(conn, "Who are you?")
	fmt.Fprint(conn, "> ")
	nameInput := bufio.NewScanner(conn)
	nameInput.Scan()
	name := nameInput.Text()

	entering <- ch
	messages <- fmt.Sprintf("%v has joined the chat", name)

	input := bufio.NewScanner(conn)
	for input.Scan() {
		messages <- fmt.Sprintf("%v: %v", name, input.Text())
	}

	leaving <- ch
	messages <- fmt.Sprintf("%v has left the chat", name)
	conn.Close()
}

func broadcaster() {
	clients := make(map[client]bool) // all connected clients
	for {
		select {
		case msg := <-messages:
			for cli := range clients {
				cli <- msg
			}
		case cli := <-entering:
			clients[cli] = true

		case cli := <-leaving:
			delete(clients, cli)
			close(cli)
		}
	}
}

func clientWriter(conn net.Conn, ch <-chan string) {
	for msg := range ch {
		fmt.Fprintln(conn, msg)
	}
}
