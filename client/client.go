package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"encoding/json"
	"net"
	"os"
	"strings"
	"crypto/rsa"
	"crypto/rand"

	cc "chaat/common"
)

type State int

const (
	STATE_READY State = iota
	STATE_CONNECTED
	STATE_TALKING
)

type UserState struct {
	conn net.Conn
	state State
	name string
	done bool
	dest string
	key *rsa.PrivateKey
}

func (user *UserState) readyHandler(tokens []string, raw_inp string) error {
	// WARN: 
	raw_inp = ""
	switch tokens[0] {
	case "#connect":
		if len(tokens) < 3 {
			return fmt.Errorf("Not enough tokens");
		}
		var err error
		user.conn, err = net.Dial("tcp", tokens[2])
		if err != nil {
			return err
		}
		user.name = tokens[1]
		user.state = STATE_CONNECTED
		go receiveHandler(user.conn)
		msg := cc.Message{
			MsgType: cc.MSG_HELLO,
			From: user.name,
			To: "SERVER",
			Data: nil,
		}
		buf, err := json.Marshal(msg)
		if err != nil {
			fmt.Println("Error marshaling:", err);
			return err
		}
		_, err = user.conn.Write(buf)
		if err != nil {
			fmt.Println("Error sending:", err);
			return err
		}
		break
	case "#exit":
		user.done = true
		break
	default:
		return fmt.Errorf("Invalid tokens")
	}
	return nil
}

func (user *UserState) talkingHandler(tokens [] string, raw_inp string) error {
	switch tokens[0] {
	case "#done":
		user.dest = "SERVER";
		user.state = STATE_CONNECTED;
		break
	default:
		msg := cc.Message{MsgType: cc.MSG_TEXT, 
			From: user.name,
			To: user.dest,
			Data: []byte(raw_inp),
		}

		buf, err := json.Marshal(msg)
		if err != nil {
			fmt.Println("ERROR while marshalling message", err)
		}
		_, err = user.conn.Write(buf)
		if err != nil {
			return err
		}
		break
	}
	return nil
}

func (user *UserState) connectedHandler(tokens []string, raw_inp string) error {
	raw_inp = raw_inp
	switch tokens[0] {
	case "#open":
		user.dest = tokens[1];
		// TODO: Send CERT_REQ here
		user.state = STATE_TALKING;
	case "#disconnect":
		user.conn.Close()
		user.state = STATE_READY
		break
	case "#exit":
		user.conn.Close()
		user.done = true
		break
	}
	return nil
}

func (

func receiveHandler(conn net.Conn) {
	buf := make([]byte, 1024)
	for true {
		length, err := conn.Read(buf)
		if err != nil {
			if !errors.Is(err, io.EOF) {
				fmt.Println("\nERROR:", err)
			}
			break
		}
		var msg cc.Message
		err = json.Unmarshal(buf[:length], &msg)
		if err != nil {
			fmt.Println("Error unmarshalling:", err)
		} else {
			fmt.Printf("(%s->%s): %s\n", string(msg.From), string(msg.To), string(msg.Data));
		}
	}
}

func main() {
	var user UserState 
	user.state = STATE_READY
	user.done = false
	user.dest = "SERVER"
	var err error
	user.key, err = rsa.GenerateKey(rand.Reader, 256)
	if err != nil {
		fmt.Printf("ERROR gen key: %s\n", err)
		return
	}
	user.key.Precompute()
	scanner := bufio.NewReader(os.Stdin)
	for !user.done {
		byte_inp, _, _ := scanner.ReadLine()
		inp := string(byte_inp)
		tokens := strings.Split(inp, " ")
		var err error
		switch user.state {
		case STATE_READY:
			err = user.readyHandler(tokens, inp)
			break
		case STATE_CONNECTED:
			err = user.connectedHandler(tokens, inp)
			break
		case STATE_TALKING:
			err = user.talkingHandler(tokens, inp)
		}
		if err != nil {
			fmt.Printf("ERROR: %s\n", err)
			break
		}
	}
	fmt.Println("Goodbye!");
}
