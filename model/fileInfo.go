package model

import (
	"gbb.go/gvp/proto/grpcXVPPb"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type FileInfo struct {
	ID                      primitive.ObjectID          `json:"_id" bson:"_id,omitempty"`
	FileId                  string                      `json:"fileId" bson:"pkey,omitempty" validate:"required"`
	FileUrl                 string                      `json:"fileUrl" bson:"fileUrl,omitempty"`
	FileName                string                      `json:"fileName" bson:"fileName,omitempty"`
	FileSize                int64                       `json:"fileSize" bson:"fileSize,omitempty"`
	Checksum                string                      `json:"checksum" bson:"checksum,omitempty"`
	MediaType               grpcXVPPb.MEDIA_TYPE        `json:"mediaType" bson:"mediaType,omitempty" validate:"gte=0"`
	MediaStreamType         grpcXVPPb.MEDIA_STREAM_TYPE `json:"mediaStreamType" bson:"mediaStreamType,omitempty"`
	OnDemandMediaMainFileId string                      `json:"onDemandMediaMainFileId" bson:"onDemandMediaMainFileId,omitempty"`
	MediaEncKey             string                      `json:"mediaEncKey" bson:"mediaEncKey,omitempty"`
	Resolution              grpcXVPPb.VIDEO_RESOLUTION  `json:"resolution" bson:"resolution,omitempty" validate:"gte=0"`
	NewsId                  string                      `json:"newsId" bson:"newsId,omitempty"`
	MainPreview             bool                        `json:"mainPreview" bson:"mainPreview,omitempty"`
	CreatedAt               int64                       `json:"createdAt" bson:"createdAt,omitempty" validate:"required"`
}
