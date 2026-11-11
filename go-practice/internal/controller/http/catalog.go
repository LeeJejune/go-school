package http

import (
	"net/http"

	pb "github.com/LeeJejune/go-school/go-practice/gen/school/v1"
	"github.com/gin-gonic/gin"
)

func (h *Handler) ListResources(c *gin.Context) {
	response, err := h.clients.Catalog.ListResources(c.Request.Context(), &pb.ListResourcesRequest{})
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, mapCatalog(response.GetResources()))
}

func (h *Handler) GetDashboard(c *gin.Context) {
	snapshot, err := h.dashboard.Get(c.Request.Context(), c.Param("id"))
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, dashboardResponse{Catalog: mapCatalog(snapshot.Catalog.GetResources()), ReadingList: mapList(snapshot.ReadingList)})
}

func (h *Handler) UpsertResource(c *gin.Context) {
	var request upsertResourceRequest
	if !bindJSON(c, &request) {
		return
	}
	response, err := h.clients.Catalog.UpsertResource(c.Request.Context(), &pb.UpsertResourceRequest{Id: c.Param("id"), Title: request.Title, Url: request.URL})
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, struct {
		Resource *resourceResponse `json:"resource"`
	}{mapResource(response.GetResource())})
}
