package main

import (
	"fmt"
	"net"
	"errors"
	"io"
	"encoding/json"
)

import cc "chaat/common"

//var connections map[net.Conn]struct{}
var connections map[string]net.Conn

func connectionHandler(conn net.Conn) {
	buf := make([]byte, 1024)
	length, err := conn.Read(buf)
	if err != nil {
		fmt.Println("Disconnected, hello error:", err);
		conn.Close()
		return 
	}
	var msg cc.Message
	err = json.Unmarshal(buf[:length], &msg)
	if err != nil {
		fmt.Println("Error unmarshaling:", err)
		fmt.Println(string(buf[:length]))
		conn.Close()
		return
	}
	if msg.MsgType != cc.MSG_HELLO {
		fmt.Println("First msg wasnt Hello, returning")
		fmt.Println(msg)
		conn.Close()
		return
	}
	uname := msg.From
	connections[uname] = conn
	for true {
		length, err = conn.Read(buf)
		if err != nil {
			delete(connections, uname)
			fmt.Println("Disconnected", uname);
			conn.Close()
			if !errors.Is(err, io.EOF) {
				fmt.Println("\nERROR:", err)
			}
			break
		}
		// fmt.Printf("Received %d bytes, \"%s\"\n", length, string(buf[:length]));

		err = json.Unmarshal(buf[:length], &msg)
		if err != nil {
			fmt.Println("Error unmarshaling:", err);
			fmt.Println(string(buf[:length]))
			return
		}

		sendconn := connections[msg.To]
		if sendconn == nil {
			fmt.Printf("User %s does not exist\n", msg.To);
		} else {
			sendconn.Write(buf[:length])
		}
	}

}

func main() {
	var laddr net.TCPAddr
	laddr.Port = 8080;
	connections = make(map[string]net.Conn)
	var listener, err = net.ListenTCP("tcp", &laddr);
	if (err != nil) {
		fmt.Println("ERROR LISTENING:", err);
		return
	}
	fmt.Println("Listening on address", listener.Addr().String());

	for true {
		var conn, err = listener.Accept()
		fmt.Printf("Accepted conn from addr %s\n", conn.RemoteAddr().String());
		if err != nil {
			fmt.Printf("ERROR: %s\n", err)
		} else {
			//connections[conn] = struct{}{}
			go connectionHandler(conn)
		}
	}
}
