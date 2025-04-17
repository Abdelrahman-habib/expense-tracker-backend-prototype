package payloads

// createdResponse represents a resource creation response example
type createdResponse struct {
	Status  int         `json:"status" example:"201"`
	Message string      `json:"message" example:"Resource created successfully"`
	Data    interface{} `json:"data"`
}

// updatedResponse represents a resource update response example
type updatedResponse struct {
	Status  int         `json:"status" example:"200"`
	Message string      `json:"message" example:"Resource updated successfully"`
	Data    interface{} `json:"data"`
}

// deletedResponse represents a resource deletion response example
type deletedResponse struct {
	Status  int    `json:"status" example:"202"`
	Message string `json:"message" example:"Resource deleted successfully"`
}

// noContentResponse represents an empty response example
type noContentResponse struct {
	Status  int    `json:"status" example:"204"`
	Message string `json:"message" example:""`
}
