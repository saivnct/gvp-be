package model

import "go.mongodb.org/mongo-driver/bson/primitive"

type Category struct {
	ID          primitive.ObjectID `json:"_id" bson:"_id,omitempty"`
	CatId       string             `json:"catId" bson:"pkey,omitempty" validate:"required"`
	Name        string             `json:"name" bson:"name,omitempty" validate:"required"`
	Description string             `json:"description" bson:"description,omitempty"`

	CreatedAt int64 `json:"createdAt" bson:"createdAt,omitempty" validate:"required"`
}
