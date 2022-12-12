package appgrpc

import (
	"context"
	"errors"
	"fmt"
	"gbb.go/gvp/dao"
	"gbb.go/gvp/model"
	"gbb.go/gvp/proto/grpcXVPPb"
	"gbb.go/gvp/static"
	"gbb.go/gvp/utils"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"log"
)

func (sv *XVPGRPCService) GetNewsTags(ctx context.Context, req *grpcXVPPb.GetNewsTagsRequest) (*grpcXVPPb.GetNewsTagsResponse, error) {
	_, ok := ctx.Value(GRPC_CTX_KEY_SESSION).(*GrpcSession)
	if !ok {
		log.Println("GetNewsTags - can not cast GrpcSession")
		return nil, status.Errorf(codes.PermissionDenied, "Invalid GrpcSession")
	}

	page := req.GetPage()
	pageSize := req.GetPageSize()

	if page < 0 {
		page = 0
	}
	if pageSize <= 0 {
		pageSize = static.Pagination_Default_PageSize
	}
	if pageSize > static.Pagination_Max_PageSize {
		pageSize = static.Pagination_Max_PageSize
	}

	filter := primitive.M{}
	totalItem, err := dao.GetNewsTagDAO().CountDocuments(ctx, filter)
	if err != nil {
		log.Println(err)
		return nil, status.Errorf(codes.Internal, fmt.Sprintf("Internal err: %v", err))
	}

	sort := primitive.M{"newsCount": -1}
	listTags, err := dao.GetNewsTagDAO().FetchListTags(ctx, page, pageSize, filter, sort)
	if err != nil {
		log.Println(err)
		return nil, status.Errorf(codes.Internal, fmt.Sprintf("Internal err: %v", err))
	}

	tags := []string{}
	for _, tag := range listTags {
		tags = append(tags, tag.Tag)
	}

	return &grpcXVPPb.GetNewsTagsResponse{
		Page:      page,
		PageSize:  pageSize,
		TotalItem: totalItem,
		Tags:      tags,
	}, nil
}

func CreateAndIncreaseTagNewsCount(tags []string) {
	log.Println("CreateAndIncreaseTagNewsCount", tags)

	if len(tags) == 0 {
		return
	}

	for _, tag := range tags {
		newsTag, err := dao.GetNewsTagDAO().FindByTag(context.Background(), tag)
		if err != nil && errors.Is(err, mongo.ErrNoDocuments) {
			newsTag = &model.NewsTag{
				Tag:         tag,
				NewsCount:   0,
				SearchCount: 0,
				CreatedAt:   utils.UTCNowMilli(),
			}
			newsTag, err = dao.GetNewsTagDAO().Save(context.Background(), newsTag)
			if err != nil {
				log.Println(err)
			}
		}
	}
	_, err := dao.GetNewsTagDAO().IncreaseListNewsCount(context.Background(), tags)
	if err != nil {
		log.Println(err)
	}
}

func CreateAndUpdateTagNewsCount(deleteTags []string, addTags []string) {
	log.Println("CreateAndUpdateTagNewsCount", deleteTags, addTags)

	if len(addTags) > 0 {
		for _, tag := range addTags {
			newsTag, err := dao.GetNewsTagDAO().FindByTag(context.Background(), tag)
			if err != nil && errors.Is(err, mongo.ErrNoDocuments) {
				newsTag = &model.NewsTag{
					Tag:         tag,
					NewsCount:   0,
					SearchCount: 0,
					CreatedAt:   utils.UTCNowMilli(),
				}
				newsTag, err = dao.GetNewsTagDAO().Save(context.Background(), newsTag)
				if err != nil {
					log.Println(err)
				}
			}
		}
		_, err := dao.GetNewsTagDAO().IncreaseListNewsCount(context.Background(), addTags)
		if err != nil {
			log.Println(err)
		}
	}

	if len(deleteTags) > 0 {
		_, err := dao.GetNewsTagDAO().DecreaseListNewsCount(context.Background(), deleteTags)
		if err != nil {
			log.Println(err)
		}
	}
}

func IncreaseTagsSearchCount(tags []string) {
	if len(tags) > 0 {
		_, err := dao.GetNewsTagDAO().IncreaseListSearchCount(context.Background(), tags)

		if err != nil {
			log.Println(err)
		}
	}

}
