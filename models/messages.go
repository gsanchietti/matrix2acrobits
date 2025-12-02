package models

// SendMessageRequest mirrors the OpenAPI schema for sending a message from Acrobits.
type SendMessageRequest struct {
	From                    string `json:"from"`
	Password                string `json:"password"`
	SMSTo                   string `json:"sms_to"`
	SMSBody                 string `json:"sms_body"`
	ContentType             string `json:"content_type"`
	DispositionNotification string `json:"disposition_notification"`
}

// SendMessageResponse reports the Matrix event ID returned to Acrobits.
type SendMessageResponse struct {
	SMSID string `json:"sms_id"`
}

// FetchMessagesRequest mirrors the OpenAPI schema for polling new messages.
type FetchMessagesRequest struct {
	Username   string `json:"username"`
	Password   string `json:"password"`
	LastID     string `json:"last_id"`
	LastSentID string `json:"last_sent_id"`
	Device     string `json:"device"`
}

// FetchMessagesResponse is the payload returned from the fetch_messages endpoint.
type FetchMessagesResponse struct {
	Date         string    `json:"date"`
	ReceivedSMSS []Message `json:"received_smss"`
	SentSMSS     []Message `json:"sent_smss"`
}

// Message represents the shared schema used in both fetch and send responses.
type Message struct {
	SMSID       string `json:"sms_id"`
	SendingDate string `json:"sending_date"`
	Sender      string `json:"sender"`
	Recipient   string `json:"recipient"`
	SMSText     string `json:"sms_text"`
	ContentType string `json:"content_type"`
	StreamID    string `json:"stream_id"`
}
