package model

import (
	"gbb.go/gvp/proto/grpcXVPPb"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type News struct {
	ID                    primitive.ObjectID    `json:"_id" bson:"_id,omitempty"`
	NewsId                string                `json:"newsId" bson:"pkey,omitempty" validate:"required"`
	Author                string                `json:"author" bson:"author,omitempty" validate:"required"`
	Title                 string                `json:"title" bson:"title,omitempty" validate:"required"`
	Description           string                `json:"description" bson:"description,omitempty"`
	Participants          []string              `json:"participants" bson:"participants,omitempty"`
	Categories            []string              `json:"categories" bson:"categories,omitempty"` //list catId
	Tags                  []string              `json:"tags" bson:"tags,omitempty"`             //list tags
	EnableComment         bool                  `json:"enableComment" bson:"enableComment,omitempty"`
	PreviewImages         []string              `json:"previewImages" bson:"previewImages,omitempty"` //list fileIds
	Medias                []string              `json:"medias" bson:"medias,omitempty"`               //list fileIds
	MediaEncKey           string                `json:"mediaEncKey" bson:"mediaEncKey,omitempty"`
	MediaEncIV            string                `json:"mediaEncIV" bson:"mediaEncIV,omitempty"`
	Views                 int64                 `json:"views" bson:"views,omitempty"`
	WeekViews             int64                 `json:"weekViews" bson:"weekViews,omitempty"`
	MonthViews            int64                 `json:"monthViews" bson:"monthViews,omitempty"`
	CurrentViewsWeek      int64                 `json:"currentViewsWeek" bson:"currentViewsWeek,omitempty" validate:"required"`
	CurrentViewsMonth     int64                 `json:"currentViewsMonth" bson:"currentViewsMonth,omitempty" validate:"required"`
	Likes                 int64                 `json:"likes" bson:"likes,omitempty"`
	WeekLikes             int64                 `json:"weekLikes" bson:"weekLikes,omitempty"`
	MonthLikes            int64                 `json:"monthLikes" bson:"monthLikes,omitempty"`
	CurrentLikesWeek      int64                 `json:"currentLikesWeek" bson:"currentLikesWeek,omitempty" validate:"required"`
	CurrentLikesMonth     int64                 `json:"currentLikesMonth" bson:"currentLikesMonth,omitempty" validate:"required"`
	LikedBy               []string              `json:"likedBy" bson:"likedBy,omitempty"`
	AccumulateRatingPoint int64                 `json:"accumulateRatingPoint" bson:"accumulateRatingPoint,omitempty" validate:"gte=0"` //sum of (rating from 0 to 5 stars)
	RatingCount           int64                 `json:"ratingCount" bson:"ratingCount,omitempty" validate:"gte=0"`                     //number of rating
	Rating                float64               `json:"rating" bson:"rating,omitempty" validate:"gte=0"`                               //number of rating
	RatedBy               []string              `json:"ratedBy" bson:"ratedBy,omitempty"`
	Status                grpcXVPPb.NEWS_STATUS `json:"status" bson:"status,omitempty" validate:"gte=0"`
	CreatedAt             int64                 `json:"createdAt" bson:"createdAt,omitempty" validate:"required"`
}
