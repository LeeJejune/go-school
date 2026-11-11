package http

import pb "github.com/LeeJejune/go-school/go-practice/gen/school/v1"

// HTTP JSON DTOs are separate from the generated gRPC messages.
type createListRequest struct {
	RequestKey  string   `json:"requestKey" binding:"required"`
	Title       string   `json:"title" binding:"required"`
	ResourceIDs []string `json:"resourceIds" binding:"required"`
}

type upsertResourceRequest struct {
	Title string `json:"title" binding:"required"`
	URL   string `json:"url" binding:"required"`
}

type setReadRequest struct {
	IsRead bool `json:"isRead"`
}

type resourceResponse struct {
	ID    string `json:"id"`
	Title string `json:"title"`
	URL   string `json:"url"`
}

type itemResponse struct {
	Resource *resourceResponse `json:"resource"`
	IsRead   bool              `json:"isRead"`
}

type readingListResponse struct {
	ID        string         `json:"id"`
	Title     string         `json:"title"`
	Items     []itemResponse `json:"items"`
	CreatedAt string         `json:"createdAt"`
}

type listResponse struct {
	ReadingList *readingListResponse `json:"readingList"`
}

type catalogResponse struct {
	Resources []*resourceResponse `json:"resources"`
}

type dashboardResponse struct {
	Catalog     catalogResponse      `json:"catalog"`
	ReadingList *readingListResponse `json:"readingList"`
}

func mapResource(resource *pb.Resource) *resourceResponse {
	if resource == nil {
		return nil
	}
	return &resourceResponse{ID: resource.Id, Title: resource.Title, URL: resource.Url}
}

func mapCatalog(resources []*pb.Resource) catalogResponse {
	out := catalogResponse{Resources: make([]*resourceResponse, 0, len(resources))}
	for _, resource := range resources {
		out.Resources = append(out.Resources, mapResource(resource))
	}
	return out
}

func mapList(list *pb.ReadingList) *readingListResponse {
	if list == nil {
		return nil
	}
	out := &readingListResponse{ID: list.Id, Title: list.Title, CreatedAt: list.CreatedAt, Items: make([]itemResponse, 0, len(list.Items))}
	for _, item := range list.Items {
		out.Items = append(out.Items, itemResponse{Resource: mapResource(item.GetResource()), IsRead: item.GetIsRead()})
	}
	return out
}
