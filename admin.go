package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"runtime/debug"
	"strconv"
	"strings"
	"sync"
)

type AdminConnection struct {
	conn   net.Conn
	reader *bufio.Reader
	writer *bufio.Writer

	sync.RWMutex
}

func NewAdminConnection(conn net.Conn) (adminConn *AdminConnection) {
	adminConn = &AdminConnection{
		conn:   conn,
		reader: bufio.NewReader(conn),
		writer: bufio.NewWriter(conn),
	}

	log.Println("Admin connection from", conn.RemoteAddr().String())

	return adminConn
}

func (conn *AdminConnection) read() {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Recovered from a panic", r)
			debug.PrintStack()

			conn.writer.Flush()
			conn.conn.Close()
		}
	}()

	defer func() {
		conn.writer.Flush()
		conn.conn.Close()
	}()

	line, err := conn.reader.ReadString('\n')

	if err != nil {
		return
	}
	line = strings.TrimSpace(line)
	log.Println("Admin command", line)
	tokens := strings.Split(line, " ")

	// dispatch
	switch tokens[0] {
	case "ALERT":
		// ALERT <id> <predefinedhint> <message...>
		id, err := strconv.Atoi(tokens[1])
		if err != nil {
			conn.writer.WriteString("FAIL Bad Id")
			return
		}

		predef, err := strconv.Atoi(tokens[2])
		message := strings.Join(tokens[3:], " ")
		if id > 0 {
			client := clientCache.Get(int64(id))
			if client == nil {
				conn.writer.WriteString("FAIL Bad Client")
				return
			}

			client.SendBroadcastAlert(int32(predef), message)
		} else {
			SendBroadcastAlert(int32(predef), message)
		}

		conn.writer.WriteString("OK\n")
	case "FF":
		// FF <id> [message]
		id, err := strconv.Atoi(tokens[1])
		if err != nil {
			conn.writer.WriteString("FAIL Bad Id")
			return
		}
		client := clientCache.Get(int64(id))
		if client == nil {
			conn.writer.WriteString("FAIL Bad Client")
			return
		}

		opponent := clientCache.Get(client.PendingMatchmakingOpponentId)
		client.ForfeitMatchmadeMatch()

		if len(tokens) > 2 {
			message := strings.Join(tokens[2:], " ")
			client.SendBroadcastAlert(2, message)
			if opponent != nil {
				opponent.SendBroadcastAlert(3, message)
			}
		}

		conn.writer.WriteString("OK\n")
	case "MMEND":
		// MMEND <id>
		id, err := strconv.Atoi(tokens[1])
		if err != nil {
			conn.writer.WriteString("FAIL Bad Id")
			return
		}
		matchmaker.EndMatch(int64(id))
		conn.writer.WriteString("OK\n")
	case "MMENDCLIENT":
		// MMENDCLIENT <id> [message]
		id, err := strconv.Atoi(tokens[1])
		if err != nil {
			conn.writer.WriteString("FAIL Bad Id")
			return
		}
		client := clientCache.Get(int64(id))
		if client == nil {
			conn.writer.WriteString("FAIL Bad Client")
			return
		}
		if client.PendingMatchmakingId > 0 {
			matchmaker.EndMatch(client.PendingMatchmakingId)

			if len(tokens) > 2 {
				message := strings.Join(tokens[2:], " ")
				client.SendBroadcastAlert(4, message)
			}
		}
		conn.writer.WriteString("OK\n")
	case "POOL":
		loadMaps()
		conn.writer.WriteString("OK\n")

	case "ACTIVE":
		for _, el := range clientConnections {
			if el != nil {
				var id int64 = 0
				var username string = "N/A"
				if el.client != nil {
					id = el.client.Id
					username = el.client.Username
				}

				conn.writer.WriteString(fmt.Sprintf("CLIENT %d %s %s\n", id, el.conn.RemoteAddr().String(), username))
			}
		}

		conn.writer.WriteString("OK\n")

	case "REFRESH":
		// REFRESH <id>
		// Refresh a client from the db
		id, err := strconv.Atoi(tokens[1])
		if err != nil {
			conn.writer.WriteString("FAIL Bad Id")
			return
		}
		client := clientCache.Get(int64(id))
		if client == nil {
			conn.writer.WriteString("FAIL Bad Client")
			return
		}

		client.Refresh()
		client.BroadcastStatsMessage()

		conn.writer.WriteString("OK\n")

	case "DC":
		// DC <id>
		// Disconnect a client

		id, err := strconv.Atoi(tokens[1])
		if err != nil {
			conn.writer.WriteString("FAIL Bad Id")
			return
		}

		for x := range clientConnections {
			if clientConnections[x].client != nil && clientConnections[x].client.Id == int64(id) {
				clientConnections[x].Close()
			}
		}

		conn.writer.WriteString("OK\n")
	}
}
