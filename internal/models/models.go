package models

import "go.mongodb.org/mongo-driver/bson/primitive"

// RequestType defines the type of service request
type RequestType string

const (
	Incident     RequestType = "Incident"
	Maintenance  RequestType = "Maintenance"
	Modification RequestType = "Modification"
)

// RequestStatus defines the status of a service request
type RequestStatus string

const (
	Open       RequestStatus = "Open"
	InProgress RequestStatus = "In Progress"
	Resolved   RequestStatus = "Resolved"
	Closed     RequestStatus = "Closed"
)

// User represents a user in the system
type User struct {
	ID       primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	Username string             `bson:"username" json:"username"`
	Password string             `bson:"password" json:"-"` // Store encrypted password
}

// Request represents a service desk request
type Request struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	Title       string             `bson:"title" json:"title"`
	Description string             `bson:"description" json:"description"`
	Type        RequestType        `bson:"type" json:"type"`
	Status      RequestStatus      `bson:"status" json:"status"`
	CreatedBy   primitive.ObjectID `bson:"createdBy" json:"createdBy"` // User ID
	CreatedAt   primitive.DateTime `bson:"createdAt" json:"createdAt"`
	UpdatedAt   primitive.DateTime `bson:"updatedAt" json:"updatedAt"`
}
