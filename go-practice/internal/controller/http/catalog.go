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
