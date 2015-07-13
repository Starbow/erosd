package main

import (
	"bufio"
	"bytes"
	"fmt"
	"github.com/stretchr/testify/assert"
	"log"
	"net"
	"os"
	"testing"
)

func init() {
	fmt.Fprintln(os.Stderr, "init()")
	dbType = "sqlite3"
	dbConnectionString = "erosd.sqlite3"
	initDb()
	initChat()
	initClientCaches()
}

func createTestRoom(fixed bool) *ChatRoom {
	room := createChatRoom("test", "test", "test", true, fixed)
	room.logger = log.New(os.Stdout, fmt.Sprintf("chat-%d:", room.id), log.Ldate|log.Ltime|log.Lshortfile)
	return room
}

func fakeClient(t *testing.T, id int64, write *bytes.Buffer) *ClientConnection {
	// Find a port that would be used as a local address.
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	conn, err := net.Dial("tcp", l.Addr().String())
	if err != nil {
		t.Fatal(err)
	}

	c := &ClientConnection{
		id:            id,
		authenticated: false,
		chatRooms:     make(map[string]*ChatRoom),
		conn:          conn,
		connType:      CLIENT_CONNECTION_TYPE_SOCKET,
		client:        &Client{Id: 1},
		logger:        log.New(os.Stdout, fmt.Sprintf("chat-%d:", id), log.Ldate|log.Ltime|log.Lshortfile),
		writer:        bufio.NewWriter(write),
	}
	return c
}


func TestAddingMemberToChatRoom(t *testing.T) {
	written := &bytes.Buffer{}
	cr := createTestRoom(false)
	fc := fakeClient(t, 1, written)
	defer fc.conn.(net.Conn).Close()

	added := cr.addMemberIfNew(fc)
	assert.True(t, added, "a new member should be added and return true")

	_, ok := cr.members[fc.id]
	assert.True(t, ok, "a member should be there after being added")

	addedTwice := cr.addMemberIfNew(fc)
	assert.False(t, addedTwice, "a new member should not be added twice and return false")

}

func TestJoiningRoom(t *testing.T) {
	written := &bytes.Buffer{}
	cr := createTestRoom(false)
	fc := fakeClient(t, 1, written)
	defer fc.conn.(net.Conn).Close()

	cr.ClientJoin(fc)

	_, ok := cr.members[fc.id]
	assert.True(t, ok, "a member should be there after joining")
}

func TestLeavingRoom(t *testing.T) {
	written := &bytes.Buffer{}
	cr := createTestRoom(false)
	fc := fakeClient(t, 1, written)
	defer fc.conn.(net.Conn).Close()

	cr.ClientJoin(fc)

	_, ok := cr.members[fc.id]
	assert.True(t, ok, "a member should be there after joining")
	log.Printf("member %v joined ok", fc.id)

	cr.ClientLeave(fc)

	_, ok = cr.members[fc.id]
	assert.False(t, ok, "a member not should be there after leaving")

	log.Printf("member %v left ok", fc.id)
}
