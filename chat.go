package main

import (
	"code.google.com/p/goprotobuf/proto"
	"errors"
	"github.com/Starbow/erosd/buffers"
	"log"
	"strings"
	"sync"
	"time"
)

var (
	chatRooms                map[string]*ChatRoom
	joinableChatRooms        map[string]*ChatRoom
	fixedChatRooms           []string
	maxChatRooms             int64 = 5
	ErrChatRoomAlreadyExists error = errors.New("The chat room name specified already exists.")
	ErrChatRoomReserved      error = errors.New("The chat room name is reserved.")
	ErrChatRoomNameTooShort  error = errors.New("The chat room name is too short")
	_                              = log.Ldate
)

func initChat() {
	chatRooms = make(map[string]*ChatRoom)
	joinableChatRooms = make(map[string]*ChatRoom)

	for x := range fixedChatRooms {
		room, err := NewChatRoom(fixedChatRooms[x], "", true, true)
		if err != nil {
			log.Println("Error creating channel", room, err)
		}
	}
}

type ChatRoom struct {
	members map[int64]*ClientConnection

	key      string
	name     string // Friendly name
	password string // Password
	joinable bool   // False if this is a server forced joining room (matchmaking)
	fixed    bool   // True if this is a server created room that never expires

	join    chan *ClientConnection
	leave   chan *ClientConnection
	message chan *protobufs.ChatRoomMessage
	abort   chan bool

	sync.RWMutex
}

func (cr *ChatRoom) run() {
	defer func() {
		// Handle closing the room
		delete(chatRooms, cr.key)
		delete(joinableChatRooms, cr.key)

		for x := range cr.members {
			delete(cr.members[x].chatRooms, cr.key)

		}
	}()
	timer := time.NewTicker(time.Second * 30)

	for {
		select {
		case <-timer.C:
			// Nobody has joined after a set time. Close the room.
			if len(cr.members) == 0 && !cr.fixed {
				return
			}
		case client := <-cr.join:
			cr.Lock()
			_, exists := cr.members[client.id]
			if exists {
				cr.Unlock()
				continue
			}

			cr.members[client.id] = client
			cr.Unlock()
			join := cr.ChatRoomUserMessage(client)
			cr.Broadcast("CHJ", &join)
			client.chatRooms[cr.key] = cr

		case client := <-cr.leave:
			cr.Lock()
			_, exists := cr.members[client.id]
			if !exists {
				cr.Unlock()
				continue
			}

			delete(cr.members, client.id)
			delete(client.chatRooms, cr.key)
			cr.Unlock()
			leave := cr.ChatRoomUserMessage(client)
			cr.Broadcast("CHL", &leave)

			if len(cr.members) == 0 && !cr.fixed {
				// Everyone has left. We're a non-fixed room. Leave.
				return
			}
		case msg := <-cr.message:
			cr.Broadcast("CHM", msg)
		case <-cr.abort:
			return
		}
	}
}

func (cr *ChatRoom) Broadcast(command string, message proto.Message) error {
	data, err := Marshal(message)
	if err != nil {
		return err
	}

	cr.RLock()
	defer cr.RUnlock()

	for x := range cr.members {
		go cr.members[x].SendData(command, 0, data)
	}

	return nil
}

func NewChatRoom(name, password string, joinable, fixed bool) (cr *ChatRoom, err error) {
	key := cleanChatRoomName(name)
	if len(key) < 3 {
		err = ErrChatRoomNameTooShort
		return
	}
	if joinable && key[:2] == "mm" {
		err = ErrChatRoomReserved
		return
	}

	_, ok := chatRooms[key]
	if ok {
		err = ErrChatRoomAlreadyExists
		return
	}

	cr = &ChatRoom{
		members:  make(map[int64]*ClientConnection),
		key:      key,
		name:     strings.TrimSpace(name),
		join:     make(chan *ClientConnection),
		leave:    make(chan *ClientConnection),
		message:  make(chan *protobufs.ChatRoomMessage, 256),
		abort:    make(chan bool, 1),
		joinable: joinable,
		fixed:    fixed,
		password: strings.TrimSpace(password),
	}

	go cr.run()

	chatRooms[key] = cr
	if joinable {
		joinableChatRooms[key] = cr
	}
	return
}

func (ch *ChatRoom) ChatRoomInfoMessage(detailed bool) protobufs.ChatRoomInfo {

	var (
		chat       protobufs.ChatRoomInfo
		users      int64 = int64(len(ch.members))
		passworded bool  = ch.password == ""
	)
	chat.Key = &ch.key
	chat.Name = &ch.name

	chat.Passworded = &passworded
	chat.Fixed = &ch.fixed
	chat.Joinable = &ch.joinable
	chat.Users = &users

	if detailed {
		// add participants
	}

	return chat

}

func (ch *ChatRoom) ChatRoomUserMessage(client *ClientConnection) protobufs.ChatRoomUser {

	var (
		chat protobufs.ChatRoomInfo = ch.ChatRoomInfoMessage(false)
		user protobufs.UserStats    = client.client.UserStatsMessage()
		join protobufs.ChatRoomUser
	)
	join.Room = &chat
	join.User = &user

	return join

}

func (ch *ChatRoom) ChatRoomMessageMessage(client *ClientConnection, message *protobufs.ChatMessage) protobufs.ChatRoomMessage {

	var (
		chat protobufs.ChatRoomInfo = ch.ChatRoomInfoMessage(false)
		user protobufs.UserStats    = client.client.UserStatsMessage()
		msg  protobufs.ChatRoomMessage
	)
	msg.Room = &chat
	msg.Sender = &user
	messageString := message.GetMessage()
	msg.Message = &messageString
	return msg

}

func cleanChatRoomName(name string) string {
	return strings.TrimSpace(strings.ToLower(name))
}
