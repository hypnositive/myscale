package handler

import (
	"net/http"
	"strconv"

	"bff-service/internal/client"
	"github.com/gin-gonic/gin"
)

type VMHandler struct {
	client *client.VMClient
}

func NewVMHandler(client *client.VMClient) *VMHandler {
	return &VMHandler{client: client}
}

func (h *VMHandler) ListVMs(c *gin.Context) {
	resp, err := h.client.ListVMs(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return 
	}
	c.JSON(http.StatusOK, resp.Vms)
}

func (h *VMHandler) VMAction(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "geçersiz vm id"})
		return
	}

	resp, err := h.client.VMAction(c.Request.Context(), int32(id))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": resp.Success,
		"message": resp.Message,
		"task_id": resp.TaskUpid,
	})
}