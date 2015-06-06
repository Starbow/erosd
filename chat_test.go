package main

import (
	"testing"
	"github.com/stretchr/testify/assert"
)

func createTestRoom(fixed bool) *ChatRoom {
	initChat()
	_, room := NewChatRoom("test", "test", true, fixed)
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

	addedTwice := cr.addMemberIfNew(fc)
	assert.False(t, addedTwice, "a new member should not be added twice and return false")

}