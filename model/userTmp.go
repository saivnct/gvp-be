package model

import "go.mongodb.org/mongo-driver/bson/primitive"

type UserTmp struct {
	ID       primitive.ObjectID `json:"_id" bson:"_id,omitempty"`
	Email    string             `json:"email" bson:"pkey,omitempty" validate:"required"`
	Username string             `json:"username" bson:"username,omitempty" validate:"required"`
	Password string             `json:"password" bson:"password,omitempty" validate:"required"`

	Authencode       string `json:"authencode" bson:"authencode,omitempty" validate:"required"`
	AuthencodeSendAt int64  `json:"authencodeSendAt" bson:"authencodeSendAt,omitempty" validate:"required"`
	NumberAuthenFail int32  `json:"numberAuthenFail" bson:"numberAuthenFail,omitempty" validate:"gte=0"`
	CreatedAt        int64  `json:"createdAt" bson:"createdAt,omitempty" validate:"required"`
}
