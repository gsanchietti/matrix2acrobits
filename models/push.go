package models

// Matrix Push Gateway API models (spec: https://spec.matrix.org/v1.16/push-gateway-api/)

// MatrixPushNotifyRequest represents the request body for POST /_matrix/push/v1/notify
type MatrixPushNotifyRequest struct {
	Notification MatrixNotification `json:"notification"`
}

// MatrixNotification contains the notification details from Matrix
type MatrixNotification struct {
	Content           map[string]interface{} `json:"content,omitempty"`
	Counts            *MatrixCounts          `json:"counts,omitempty"`
	Devices           []MatrixDevice         `json:"devices"`
	EventID           string                 `json:"event_id,omitempty"`
	Prio              string                 `json:"prio,omitempty"` // "high" or "low"
	RoomAlias         string                 `json:"room_alias,omitempty"`
	RoomID            string                 `json:"room_id,omitempty"`
	RoomName          string                 `json:"room_name,omitempty"`
	Sender            string                 `json:"sender,omitempty"`
	SenderDisplayName string                 `json:"sender_display_name,omitempty"`
	Type              string                 `json:"type,omitempty"`
	UserIsTarget      bool                   `json:"user_is_target,omitempty"`
}

// MatrixCounts represents unread message counts
type MatrixCounts struct {
	MissedCalls int `json:"missed_calls,omitempty"`
	Unread      int `json:"unread,omitempty"`
}

// MatrixDevice represents a device that should receive the push
type MatrixDevice struct {
	AppID     string                 `json:"app_id"`
	Data      map[string]interface{} `json:"data,omitempty"`
	Pushkey   string                 `json:"pushkey"`
	PushkeyTS int64                  `json:"pushkey_ts,omitempty"`
	Tweaks    map[string]interface{} `json:"tweaks,omitempty"`
}

// MatrixPushNotifyResponse represents the response for POST /_matrix/push/v1/notify
type MatrixPushNotifyResponse struct {
	Rejected []string `json:"rejected"`
}

// Acrobits Push Notification API models (spec: https://doc.acrobits.net/api/server/http_push.html)

// AcrobitsPushRequest represents a single push notification to Acrobits PNM
type AcrobitsPushRequest struct {
	Verb        string `json:"verb"`        // NotifyTextMessage, NotifyGenericTextMessage, etc.
	AppID       string `json:"AppId"`       // Application ID
	DeviceToken string `json:"DeviceToken"` // Device token
	Selector    string `json:"Selector,omitempty"`

	// For NotifyTextMessage (iOS 13+)
	Badge           int    `json:"Badge,omitempty"`
	Sound           string `json:"Sound,omitempty"`
	UserName        string `json:"UserName,omitempty"`
	UserDisplayName string `json:"UserDisplayName,omitempty"`
	Message         string `json:"Message,omitempty"`
	ContentType     string `json:"ContentType,omitempty"`
	ID              string `json:"Id,omitempty"`
	ThreadID        string `json:"ThreadId,omitempty"`
}

// AcrobitsPushResponse represents the response from Acrobits PNM
type AcrobitsPushResponse struct {
	Code     int    `json:"code"`
	Response string `json:"response"`
}

// Matrix Client-Server API pusher models (spec: https://spec.matrix.org/v1.16/client-server-api/#post_matrixclientv3pushersset)

// SetPusherRequest represents the request body for POST /_matrix/client/v3/pushers/set
type SetPusherRequest struct {
	AppDisplayName    string      `json:"app_display_name"`              // Human-readable app name
	AppID             string      `json:"app_id"`                        // Reverse-DNS style app identifier
	Append            bool        `json:"append"`                        // If true, add pusher; if false, replace existing
	Data              *PusherData `json:"data"`                          // Pusher-specific data (required if kind is not null)
	DeviceDisplayName string      `json:"device_display_name,omitempty"` // Human-readable device name
	Kind              *string     `json:"kind"`                          // "http", "email", or null to delete pusher
	Lang              string      `json:"lang"`                          // Preferred language (e.g., "en" or "en-US")
	ProfileTag        string      `json:"profile_tag,omitempty"`         // Identifier for device-specific rules
	Pushkey           string      `json:"pushkey"`                       // Unique identifier (routing token)
}

// PusherData contains pusher-specific configuration
type PusherData struct {
	Format string `json:"format,omitempty"` // "event_id_only" for HTTP pushers
	URL    string `json:"url,omitempty"`    // HTTPS URL for push gateway (required for http kind)
}
