package payloads

import (
	"net/http"

	"github.com/go-chi/render"
)

const (
	DeleteMessage = "Resource deleted successfully"
	UpdateMessage = "Resource updated successfully"
	CreateMessage = "Resource created successfully"
	OkMessage     = "Success"
)

// Response represents the standard API response format
// @Description Standard API response wrapper
type Response struct {
	Status  int         `json:"status" example:"200" enums:"200,202,204"`
	Message string      `json:"message,omitempty" example:"Success" enums:"Success,Resource created successfully,Resource updated successfully,Resource deleted successfully"`
	Data    interface{} `json:"data,omitempty"`
	Meta    struct {
		Query     string `json:"query,omitempty"`
		Limit     int32  `json:"limit,omitempty"`
		Count     int    `json:"count,omitempty"`
		NextToken string `json:"next_token,omitempty"`
	} `json:"meta"`
}

func (rd *Response) Render(w http.ResponseWriter, r *http.Request) error {
	render.Status(r, rd.Status)
	return nil
}

// NewResponse creates a new standard response
func NewResponse(status int, message string, data interface{}) render.Renderer {
	return &Response{
		Status:  status,
		Message: message,
		Data:    data,
	}
}

// Common response helpers
func OK(data interface{}) render.Renderer {
	return NewResponse(http.StatusOK, OkMessage, data)
}

func Created(data interface{}) render.Renderer {
	return NewResponse(http.StatusCreated, CreateMessage, data)
}

func Updated(data interface{}) render.Renderer {
	return NewResponse(http.StatusOK, UpdateMessage, data)
}

func Deleted() render.Renderer {
	return NewResponse(http.StatusOK, DeleteMessage, nil)
}

func NoContent() render.Renderer {
	return NewResponse(http.StatusNoContent, "", nil)
}

func List(data interface{}, count int) render.Renderer {
	resp := &Response{
		Status:  http.StatusOK,
		Message: OkMessage,
		Data:    data,
	}
	resp.Meta.Count = count
	return resp
}

// Search creates a new search response
func Search(data interface{}, query string, limit int32, count int) render.Renderer {
	resp := &Response{
		Status:  http.StatusOK,
		Message: OkMessage,
		Data:    data,
	}
	resp.Meta.Query = query
	resp.Meta.Limit = limit
	resp.Meta.Count = count
	return resp
}

// Paginated creates a new paginated response
func Paginated(data interface{}, nextToken string, limit int32) render.Renderer {
	resp := &Response{
		Status:  http.StatusOK,
		Message: OkMessage,
		Data:    data,
	}
	resp.Meta.NextToken = nextToken
	resp.Meta.Limit = limit
	return resp
}
