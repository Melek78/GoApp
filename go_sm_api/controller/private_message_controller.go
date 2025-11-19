package controller

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/abeme/go_sm_api/service"
	"github.com/abeme/go_sm_api/ws"
	"github.com/gin-gonic/gin"
)

type PrivateMessageController struct {
	pmSvc   service.PrivateMessageService
	userSvc service.UserService
	hub     *ws.Hub
}

func NewPrivateMessageController(pmSvc service.PrivateMessageService, userSvc service.UserService, hub *ws.Hub) *PrivateMessageController {
	return &PrivateMessageController{pmSvc: pmSvc, userSvc: userSvc, hub: hub}
}

// ListConversation returns messages between authenticated user and other user.
func (p *PrivateMessageController) ListConversation(c *gin.Context) {
	otherUserID := c.Param("otherUserID")
	uidVal, _ := c.Get("user_id")
	userID, _ := uidVal.(string)
	if otherUserID == "" || userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing user ids"})
		return
	}
	limitStr := c.DefaultQuery("limit", "50")
	beforeStr := c.DefaultQuery("before", "0")
	limit, _ := strconv.Atoi(limitStr)
	beforeID64, _ := strconv.ParseUint(beforeStr, 10, 64)
	msgs, err := p.pmSvc.ListConversation(userID, otherUserID, limit, uint(beforeID64))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"messages": msgs})
}

type markReadRequest struct {
	User string `json:"user" binding:"required"`
	Ids  []uint `json:"ids" binding:"required"`
}

// MarkRead marks messages as read and emits websocket read receipts.
func (p *PrivateMessageController) MarkRead(c *gin.Context) {
	var req markReadRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	uidVal, _ := c.Get("user_id")
	recipientID, _ := uidVal.(string)
	if recipientID == "" || req.User == "" || len(req.Ids) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}
	updated, err := p.pmSvc.MarkRead(recipientID, req.User, req.Ids)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if updated > 0 {
		receipt := map[string]interface{}{
			"type": "private_read",
			"from": recipientID,
			"ids":  req.Ids,
			"ts":   time.Now().Unix(),
		}
		if b, err := json.Marshal(receipt); err == nil {
			p.hub.SendToUser(req.User, b)
		}
	}
	c.JSON(http.StatusOK, gin.H{"updated": updated})
}

//
