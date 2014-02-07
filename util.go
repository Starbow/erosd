package main

import (
	"code.google.com/p/goprotobuf/proto"
	"errors"
	"log"
	"strconv"
	"strings"
)

//Take a "CMD TxID Len\n" input and split it up
func Unpack(data string) (event string, txid int, size int, err error) {
	data = strings.TrimRight(data, "\n")
	data = strings.TrimRight(data, "\r")
	result := strings.Split(data, " ")
	if len(result) != 3 {
		err = errors.New("Unable to extract event data.")
		return
	}
	txid, err = strconv.Atoi(result[1])
	if err != nil {
		return
	}

	size, err = strconv.Atoi(result[2])
	if err != nil {
		return
	}

	event = result[0]
	err = nil
	log.Println(data, event, size)
	return
}

//Data -> proto.Message
func Unmarshal(data []byte, message proto.Message) error {
	return proto.Unmarshal(data, message)
}

//message -> data
func Marshal(message proto.Message) (data []byte, err error) {
	return proto.Marshal(message)
}

//Broadcast a message to all active connections.
func broadcastMessage(command string, message proto.Message) {
	data, err := Marshal(message)

	if err != nil {
		panic(err)
	}
	for _, v := range clientConnections {

		go v.SendData(command, 0, data)
	}
}

//Broadcast a message to a specific client.
func (c *Client) Broadcast(command string, message proto.Message) {
	data, err := Marshal(message)

	if err != nil {
		panic(err)
	}
	for _, v := range clientConnections {

		if v.client.Id == c.Id {
			go v.SendData(command, 0, data)
		}
	}
}
