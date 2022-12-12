package model

import (
	"gbb.go/gvp/proto/grpcXVPPb"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type User struct {
	ID       primitive.ObjectID  `json:"_id" bson:"_id,omitempty"`
	Username string              `json:"username" bson:"pkey,omitempty" validate:"required"`
	Email    string              `json:"email" bson:"email,omitempty" validate:"required"`
	Password string              `json:"password" bson:"password,omitempty" validate:"required"`
	Role     grpcXVPPb.USER_ROLE `json:"role" bson:"role,omitempty" validate:"required"`
	//info
	PhoneNumber string                `json:"phoneNumber" bson:"phoneNumber,omitempty"`
	FirstName   string                `json:"firstName" bson:"firstName,omitempty"`
	LastName    string                `json:"lastName" bson:"lastName,omitempty"`
	Avatar      string                `json:"avatar" bson:"avatar,omitempty"`
	Gender      grpcXVPPb.USER_GENDER `json:"gender" bson:"gender,omitempty"`
	Birthday    int64                 `json:"birthday" bson:"birthday,omitempty"`

	CreatedAt int64 `json:"createdAt" bson:"createdAt,omitempty" validate:"required"`
}

func (user *User) IsNotGuest() bool {
	return user.Role > grpcXVPPb.USER_ROLE_GUEST
}

func (user *User) IsContentUserPermission() bool {
	return user.Role >= grpcXVPPb.USER_ROLE_CONTENT_USER && user.Role != grpcXVPPb.USER_ROLE_KOL_USER
}

func (user *User) IsModeratorPermission() bool {
	return user.Role >= grpcXVPPb.USER_ROLE_MODERATOR
}

func (user *User) IsAdminPermission() bool {
	return user.Role >= grpcXVPPb.USER_ROLE_ADMIN
}

func (user *User) IsSuperAdminPermission() bool {
	return user.Role >= grpcXVPPb.USER_ROLE_SUPER_ADMIN
}
