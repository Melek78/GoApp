package controller

import (
	"net/http"
	"strconv"

	"github.com/abeme/go_sm_api/service"
	"github.com/gin-gonic/gin"
)

type CreateGroupRequest struct {
	Name string `json:"name" binding:"required"`
}

type GroupController struct {
	svc *service.GroupService
}

func NewGroupController(svc *service.GroupService) *GroupController {
	return &GroupController{svc: svc}
}

func (g *GroupController) Create(c *gin.Context) {
	var req CreateGroupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	userID, _ := c.Get("user_id")
	uidStr, _ := userID.(string)
	grp, err := g.svc.CreateGroup(req.Name, uidStr)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"id": grp.ID, "name": grp.Name})
}

func (g *GroupController) Join(c *gin.Context) {
	gid := c.Param("id")
	id64, err := strconv.ParseUint(gid, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid group id"})
		return
	}
	userID, _ := c.Get("user_id")
	uidStr, _ := userID.(string)
	if err := g.svc.JoinGroup(uint(id64), uidStr); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"joined": true})
}
