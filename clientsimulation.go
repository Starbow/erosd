package main

import (
	"fmt"
	"log"
	"math/rand"
	"sync/atomic"
	"time"
)

var (
	simulationIdBase int64 = 0
)

type ClientSimulation struct {
	client *Client
	conn   *ClientConnection
}

const SIMULATOR_IS_CHATTY = true

func (cs *ClientSimulation) Run() {
	for {
		// Queue us
		matchmaker.register <- cs.conn

		// Wait for queue to be success
		<-matchmaker.callback

		el, ok := matchmaker.participants[cs.conn]

		if ok {
			select {
			case <-el.abort:
				// Aborted. Try more.
				continue
			case <-el.match:
				opponent := el.opponent
				elapsed := time.Since(el.enrollTime)
				cs.client.TotalQueueTime += elapsed.Seconds()

				// We have an opponent! Great success.
				// We don't want to run this logic twice, so the higher ID will run it
				if cs.client.Id > opponent.connection.client.Id {
					if SIMULATOR_IS_CHATTY {
						log.Printf("%s (%d) found %s (%d) after %d seconds", cs.client.Username, cs.client.LadderPoints, opponent.connection.client.Username, opponent.connection.client.LadderPoints, int64(elapsed.Seconds()))
					}
					victor := rand.Intn(10)
					var (
						winner *Client
						loser  *Client
					)
					if victor > 3 {
						winner = cs.client
						loser = opponent.connection.client
					} else {
						winner = opponent.connection.client
						loser = cs.client
					}

					winnerOldPoints := winner.LadderPoints
					loserOldPoints := loser.LadderPoints
					winner.Defeat(loser, BATTLENET_REGION_NA)

					if SIMULATOR_IS_CHATTY {
						log.Printf("%s defeated %s. Rating change %d -> %d and %d -> %d respectively", winner.Username, loser.Username, winnerOldPoints, winner.LadderPoints, loserOldPoints, loser.LadderPoints)
					}
				}
			}
		}
	}
}

func NewClientSimulation(points, radius int64) *ClientSimulation {
	id := atomic.AddInt64(&simulationIdBase, 1)
	client := NewClient()
	client.Id = id
	client.Username = fmt.Sprintf("User%d", id)
	client.LadderPoints = points
	client.LadderSearchRadius = radius

	return &ClientSimulation{
		client: client,
		conn: &ClientConnection{
			client: client,
		},
	}
}
