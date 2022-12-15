package model

// A Message contains the message to send to a recipient
type Message struct {

	// The title of the message
	Title string `json:"title"`

	// The body of the message
	Body string `json:"body"`
}

type Notification struct {
	Id        string            `json:"id"`
	Recipient string            `json:"recipient,omitempty"`
	Data      map[string]string `json:"data,omitempty"`
	Message   *Message          `json:"message,omitempty"`
}

// PushRequest is a push notification request
type PushRequest struct {

	// The list of recipients registration tokens
	Recipients []string `json:"recipients,omitempty"`

	// Custom key-value pair to include within the notification
	Data map[string]string `json:"data,omitempty"`

	// The notification message
	Message *Message `json:"message,omitempty"`
}

type NotificationError struct {
	Recipient string `json:"recipient,omitempty"`
	Message   string `json:"message,omitempty"`
}

// PushResponse is a response to a push notification request
type PushResponse struct {

	// The number of successful notifications sent
	NotificationsCount int `json:"notificationsCount"`

	// The number of notifications errors
	ErrorsCount int `json:"errorsCount"`

	// The list of successful notifications details
	Notifications []Notification `json:"notifications,omitempty"`

	// The list of notifications error details
	Errors []NotificationError `json:"errors,omitempty"`
}

// Error includes failed HTTP request information
type Error struct {
	// The HTTP status
	Status int `json:"status"`
	// The error message
	Message string `json:"message,omitempty"`
}
