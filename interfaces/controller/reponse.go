package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func ResponseOK(c *gin.Context, data interface{}) {
	response(c, http.StatusOK, data)
}

func ResponseCreated(c *gin.Context, data interface{}) {
	response(c, http.StatusCreated, data)
}

func ResponseInternalServerError(c *gin.Context, msg string) {
	responseError(c, http.StatusInternalServerError, msg)
}

func ResponseBadRequest(c *gin.Context, msg string) {
	responseError(c, http.StatusBadRequest, msg)
}

func response(c *gin.Context, code int, data interface{}) {
	if data == nil {
		c.String(http.StatusOK, "")
		return
	}
	resp := APIResponse{Data: data}
	c.JSON(code, resp)
}

func responseError(c *gin.Context, code int, msg string) {
	resp := APIErrorResponse{Message: msg}
	c.JSON(code, resp)
}
