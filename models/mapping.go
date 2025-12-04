package models

// MappingRequest defines the payload used by the SMS-to-Matrix mapping API.
type MappingRequest struct {
	SMSNumber string `json:"sms_number"`
	MatrixID  string `json:"matrix_id,omitempty"`
	RoomID    string `json:"room_id"`
	UserName  string `json:"user_name,omitempty"`
}

// MappingResponse is returned once a mapping has been created or looked up.
type MappingResponse struct {
	SMSNumber string `json:"sms_number"`
	MatrixID  string `json:"matrix_id"`
	RoomID    string `json:"room_id"`
	UserName  string `json:"user_name,omitempty"`
	UpdatedAt string `json:"updated_at"`
}
