package model

import "go.mongodb.org/mongo-driver/bson/primitive"

type NewsParticipant struct {
	ID          primitive.ObjectID `json:"_id" bson:"_id,omitempty"`
	Participant string             `json:"participant" bson:"pkey,omitempty" validate:"required"`
	NewsCount   int64              `json:"newsCount" bson:"newsCount,omitempty" validate:"gte=0"`
	SearchCount int64              `json:"searchCount" bson:"searchCount,omitempty" validate:"gte=0"`
	CreatedAt   int64              `json:"createdAt" bson:"createdAt,omitempty" validate:"required"`
}
