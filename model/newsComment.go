package model

import "go.mongodb.org/mongo-driver/bson/primitive"

type NewsComment struct {
	ID               primitive.ObjectID `json:"_id" bson:"_id,omitempty"`
	CommentId        string             `json:"commentId" bson:"pkey,omitempty" validate:"required"`
	NewsId           string             `json:"newsId" bson:"newsId,omitempty" validate:"required"`
	CommentAncestors []string           `json:"commentAncestors" bson:"commentAncestors,omitempty"`
	Username         string             `json:"username" bson:"username,omitempty" validate:"required"`
	Content          string             `json:"content" bson:"content,omitempty" validate:"required"`
	CreatedAt        int64              `json:"createdAt" bson:"createdAt,omitempty" validate:"required"`
}
