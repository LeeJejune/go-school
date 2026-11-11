package http

import (
	"net/http"

	pb "github.com/LeeJejune/go-school/go-practice/gen/school/v1"
	"github.com/gin-gonic/gin"
)

func (h *Handler) CreateList(c *gin.Context) {
	var request createListRequest
	if !bindJSON(c, &request) {
		return
	}
	response, err := h.clients.Reading.CreateList(c.Request.Context(), &pb.CreateListRequest{
		RequestKey: request.RequestKey, Title: request.Title, ResourceIds: request.ResourceIDs,
	})
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, listResponse{ReadingList: mapList(response.GetReadingList())})
}

func (h *Handler) GetList(c *gin.Context) {
	response, err := h.clients.Reading.GetList(c.Request.Context(), &pb.GetListRequest{Id: c.Param("id")})
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, listResponse{ReadingList: mapList(response.GetReadingList())})
}

func (h *Handler) SetRead(c *gin.Context) {
	var request setReadRequest
	if !bindJSON(c, &request) {
		return
	}
	_, err := h.clients.Reading.SetRead(c.Request.Context(), &pb.SetReadRequest{
		ListId: c.Param("id"), ResourceId: c.Param("resourceID"), IsRead: request.IsRead,
	})
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, struct{}{})
}
