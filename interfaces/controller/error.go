package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func ResponseInternalServerError(c *gin.Context, msg string) {
	c.JSON(http.StatusInternalServerError, gin.H{
		"code":    http.StatusInternalServerError,
		"message": msg,
	})
}

func ResponseBadRequest(c *gin.Context, msg string) {
	c.JSON(http.StatusBadRequest, gin.H{
		"code":    http.StatusBadRequest,
		"message": msg,
	})
}
