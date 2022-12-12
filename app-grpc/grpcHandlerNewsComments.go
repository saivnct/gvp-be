package appgrpc

import (
	"context"
	"fmt"
	"gbb.go/gvp/dao"
	"gbb.go/gvp/model"
	"gbb.go/gvp/proto/grpcXVPPb"
	"gbb.go/gvp/static"
	"gbb.go/gvp/utils"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"log"
	"strings"
)

func (sv *XVPGRPCService) CreateNewsComment(ctx context.Context, req *grpcXVPPb.CreateNewsCommentRequest) (*grpcXVPPb.CreateNewsCommentResponse, error) {
	grpcSession, ok := ctx.Value(GRPC_CTX_KEY_SESSION).(*GrpcSession)
	if !ok {
		log.Println("CreateNewsComment - can not cast GrpcSession")
		return nil, status.Errorf(codes.PermissionDenied, "Invalid GrpcSession")
	}

	if grpcSession.User.Role < grpcXVPPb.USER_ROLE_NORMAL_USER {
		return nil, status.Errorf(codes.PermissionDenied, "Not Allowed Guest User")
	}

	newsId := strings.TrimSpace(req.GetNewsId())
	if len(newsId) == 0 {
		return nil, status.Errorf(codes.InvalidArgument, "Invalid Arguments")
	}

	content := strings.TrimSpace(req.GetContent())
	if len(content) == 0 {
		return nil, status.Errorf(codes.InvalidArgument, "Invalid Arguments")
	}

	news, err := dao.GetNewsDAO().FindByNewsId(ctx, newsId)
	if news == nil {
		return nil, status.Errorf(codes.InvalidArgument, "Invalid News: %v", req.GetNewsId())
	}

	commentAncestors := []string{}

	parentCommentId := strings.TrimSpace(req.GetParentCommentId())
	if len(parentCommentId) > 0 {
		parentComment, _ := dao.GetNewsCommentDAO().FindByCommentId(ctx, parentCommentId)
		if parentComment == nil {
			return nil, status.Errorf(codes.InvalidArgument, "Invalid Parent Comment: %v", parentCommentId)
		}

		if parentComment.NewsId != newsId {
			return nil, status.Errorf(codes.InvalidArgument, "Invalid Parent Comment: %v", parentCommentId)
		}

		commentAncestors = parentComment.CommentAncestors
	}

	commentId := utils.GenerateUUID()
	commentAncestors = append(commentAncestors, commentId)

	newsComment := &model.NewsComment{
		CommentId:        commentId,
		NewsId:           newsId,
		CommentAncestors: commentAncestors,
		Username:         grpcSession.User.Username,
		Content:          content,
		CreatedAt:        utils.UTCNowMilli(),
	}

	newsComment, err = dao.GetNewsCommentDAO().Save(ctx, newsComment)
	if err != nil {
		log.Println(err)
		return nil, status.Errorf(codes.Internal, fmt.Sprintf("Internal err: %v", err))
	}

	return &grpcXVPPb.CreateNewsCommentResponse{
		NewsCommentInfo: GetNewsCommentInfo(newsComment),
	}, nil
}

func (sv *XVPGRPCService) UpdateNewsComment(ctx context.Context, req *grpcXVPPb.UpdateNewsCommentRequest) (*grpcXVPPb.UpdateNewsCommentResponse, error) {
	grpcSession, ok := ctx.Value(GRPC_CTX_KEY_SESSION).(*GrpcSession)
	if !ok {
		log.Println("UpdateNewsComment - can not cast GrpcSession")
		return nil, status.Errorf(codes.PermissionDenied, "Invalid GrpcSession")
	}

	if grpcSession.User.Role < grpcXVPPb.USER_ROLE_NORMAL_USER {
		return nil, status.Errorf(codes.PermissionDenied, "Not Allowed Guest User")
	}

	commentId := strings.TrimSpace(req.GetCommentId())
	if len(commentId) == 0 {
		return nil, status.Errorf(codes.InvalidArgument, "Invalid Arguments")
	}

	content := strings.TrimSpace(req.GetContent())
	if len(content) == 0 {
		return nil, status.Errorf(codes.InvalidArgument, "Invalid Arguments")
	}

	newsComment, err := dao.GetNewsCommentDAO().FindByCommentId(ctx, commentId)
	if newsComment == nil {
		return nil, status.Errorf(codes.InvalidArgument, "Invalid Comment: %v", commentId)
	}

	if newsComment.Username != grpcSession.User.Username && !grpcSession.User.IsModeratorPermission() {
		return nil, status.Errorf(codes.PermissionDenied, "Permission Denied")
	}

	if content == newsComment.Content {
		return &grpcXVPPb.UpdateNewsCommentResponse{
			NewsCommentInfo: GetNewsCommentInfo(newsComment),
		}, nil
	}

	updateFields := primitive.M{}
	updateFields["content"] = content

	update := primitive.M{"$set": updateFields}

	newsComment, err = dao.GetNewsCommentDAO().UpdateByCommentId(ctx, commentId, update, []interface{}{}, false)
	if err != nil {
		log.Println(err)
		return nil, status.Errorf(codes.Internal, fmt.Sprintf("Internal err: %v", err))
	}

	return &grpcXVPPb.UpdateNewsCommentResponse{
		NewsCommentInfo: GetNewsCommentInfo(newsComment),
	}, nil
}

