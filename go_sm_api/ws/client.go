package ws

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/abeme/go_sm_api/service"
	"github.com/gorilla/websocket"
)

const (
	writeWait = 10 * time.Second
	pongWait  = 60 * time.Second
)

type Client struct {
	hub    *Hub
	conn   *websocket.Conn
	send   chan []byte
	userID string
	pmSvc  service.PrivateMessageService
}

func (c *Client) readPump() {
	defer func() {
		c.hub.UnregisterClient(c)
		_ = c.conn.Close()
	}()
	c.conn.SetReadLimit(512)
	_ = c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error { _ = c.conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })
	for {
		_, raw, err := c.conn.ReadMessage()
		if err != nil {
			log.Printf("read error: %v", err)
			break
		}
		// Envelope structure
		var env struct {
			Type   string `json:"type"`
			To     string `json:"to"`
			Body   string `json:"body"`
			TempID string `json:"tempId"`
		}
		if err := json.Unmarshal(raw, &env); err != nil {
			c.send <- []byte(`{"type":"error","error":"invalid_json"}`)
			continue
		}
		switch env.Type {
		case "private":
			if env.To == "" || env.Body == "" {
				c.send <- []byte(`{"type":"error","error":"missing_fields"}`)
				continue
			}
			pm, err := c.pmSvc.Send(c.userID, env.To, env.Body)
			if err != nil {
				c.send <- []byte(`{"type":"error","error":"send_failed"}`)
				continue
			}
			ts := pm.CreatedAt.Unix()
			// Ack payload
			ack := map[string]interface{}{
				"type":   "private_ack",
				"tempId": env.TempID,
				"id":     pm.ID,
				"from":   pm.SenderID,
				"to":     pm.RecipientID,
				"body":   pm.Body,
				"ts":     ts,
			}
			ackBytes, _ := json.Marshal(ack)
			c.send <- ackBytes
			// Event payload broadcast to both parties
			evt := map[string]interface{}{
				"type": "private",
				"id":   pm.ID,
				"from": pm.SenderID,
				"to":   pm.RecipientID,
				"body": pm.Body,
				"ts":   ts,
				"read": false,
			}
			evtBytes, _ := json.Marshal(evt)
			// deliver to recipient
			c.hub.SendToUser(pm.RecipientID, evtBytes)
			// echo event to sender (in addition to ack)
			c.send <- evtBytes
		default:
			// Unknown type
			c.send <- []byte(`{"type":"error","error":"unsupported_type"}`)
		}
	}
}

func (c *Client) writePump() {
	ticker := time.NewTicker((pongWait * 9) / 10)
	defer func() {
		ticker.Stop()
		_ = c.conn.Close()
	}()
	for {
		select {
		case message, ok := <-c.send:
			_ = c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// hub closed the channel
				_ = c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			if err := c.conn.WriteMessage(websocket.TextMessage, message); err != nil {
				return
			}
		case <-ticker.C:
			_ = c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func (c *Client) Serve(ctx context.Context) {
	go c.writePump()
	c.readPump()
}
