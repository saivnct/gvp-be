package appgrpc

import (
	"context"
	"fmt"
	"gbb.go/gvp/dao"
	"gbb.go/gvp/model"
	"gbb.go/gvp/proto/grpcXVPPb"
	"gbb.go/gvp/utils"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"log"
	"strings"
)

func (sv *XVPGRPCService) CreateCategory(ctx context.Context, req *grpcXVPPb.CreateCategoryRequest) (*grpcXVPPb.CreateCategoryResponse, error) {
	catName := strings.TrimSpace(req.GetCatName())
	catDescription := strings.TrimSpace(req.GetCatDescription())

	if len(catName) == 0 {
		return nil, status.Errorf(codes.InvalidArgument, "Invalid Arguments")
	}

	category := &model.Category{
		CatId:       utils.GenerateUUID(),
		Name:        catName,
		Description: catDescription,
		CreatedAt:   utils.UTCNowMilli(),
	}

	category, err := dao.GetCategoryDAO().Save(ctx, category)
	if err != nil {
		log.Println(err)
		return nil, status.Errorf(codes.Internal, fmt.Sprintf("Internal err: %v", err))
	}

	return &grpcXVPPb.CreateCategoryResponse{
		CategoryInfo: &grpcXVPPb.CategoryInfo{
			CatId:          category.CatId,
			CatName:        category.Name,
			CatDescription: category.Description,
		},
	}, nil
}

func (sv *XVPGRPCService) UpdateCategory(ctx context.Context, req *grpcXVPPb.UpdateCategoryRequest) (*grpcXVPPb.UpdateCategoryResponse, error) {
	catId := req.GetCatId()
	catName := strings.TrimSpace(req.GetCatName())
	catDescription := strings.TrimSpace(req.GetCatDescription())

	if len(catId) == 0 {
		return nil, status.Errorf(codes.InvalidArgument, "Invalid Arguments")
	}

	if len(catName) == 0 && len(catDescription) == 0 {
		return nil, status.Errorf(codes.InvalidArgument, "Invalid Arguments")
	}

	category, err := dao.GetCategoryDAO().FindByCatId(ctx, req.GetCatId())
	if category == nil {
		return nil, status.Errorf(codes.NotFound, "Invalid Category")
	}

	updateFields := primitive.M{}
	if len(catName) > 0 {
		updateFields["name"] = catName
	}

	if len(catDescription) > 0 {
		updateFields["description"] = catDescription
	}

	update := primitive.M{"$set": updateFields}

	category, err = dao.GetCategoryDAO().UpdateByCatId(ctx, category.CatId, update, []interface{}{}, false)
	if err != nil {
		log.Println(err)
		return nil, status.Errorf(codes.Internal, fmt.Sprintf("Internal err: %v", err))
	}

	return &grpcXVPPb.UpdateCategoryResponse{
		CategoryInfo: &grpcXVPPb.CategoryInfo{
			CatId:          category.CatId,
			CatName:        category.Name,
			CatDescription: category.Description,
		},
	}, nil
}

func (sv *XVPGRPCService) DeleteCategory(ctx context.Context, req *grpcXVPPb.DeleteCategoryRequest) (*grpcXVPPb.DeleteCategoryResponse, error) {
	if len(req.CatId) == 0 {
		return nil, status.Errorf(codes.InvalidArgument, "Invalid Arguments")
	}

	totalNews, err := dao.GetNewsDAO().CountByCatId(ctx, req.CatId)
	if err != nil {
		log.Println(err)
		return nil, status.Errorf(codes.Internal, fmt.Sprintf("Internal err: %v", err))
	}

	if totalNews > 0 {
		log.Println("not empty Category")
		return nil, status.Errorf(codes.Aborted, "not empty Category")
	}

	_, err = dao.GetCategoryDAO().DeleteByCatId(ctx, req.CatId)
	if err != nil {
		log.Println(err)
		return nil, status.Errorf(codes.Internal, fmt.Sprintf("Internal err: %v", err))
	}

	return &grpcXVPPb.DeleteCategoryResponse{
		CatId: req.CatId,
	}, nil
}

func (sv *XVPGRPCService) GetCategory(ctx context.Context, req *grpcXVPPb.GetCategoryRequest) (*grpcXVPPb.GetCategoryResponse, error) {
	if len(req.CatId) == 0 {
		return nil, status.Errorf(codes.InvalidArgument, "Invalid Arguments")
	}

	category, _ := dao.GetCategoryDAO().FindByCatId(ctx, req.CatId)
	if category == nil {
		return nil, status.Errorf(codes.NotFound, "Invalid Category")
	}

	return &grpcXVPPb.GetCategoryResponse{
		CategoryInfo: &grpcXVPPb.CategoryInfo{
			CatId:          category.CatId,
			CatName:        category.Name,
			CatDescription: category.Description,
		},
	}, nil
}

func (sv *XVPGRPCService) GetAllCategory(ctx context.Context, req *grpcXVPPb.GetAllCategoryRequest) (*grpcXVPPb.GetAllCategoryResponse, error) {

	allCategories, err := dao.GetCategoryDAO().GetAll(ctx)
	if err != nil {
		log.Println(err)
		return nil, status.Errorf(codes.Internal, fmt.Sprintf("Internal err: %v", err))
	}

	categoryInfos := []*grpcXVPPb.CategoryInfo{}

	for _, category := range allCategories {
		categoryInfo := &grpcXVPPb.CategoryInfo{
			CatId:          category.CatId,
			CatName:        category.Name,
			CatDescription: category.Description,
		}

		categoryInfos = append(categoryInfos, categoryInfo)
	}

	return &grpcXVPPb.GetAllCategoryResponse{
		CategoryInfos: categoryInfos,
	}, nil
}