func (sv *XVPGRPCService) DeleteNewsComment(ctx context.Context, req *grpcXVPPb.DeleteNewsCommentRequest) (*grpcXVPPb.DeleteNewsCommentResponse, error) {
	grpcSession, ok := ctx.Value(GRPC_CTX_KEY_SESSION).(*GrpcSession)
	if !ok {
		log.Println("UpdateNewsComment - can not cast GrpcSession")
		return nil, status.Errorf(codes.PermissionDenied, "Invalid GrpcSession")
	}

	if grpcSession.User.Role < grpcXVPPb.USER_ROLE_NORMAL_USER {
		return nil, status.Errorf(codes.PermissionDenied, "Not Allowed Guest User")
	}

	commentId := strings.TrimSpace(req.GetCommentId())
	if len(commentId) == 0 {
		return nil, status.Errorf(codes.InvalidArgument, "Invalid Arguments")
	}

	newsComment, err := dao.GetNewsCommentDAO().FindByCommentId(ctx, commentId)
	if newsComment == nil {
		return nil, status.Errorf(codes.InvalidArgument, "Invalid Comment: %v", commentId)
	}

	if newsComment.Username != grpcSession.User.Username && !grpcSession.User.IsModeratorPermission() {
		return nil, status.Errorf(codes.PermissionDenied, "Permission Denied")
	}

	newsComment, err = dao.GetNewsCommentDAO().DeleteByCommentId(ctx, commentId)
	if err != nil {
		log.Println(err)
		return nil, status.Errorf(codes.Internal, fmt.Sprintf("Internal err: %v", err))
	}

	return &grpcXVPPb.DeleteNewsCommentResponse{
		CommentId: commentId,
	}, nil
}

func (sv *XVPGRPCService) GetNewsComments(ctx context.Context, req *grpcXVPPb.GetNewsCommentsRequest) (*grpcXVPPb.GetNewsCommentsResponse, error) {
	newsId := strings.TrimSpace(req.GetNewsId())
	if len(newsId) == 0 {
		return nil, status.Errorf(codes.InvalidArgument, "Invalid Arguments")
	}

	page := req.GetPage()
	pageSize := req.GetPageSize()

	if page < 0 {
		page = 0
	}
	if pageSize <= 0 || pageSize > 30 {
		pageSize = static.Pagination_Default_PageSize
	}

	parentCommentId := strings.TrimSpace(req.GetParentCommentId())

	totalItem, err := dao.GetNewsCommentDAO().CountListNewsComments(ctx, newsId, parentCommentId)
	if err != nil {
		log.Println(err)
		return nil, status.Errorf(codes.Internal, fmt.Sprintf("Internal err: %v", err))
	}

	listNewsComments, err := dao.GetNewsCommentDAO().FetchListNewsComments(ctx, page, pageSize, newsId, parentCommentId)
	if err != nil {
		log.Println(err)
		return nil, status.Errorf(codes.Internal, fmt.Sprintf("Internal err: %v", err))
	}

	newsCommentInfos := []*grpcXVPPb.NewsCommentInfo{}
	for _, newsComment := range listNewsComments {
		newsCommentInfos = append(newsCommentInfos, GetNewsCommentInfo(newsComment))
	}

	return &grpcXVPPb.GetNewsCommentsResponse{
		Page:            page,
		PageSize:        pageSize,
		TotalItem:       totalItem,
		NewsCommentInfo: newsCommentInfos,
	}, nil
}

func GetNewsCommentInfo(newsComment *model.NewsComment) *grpcXVPPb.NewsCommentInfo {
	return &grpcXVPPb.NewsCommentInfo{
		NewsId:           newsComment.NewsId,
		CommentId:        newsComment.CommentId,
		CommentAncestors: newsComment.CommentAncestors,
		Username:         newsComment.Username,
		Content:          newsComment.Content,
		CommentAt:        newsComment.CreatedAt,
	}
}
