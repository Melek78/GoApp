package ws

import (
	"context"
	"fmt"
	"log"
	"strconv"

	"github.com/abeme/go_sm_api/service"
	"github.com/redis/go-redis/v9"
)

// Hub holds connections and subscribes to Redis channels for cross-instance delivery
type Hub struct {
	rdb      *redis.Client
	groupSvc *service.GroupService
	// maps
	clients    map[string]map[*Client]bool // userID -> set of clients
	register   chan *Client
	unregister chan *Client
	broadcast  chan *Message
}

type Message struct {
	TargetUser string // if set, private
	Group      string // channel name like group:<id>
	Payload    []byte
}

func NewHub(rdb *redis.Client, groupSvc *service.GroupService) *Hub {
	h := &Hub{
		rdb:        rdb,
		groupSvc:   groupSvc,
		clients:    make(map[string]map[*Client]bool),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		broadcast:  make(chan *Message, 256),
	}
	go h.run()
	return h
}

func (h *Hub) run() {
	// subscribe to Redis pubsub for group and private channels pattern
	pubsub := h.rdb.PSubscribe(context.Background(), "group:*", "private:*")
	ch := pubsub.Channel()
	go func() {
		for msg := range ch {
			// incoming pubsub message -> broadcast to local clients
			// topic is msg.Channel, payload is msg.Payload
			m := &Message{Group: msg.Channel, Payload: []byte(msg.Payload)}
			h.broadcast <- m
		}
	}()

	for {
		select {
		case c := <-h.register:
			if _, ok := h.clients[c.userID]; !ok {
				h.clients[c.userID] = make(map[*Client]bool)
			}
			h.clients[c.userID][c] = true
			log.Printf("client registered: %s", c.userID)
		case c := <-h.unregister:
			if conns, ok := h.clients[c.userID]; ok {
				if _, exists := conns[c]; exists {
					delete(conns, c)
					close(c.send)
				}
				if len(conns) == 0 {
					delete(h.clients, c.userID)
				}
			}
		case m := <-h.broadcast:
			if m.TargetUser != "" {
				// send to specific user
				if conns, ok := h.clients[m.TargetUser]; ok {
					for c := range conns {
						select {
						case c.send <- m.Payload:
						default:
							close(c.send)
							delete(conns, c)
						}
					}
				}
			} else if m.Group != "" {
				// m.Group is channel name like "group:123"
				// extract ID
				var groupID uint64
				if parts := m.Group; parts != "" {
					// parts expected "group:<id>"
					if idx := len("group:"); len(parts) > idx {
						if id, err := strconv.ParseUint(parts[idx:], 10, 64); err == nil {
							groupID = id
						}
					}
				}
				if groupID == 0 {
					// fallback: broadcast to all
					for _, conns := range h.clients {
						for c := range conns {
							select {
							case c.send <- m.Payload:
							default:
								close(c.send)
								delete(conns, c)
							}
						}
					}
					continue
				}
				// lookup members and send only to them
				if h.groupSvc != nil {
					if members, err := h.groupSvc.GetMembers(uint(groupID)); err == nil {
						// create a set for fast lookup
						memberSet := make(map[string]bool, len(members))
						for _, id := range members {
							memberSet[id] = true
						}
						for userID, conns := range h.clients {
							if !memberSet[userID] {
								continue
							}
							for c := range conns {
								select {
								case c.send <- m.Payload:
								default:
									close(c.send)
									delete(conns, c)
								}
							}
						}
					}
				}
			}
		}
	}
}

func (h *Hub) RegisterClient(c *Client) {
	h.register <- c
}

func (h *Hub) UnregisterClient(c *Client) {
	h.unregister <- c
}

func (h *Hub) PublishGroup(ctx context.Context, channel string, payload string) error {
	return h.rdb.Publish(ctx, channel, payload).Err()
}

// SendToUser enqueues a payload for delivery to all active connections of a user.
func (h *Hub) SendToUser(userID string, payload []byte) {
	h.broadcast <- &Message{TargetUser: userID, Payload: payload}
}

// SendToGroup enqueues a payload locally for a group; it will be processed like a pubsub message.
func (h *Hub) SendToGroup(groupID uint, payload []byte) {
	ch := fmt.Sprintf("group:%d", groupID)
	h.broadcast <- &Message{Group: ch, Payload: payload}
}
