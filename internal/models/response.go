package models

// Response standart API yanıt formatı
type Response struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

// HealthResponse sistem sağlık durumu için yanıt
type HealthResponse struct {
	Status   string            `json:"status"`
	Services map[string]string `json:"services"`
}
