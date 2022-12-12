package model

import "go.mongodb.org/mongo-driver/bson/primitive"

type EmailLockReg struct {
	ID primitive.ObjectID `json:"_id" bson:"_id,omitempty"`

	Email                  string `json:"email" bson:"pkey,omitempty" validate:"required"`
	NumAuthencodeSend      int32  `json:"numAuthencodeSend" bson:"numAuthencodeSend,omitempty" validate:"gte=0"`
	LastDateAuthencodeSend int64  `json:"lastDateAuthencodeSend" bson:"lastDateAuthencodeSend,omitempty"`
	Locked                 bool   `json:"locked" bson:"locked,omitempty"`
}
