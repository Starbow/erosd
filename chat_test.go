package main

import (
	"testing"
	"github.com/stretchr/testify/assert"
	"log"
	"os"
	"fmt"
)

func createTestRoom(fixed bool) *ChatRoom {
	initChat()
	room, _ := createChatRoom("test", "test", "test", true, fixed)
	room.logger = log.New(os.Stdout, fmt.Sprintf("chat-%d:", room.id), log.Ldate|log.Ltime|log.Lshortfile)
	return room
}

func fakeClient(id int64) *ClientConnection {
	c := &ClientConnection{
		id: id,
		authenticated: false,
		chatRooms:     make(map[string]*ChatRoom),
		connType: CLIENT_CONNECTION_TYPE_WEBSOCKET,
	}
	return c
}

func TestAddingMemberToChatRoom(t *testing.T) {
	cr := createTestRoom(false)
	fc := fakeClient(1)

	added := cr.addMemberIfNew(fc)
	assert.True(t, added, "a new member should be added and return true")

	_, ok := cr.members[fc.id]
	assert.True(t, ok, "a member should be there after being added")

	addedTwice := cr.addMemberIfNew(fc)
	assert.False(t, addedTwice, "a new member should not be added twice and return false")

}

func TestJoiningRoom(t *testing.T) {
	cr := createTestRoom(false)
	fc := fakeClient(1)

	cr.ClientJoin(fc)

	_, ok := cr.members[fc.id]
	assert.True(t, ok, "a member should be there after joining")
}
