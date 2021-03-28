package controller

type APIResponse struct {
	Data interface{} `json:"data"`
}

type APIErrorResponse struct {
	Message string `json:"message"`
}
