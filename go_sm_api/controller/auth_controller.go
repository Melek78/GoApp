package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/abeme/go_sm_api/entity"
	"github.com/abeme/go_sm_api/service"
	"github.com/abeme/go_sm_api/utils"
)

type AuthController struct {
	svc service.UserService
}

func NewAuthController(svc service.UserService) *AuthController {
	return &AuthController{svc: svc}
}

func (a *AuthController) SignUp(c *gin.Context) {
	var req entity.SignUpRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	u, err := a.svc.CreateUser(req.Email, req.Password)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"id": u.ID, "email": u.Email})
}

func (a *AuthController) Login(c *gin.Context) {
	var req entity.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	u, err := a.svc.Authenticate(req.Email, req.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
	}
	token, err := utils.GenerateToken(u.ID, u.Email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate token"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"token": token})
}

// ...existing code...
