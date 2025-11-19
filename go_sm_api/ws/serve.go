package ws

import (
	"context"
	"log"
	"net/http"
	"strings"

	"github.com/abeme/go_sm_api/service"
	"github.com/abeme/go_sm_api/utils"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

// ServeWS upgrades the HTTP connection to a WebSocket, authenticates the user via JWT,
// registers the client with the hub, and starts pumps.
func ServeWS(h *Hub, pmSvc service.PrivateMessageService, c *gin.Context) {
	// get token from Authorization header
	auth := c.GetHeader("Authorization")
	if auth == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "missing Authorization header"})
		return
	}
	parts := strings.SplitN(auth, " ", 2)
	if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid Authorization header"})
		return
	}
	claims, err := utils.ValidateToken(parts[1])
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
		return
	}

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("upgrade failed: %v", err)
		return
	}

	client := &Client{
		hub:    h,
		conn:   conn,
		send:   make(chan []byte, 256),
		userID: claims.Subject,
		pmSvc:  pmSvc,
	}

	h.RegisterClient(client)
	go client.Serve(context.Background())
}
