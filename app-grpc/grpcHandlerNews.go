package appgrpc

import (
	"bytes"
	"context"
	"crypto/md5"
	"fmt"
	"gbb.go/gvp/dao"
	"gbb.go/gvp/model"
	"gbb.go/gvp/proto/grpcXVPPb"
	"gbb.go/gvp/s3Handler"
	"gbb.go/gvp/static"
	"gbb.go/gvp/utils"
	"github.com/gabriel-vasile/mimetype"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"golang.org/x/exp/slices"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"io"
	"log"
	"strings"
)

func (sv *XVPGRPCService) CreateNews(ctx context.Context, req *grpcXVPPb.CreateNewsRequest) (*grpcXVPPb.CreateNewsResponse, error) {
	title := strings.TrimSpace(req.GetTitle())
	description := strings.TrimSpace(req.GetDescription())
	catIds := req.GetCatIds()
	tags := req.GetTags()
	participants := req.GetParticipants()
	enableComment := req.GetEnableComment()

	grpcSession, ok := ctx.Value(GRPC_CTX_KEY_SESSION).(*GrpcSession)
	if !ok {
		log.Println("CreateNews - can not cast GrpcSession")
		return nil, status.Errorf(codes.PermissionDenied, "Invalid GrpcSession")
	}

	if len(title) == 0 || len(catIds) == 0 {
		return nil, status.Errorf(codes.InvalidArgument, "Invalid Arguments")
	}

	for _, catId := range catIds {
		category, _ := dao.GetCategoryDAO().FindByCatId(ctx, catId)
		if category == nil {
			return nil, status.Errorf(codes.InvalidArgument, "Invalid Category %v", catId)
		}
	}

	mediaEncKey, err := utils.GenerateRandomHex(16)
	if err != nil {
		log.Println(err)
		return nil, status.Errorf(codes.Internal, fmt.Sprintf("Internal err: %v", err))
	}

	mediaEncIV, err := utils.GenerateRandomHex(16)
	if err != nil {
		log.Println(err)
		return nil, status.Errorf(codes.Internal, fmt.Sprintf("Internal err: %v", err))
	}

	news := &model.News{
		NewsId:                utils.GenerateUUID(),
		Author:                grpcSession.User.Username,
		Title:                 title,
		Description:           description,
		Participants:          participants,
		Categories:            catIds,
		Tags:                  tags,
		EnableComment:         enableComment,
		PreviewImages:         []string{},
		Medias:                []string{},
		MediaEncKey:           mediaEncKey,
		MediaEncIV:            mediaEncIV,
		Views:                 0,
		WeekViews:             0,
		MonthViews:            0,
		CurrentViewsWeek:      utils.UTCNowBeginningOfWeek(),
		CurrentViewsMonth:     utils.UTCNowBeginningOfMonth(),
		Likes:                 0,
		WeekLikes:             0,
		MonthLikes:            0,
		CurrentLikesWeek:      utils.UTCNowBeginningOfWeek(),
		CurrentLikesMonth:     utils.UTCNowBeginningOfMonth(),
		LikedBy:               []string{},
		AccumulateRatingPoint: 0,
		RatingCount:           0,
		Rating:                0,
		RatedBy:               []string{},
		Status:                grpcXVPPb.NEWS_STATUS_ACTIVED, //TODO - will be changed to PENDING on production deployment
		CreatedAt:             utils.UTCNowMilli(),
	}

	news, err = dao.GetNewsDAO().Save(ctx, news)
	if err != nil {
		log.Println(err)
		return nil, status.Errorf(codes.Internal, fmt.Sprintf("Internal err: %v", err))
	}

	if len(tags) > 0 {
		go CreateAndIncreaseTagNewsCount(tags)
	}

	if len(participants) > 0 {
		go CreateAndIncreaseParticipantNewsCount(participants)
	}

	newsInfo, err := GetNewsInfo(news, grpcSession.User)
	if err != nil {
		log.Println(err)
		return nil, status.Errorf(codes.Internal, fmt.Sprintf("Internal err: %v", err))
	}

	return &grpcXVPPb.CreateNewsResponse{
		NewsInfo: newsInfo,
	}, nil
}

func (sv *XVPGRPCService) UpdateNewsInfo(ctx context.Context, req *grpcXVPPb.UpdateNewsInfoRequest) (*grpcXVPPb.UpdateNewsInfoResponse, error) {
	newsId := strings.TrimSpace(req.GetNewsId())
	title := strings.TrimSpace(req.GetTitle())
	description := strings.TrimSpace(req.GetDescription())
	catIds := req.GetCatIds()
	newTags := req.GetTags()
	newParticipants := req.GetParticipants()
	enableComment := req.GetEnableComment()

	grpcSession, ok := ctx.Value(GRPC_CTX_KEY_SESSION).(*GrpcSession)
	if !ok {
		log.Println("UpdateNewsInfo - can not cast GrpcSession")
		return nil, status.Errorf(codes.PermissionDenied, "Invalid GrpcSession")
	}

	if len(newsId) == 0 {
		return nil, status.Errorf(codes.InvalidArgument, "Invalid Arguments")
	}

	news, err := dao.GetNewsDAO().FindByNewsId(ctx, newsId)
	if news == nil {
		return nil, status.Errorf(codes.InvalidArgument, "Invalid News: %v", req.GetNewsId())
	}

	if news.Author != grpcSession.User.Username && !grpcSession.User.IsModeratorPermission() {
		log.Println("UpdateNewsInfo - Permission Denied GrpcSession", grpcSession.User.Username)
		return nil, status.Errorf(codes.PermissionDenied, "Permission Denied")
	}

	oldTags := news.Tags
	oldParticipants := news.Participants

	updateFields := primitive.M{}
	if len(title) > 0 {
		updateFields["title"] = title
	}
	if len(description) > 0 {
		updateFields["description"] = description
	}
	if len(catIds) > 0 {
		for _, catId := range catIds {
			category, _ := dao.GetCategoryDAO().FindByCatId(ctx, catId)
			if category == nil {
				return nil, status.Errorf(codes.InvalidArgument, "Invalid Category %v", catId)
			}
		}

		updateFields["categories"] = catIds
	}
	if len(newTags) > 0 {
		updateFields["tags"] = newTags
	}
	if len(newParticipants) > 0 {
		updateFields["participants"] = newParticipants
	}

	if enableComment != news.EnableComment {
		updateFields["enableComment"] = enableComment
	}

	update := primitive.M{"$set": updateFields}

	news, err = dao.GetNewsDAO().UpdateByNewsId(ctx, newsId, update, []interface{}{}, false)
	if err != nil {
		log.Println(err)
		return nil, status.Errorf(codes.Internal, fmt.Sprintf("Internal err: %v", err))
	}

	if len(newTags) > 0 {
		go func(oldTags []string, newTags []string) {
			deleteTags := []string{}
			addTags := []string{}
			for _, tag := range oldTags {
				idx := slices.IndexFunc(newTags, func(t string) bool { return t == tag })
				if idx < 0 {
					deleteTags = append(deleteTags, tag)
				}
			}

			for _, tag := range newTags {
				idx := slices.IndexFunc(oldTags, func(t string) bool { return t == tag })
				if idx < 0 {
					addTags = append(addTags, tag)
				}
			}

			if len(deleteTags) > 0 || len(addTags) > 0 {
				CreateAndUpdateTagNewsCount(deleteTags, addTags)
			}
		}(oldTags, newTags)
	}

	if len(newParticipants) > 0 {
		go func(oldParticipants []string, newParticipants []string) {
			deleteParticipants := []string{}
			addParticipants := []string{}
			for _, participant := range oldParticipants {
				idx := slices.IndexFunc(newParticipants, func(p string) bool { return p == participant })
				if idx < 0 {
					deleteParticipants = append(deleteParticipants, participant)
				}
			}

			for _, participant := range newParticipants {
				idx := slices.IndexFunc(oldParticipants, func(p string) bool { return p == participant })
				if idx < 0 {
					addParticipants = append(addParticipants, participant)
				}
			}

			if len(deleteParticipants) > 0 || len(addParticipants) > 0 {
				CreateAndUpdateParticipantNewsCount(deleteParticipants, addParticipants)
			}
		}(oldParticipants, newParticipants)
	}

	newsInfo, err := GetNewsInfo(news, grpcSession.User)
	if err != nil {
		log.Println(err)
		return nil, status.Errorf(codes.Internal, fmt.Sprintf("Internal err: %v", err))
	}

	return &grpcXVPPb.UpdateNewsInfoResponse{
		NewsInfo: newsInfo,
	}, nil
}

func (sv *XVPGRPCService) DeleteNews(ctx context.Context, req *grpcXVPPb.DeleteNewsRequest) (*grpcXVPPb.DeleteNewsResponse, error) {
	newsId := strings.TrimSpace(req.GetNewsId())

	grpcSession, ok := ctx.Value(GRPC_CTX_KEY_SESSION).(*GrpcSession)
	if !ok {
		log.Println("DeleteNews - can not cast GrpcSession")
		return nil, status.Errorf(codes.PermissionDenied, "Invalid GrpcSession")
	}

	if len(newsId) == 0 {
		return nil, status.Errorf(codes.InvalidArgument, "Invalid Arguments")
	}

	news, err := dao.GetNewsDAO().FindByNewsId(ctx, newsId)
	if news == nil {
		return nil, status.Errorf(codes.InvalidArgument, "Invalid News: %v", req.GetNewsId())
	}

	if news.Author != grpcSession.User.Username && !grpcSession.User.IsModeratorPermission() {
		log.Println("DeleteNews - Permission Denied GrpcSession", grpcSession.User.Username)
		return nil, status.Errorf(codes.PermissionDenied, "Permission Denied")
	}

	news, err = dao.GetNewsDAO().DeleteByNewsId(ctx, newsId)
	if err != nil {
		log.Println(err)
		return nil, status.Errorf(codes.Internal, fmt.Sprintf("Internal err: %v", err))
	}

	//DELETE FILES
	go func(news *model.News) {
		err = s3Handler.GetS3FileStore().DeleteFolder(news.NewsId)
		if err != nil {
			log.Println(err)
		}

		CreateAndUpdateTagNewsCount(news.Tags, []string{})
		CreateAndUpdateParticipantNewsCount(news.Participants, []string{})

		err := dao.GetFileInfoDAO().DeleteByNewsId(context.Background(), news.NewsId)
		if err != nil {
			log.Println(err)
		}

	}(news)

	return &grpcXVPPb.DeleteNewsResponse{
		NewsId: newsId,
	}, nil
}

func (sv *XVPGRPCService) UploadNewsPreviewImage(stream grpcXVPPb.XVPService_UploadNewsPreviewImageServer) error {
	req, err := stream.Recv()
	if err != nil {
		log.Println(err)
		return err
	}

	uploadNewsPreviewImageInfo := req.GetUploadNewsPreviewImageInfo()

	newsId := strings.TrimSpace(uploadNewsPreviewImageInfo.GetNewsId())
	fileUrl := strings.TrimSpace(uploadNewsPreviewImageInfo.GetFileUrl())
	mainPreview := uploadNewsPreviewImageInfo.GetMainPreview()
	mediaType := uploadNewsPreviewImageInfo.GetMediaType()

	grpcSession, ok := stream.Context().Value(GRPC_CTX_KEY_SESSION).(*GrpcSession)
	if !ok {
		log.Println("UploadNewsPreviewImage - can not cast GrpcSession")
		return status.Errorf(codes.PermissionDenied, "Invalid GrpcSession")
	}

	//log.Println("UploadNewsPreviewImage - GrpcSession", grpcSession.User.Username)

	if len(newsId) == 0 {
		return status.Errorf(codes.InvalidArgument, "Invalid newsId")
	}

	if mediaType != grpcXVPPb.MEDIA_TYPE_IMAGE && mediaType != grpcXVPPb.MEDIA_TYPE_VIDEO {
		return status.Errorf(codes.InvalidArgument, "Invalid mediaType")
	}

	news, err := dao.GetNewsDAO().FindByNewsId(context.Background(), newsId)
	if news == nil {
		return status.Errorf(codes.InvalidArgument, "Invalid News: %v", newsId)
	}

	if news.Author != grpcSession.User.Username && !grpcSession.User.IsModeratorPermission() {
		log.Println("UploadNewsPreviewImage - Permission Denied GrpcSession", grpcSession.User.Username)
		return status.Errorf(codes.PermissionDenied, "Permission Denied")
	}

	var fileInfo *model.FileInfo
	if len(fileUrl) > 0 {
		fileInfo = &model.FileInfo{
			FileId:      utils.GenerateUUID(),
			FileUrl:     fileUrl,
			MediaType:   mediaType,
			NewsId:      newsId,
			MainPreview: mainPreview,
			CreatedAt:   utils.UTCNowMilli(),
		}
	} else {
		fileData := bytes.Buffer{}
		var fileSize int64 = 0

		for {
			//log.Println("waiting to receive more media data")

			req, err = stream.Recv()
			if err == io.EOF {
				log.Println("received file data from client")
				break
			}
			if err != nil {
				log.Println("Cannot receive chunk data", err)
				return err
			}

			chunk := req.GetChunkData()
			size := len(chunk)

			//log.Printf("received a chunk with size: %d\n", size)

			fileSize += int64(size)
			if fileSize > static.MaxMediaPreviewFileSize {
				return status.Errorf(codes.InvalidArgument, "image file is too large: %d > %d", fileSize, static.MaxMediaPreviewFileSize)
			}
			_, err = fileData.Write(chunk)
			if err != nil {
				return status.Errorf(codes.Internal, "cannot write chunk data: %v", err)
			}
		}

		checksum := fmt.Sprintf("%x", md5.Sum(fileData.Bytes()))
		if uploadNewsPreviewImageInfo.GetChecksum() != checksum {
			log.Println("Invalid Checksum", checksum, uploadNewsPreviewImageInfo.GetChecksum())
			return status.Errorf(codes.InvalidArgument, "Invalid Checksum: %v - %v", checksum, uploadNewsPreviewImageInfo.GetChecksum())
		}

		mimetype := mimetype.Detect(fileData.Bytes())
		mimetypeString := mimetype.String()
		if mediaType == grpcXVPPb.MEDIA_TYPE_IMAGE {
			if !strings.HasPrefix(mimetypeString, "image") {
				return status.Errorf(codes.InvalidArgument, "Invalid image file")
			}
		} else if mediaType == grpcXVPPb.MEDIA_TYPE_VIDEO {
			if !strings.HasPrefix(mimetypeString, "video") {
				return status.Errorf(codes.InvalidArgument, "Invalid video file")
			}
		}

		//log.Println("mimetype", mimetype.String())
		//log.Println("extension", mimetype.Extension())

		fileName := fmt.Sprintf("%s%s%s", static.S3NamePrefxix, checksum, mimetype.Extension())

		//upload to S3
		s3FileStore := s3Handler.GetS3FileStore()
		uploadOutput, err := s3FileStore.UploadFileToFolder(newsId, fileName, fileData, mimetype)
		if err != nil {
			return status.Errorf(codes.Internal, "cannot save file to the s3: %v", err)
		}

		log.Println("Done Upload File to S3... etag:", *uploadOutput.ETag)

		fileInfo = &model.FileInfo{
			FileId:      utils.GenerateUUID(),
			FileName:    fileName,
			FileSize:    fileSize,
			Checksum:    checksum,
			MediaType:   mediaType,
			NewsId:      newsId,
			MainPreview: mainPreview,
			CreatedAt:   utils.UTCNowMilli(),
		}

		//diskFileStore := fileHandler.NewDiskFileStore("media")
		//savedName, err := diskFileStore.SaveToDisk(fileInfo.FileName, mediaMsgInfo.FileExtension, fileData)
		//if err != nil {
		//	return status.Errorf(codes.Internal, "cannot save image to the store: %v", err)
		//}

		log.Printf("saved file to s3 with name: %s, size: %d, FileId: %s, checkSum: %s\n", fileName, fileSize, fileInfo.FileId, fileInfo.Checksum)
	}

	fileInfo, err = dao.GetFileInfoDAO().Save(context.Background(), fileInfo)
	if err != nil {
		return status.Errorf(codes.Internal, "cannot save fileInfo: %v", err)
	}

	news, err = dao.GetNewsDAO().AppendPreviewImage(context.Background(), newsId, fileInfo.FileId)
	if err != nil {
		return status.Errorf(codes.Internal, "cannot update News: %v", err)
	}

	newsInfo, err := GetNewsInfo(news, grpcSession.User)
	if err != nil {
		log.Println(err)
		return status.Errorf(codes.Internal, fmt.Sprintf("Internal err: %v", err))
	}

	res := &grpcXVPPb.UploadNewsPreviewImageResponse{
		NewsInfo: newsInfo,
	}

	err = stream.SendAndClose(res)
	if err != nil {
		return status.Errorf(codes.Unknown, "cannot send response: %v", err)
	}

	return nil
}

func (sv *XVPGRPCService) DeleteNewsPreviewImage(ctx context.Context, req *grpcXVPPb.DeleteNewsPreviewImageRequest) (*grpcXVPPb.DeleteNewsPreviewImageResponse, error) {
	newsId := strings.TrimSpace(req.GetNewsId())
	fileId := strings.TrimSpace(req.GetFileId())

	if len(newsId) == 0 || len(fileId) == 0 {
		return nil, status.Errorf(codes.InvalidArgument, "Invalid Arguments")
	}

	grpcSession, ok := ctx.Value(GRPC_CTX_KEY_SESSION).(*GrpcSession)
	if !ok {
		log.Println("DeleteNewsPreviewImage - can not cast GrpcSession")
		return nil, status.Errorf(codes.PermissionDenied, "Invalid GrpcSession")
	}

	if len(newsId) == 0 {
		return nil, status.Errorf(codes.InvalidArgument, "Invalid Arguments")
	}

	news, err := dao.GetNewsDAO().FindByNewsId(context.Background(), newsId)
	if news == nil {
		return nil, status.Errorf(codes.InvalidArgument, "Invalid News: %v", newsId)
	}

	if news.Author != grpcSession.User.Username && !grpcSession.User.IsModeratorPermission() {
		log.Println("DeleteNewsPreviewImage - Permission Denied GrpcSession", grpcSession.User.Username)
		return nil, status.Errorf(codes.PermissionDenied, "Permission Denied")
	}

	idx := slices.IndexFunc(news.PreviewImages, func(c string) bool { return c == fileId })
	if idx < 0 {
		return nil, status.Errorf(codes.InvalidArgument, fmt.Sprintf("Invalid fileId: %v", fileId))
	}

	news, err = dao.GetNewsDAO().RemovePreviewImage(ctx, newsId, fileId)
	if err != nil {
		log.Println(err)
		return nil, status.Errorf(codes.Internal, fmt.Sprintf("Internal err: %v", err))
	}

	//DELETE FILE
	go func(newsId string, fileId string) {
		fileInfo, err := dao.GetFileInfoDAO().FindByFileId(context.Background(), fileId)
		if err != nil {
			log.Println(err)
			return
		}

		if fileInfo.NewsId == newsId {
			_, err = dao.GetFileInfoDAO().DeleteByFileId(context.Background(), fileId)
			if err != nil {
				log.Println(err)
				return
			}

			if len(fileInfo.FileName) > 0 {
				totalFilesWithFileNameAndNewsId, err := dao.GetFileInfoDAO().CountByFileNameAndNewsId(context.Background(), fileInfo.FileName, newsId)
				if err != nil {
					log.Println(err)
					return
				}

				if totalFilesWithFileNameAndNewsId == 0 {
					_, err = s3Handler.GetS3FileStore().DeleteObjectFromFolder(newsId, fileInfo.FileName)
					if err != nil {
						log.Println(err)
					}
				}
			}
		}
	}(newsId, fileId)

	newsInfo, err := GetNewsInfo(news, grpcSession.User)
	if err != nil {
		log.Println(err)
		return nil, status.Errorf(codes.Internal, fmt.Sprintf("Internal err: %v", err))
	}

	return &grpcXVPPb.DeleteNewsPreviewImageResponse{
		NewsInfo: newsInfo,
	}, nil
}

func (sv *XVPGRPCService) UploadNewsMedia(stream grpcXVPPb.XVPService_UploadNewsMediaServer) error {
	req, err := stream.Recv()
	if err != nil {
		log.Println(err)
		return err
	}

	uploadNewsMediaInfo := req.GetUploadNewsMediaInfo()

	newsId := strings.TrimSpace(uploadNewsMediaInfo.GetNewsId())
	fileUrl := strings.TrimSpace(uploadNewsMediaInfo.GetFileUrl())
	mediaType := uploadNewsMediaInfo.GetMediaType()
	resolution := uploadNewsMediaInfo.GetResolution()
	mediaStreamType := uploadNewsMediaInfo.GetMediaStreamType()
	mediaEncKey := uploadNewsMediaInfo.GetMediaEncKey()
	onDemandMediaMainFile := uploadNewsMediaInfo.GetOnDemandMediaMainFile()
	onDemandMediaMainFileId := uploadNewsMediaInfo.GetOnDemandMediaMainFileId()
	fileName := uploadNewsMediaInfo.GetFileName()

	if onDemandMediaMainFile || mediaStreamType != grpcXVPPb.MEDIA_STREAM_TYPE_ON_DEMAND {
		onDemandMediaMainFileId = ""
	}

	grpcSession, ok := stream.Context().Value(GRPC_CTX_KEY_SESSION).(*GrpcSession)
	if !ok {
		log.Println("UploadNewsMedia - can not cast GrpcSession")
		return status.Errorf(codes.PermissionDenied, "Invalid GrpcSession")
	}

	//log.Println("UploadNewsMedia - GrpcSession", grpcSession.User.Username)

	if len(newsId) == 0 {
		return status.Errorf(codes.InvalidArgument, "Invalid Arguments")
	}

	news, err := dao.GetNewsDAO().FindByNewsId(context.Background(), newsId)
	if news == nil {
		return status.Errorf(codes.InvalidArgument, "Invalid News: %v", newsId)
	}

	if news.Author != grpcSession.User.Username && !grpcSession.User.IsModeratorPermission() {
		log.Println("UploadNewsMedia - Permission Denied GrpcSession", grpcSession.User.Username)
		return status.Errorf(codes.PermissionDenied, "Permission Denied")
	}

	if mediaStreamType == grpcXVPPb.MEDIA_STREAM_TYPE_ON_DEMAND && !onDemandMediaMainFile {
		if len(onDemandMediaMainFileId) == 0 {
			return status.Errorf(codes.InvalidArgument, fmt.Sprintf("Invalid OnDemand Media Sub File"))
		}

		idx := slices.IndexFunc(news.Medias, func(c string) bool { return c == onDemandMediaMainFileId })
		if idx < 0 {
			return status.Errorf(codes.InvalidArgument, fmt.Sprintf("Invalid OnDemand dMedia Main FileId: %v", onDemandMediaMainFileId))
		}
	}

	var fileInfo *model.FileInfo
	if len(fileUrl) > 0 {
		fileInfo = &model.FileInfo{
			FileId:                  utils.GenerateUUID(),
			FileUrl:                 fileUrl,
			MediaType:               mediaType,
			MediaStreamType:         mediaStreamType,
			OnDemandMediaMainFileId: onDemandMediaMainFileId,
			MediaEncKey:             mediaEncKey,
			Resolution:              resolution,
			NewsId:                  newsId,
			CreatedAt:               utils.UTCNowMilli(),
		}
	} else {
		fileData := bytes.Buffer{}
		var fileSize int64 = 0

		for {
			//log.Println("waiting to receive more media data")

			req, err = stream.Recv()
			if err == io.EOF {
				//log.Println("received file data from client")
				break
			}
			if err != nil {
				log.Println("Cannot receive chunk data", err)
				return err
			}

			chunk := req.GetChunkData()
			size := len(chunk)

			//log.Printf("received a chunk with size: %d\n", size)

			fileSize += int64(size)
			if fileSize > static.MaxMediaFileSize {
				return status.Errorf(codes.InvalidArgument, "media file is too large: %d > %d", fileSize, static.MaxMediaFileSize)
			}
			_, err = fileData.Write(chunk)
			if err != nil {
				return status.Errorf(codes.Internal, "cannot write chunk data: %v", err)
			}
		}

		checksum := fmt.Sprintf("%x", md5.Sum(fileData.Bytes()))
		if uploadNewsMediaInfo.GetChecksum() != checksum {
			log.Println("Invalid Checksum", checksum, uploadNewsMediaInfo.GetChecksum())
			return status.Errorf(codes.InvalidArgument, "Invalid Checksum: %v - %v", checksum, uploadNewsMediaInfo.GetChecksum())
		}

		mimetype := mimetype.Detect(fileData.Bytes())
		mimetypeString := mimetype.String()

		//log.Println("mimetype", mimetype.String())
		//log.Println("extension", mimetype.Extension())

		switch mediaType {
		case grpcXVPPb.MEDIA_TYPE_IMAGE:
			if !strings.HasPrefix(mimetypeString, "image") {
				return status.Errorf(codes.InvalidArgument, "Invalid media image file")
			}
		case grpcXVPPb.MEDIA_TYPE_VIDEO:
			if !strings.HasPrefix(mimetypeString, "video") {
				if mediaStreamType == grpcXVPPb.MEDIA_STREAM_TYPE_ON_DEMAND {
					if onDemandMediaMainFile && mimetypeString != "application/vnd.apple.mpegurl" {
						return status.Errorf(codes.InvalidArgument, "Invalid on demand media video main file")
					}
					if mimetypeString != "application/vnd.apple.mpegurl" && mimetypeString != "application/octet-stream" {
						return status.Errorf(codes.InvalidArgument, "Invalid on demand media video file")
					}
				} else {
					return status.Errorf(codes.InvalidArgument, "Invalid media video file")
				}

			}
		case grpcXVPPb.MEDIA_TYPE_AUDIO:
			//TODO - CHECK FOR AUDIO FILE
			if !strings.HasPrefix(mimetypeString, "audio") {
				return status.Errorf(codes.InvalidArgument, "Invalid audio file")
			}
		}

		if mediaStreamType != grpcXVPPb.MEDIA_STREAM_TYPE_ON_DEMAND {
			fileName = fmt.Sprintf("%s%s%s", static.S3NamePrefxix, checksum, mimetype.Extension())
		}

		//upload to S3
		s3FileStore := s3Handler.GetS3FileStore()
		_, err := s3FileStore.UploadFileToFolder(newsId, fileName, fileData, mimetype)
		if err != nil {
			return status.Errorf(codes.Internal, "cannot save file to the s3: %v", err)
		}

		//log.Println("Done Upload File to S3... etag:", *uploadOutput.ETag)

		fileInfo = &model.FileInfo{
			FileId:                  fmt.Sprintf("%v-%v", newsId, checksum),
			FileName:                fileName,
			FileSize:                fileSize,
			Checksum:                checksum,
			MediaType:               mediaType,
			MediaStreamType:         mediaStreamType,
			OnDemandMediaMainFileId: onDemandMediaMainFileId,
			MediaEncKey:             mediaEncKey,
			Resolution:              resolution,
			NewsId:                  newsId,
			CreatedAt:               utils.UTCNowMilli(),
		}

		//diskFileStore := fileHandler.NewDiskFileStore("media")
		//savedName, err := diskFileStore.SaveToDisk(fileInfo.FileName, mediaMsgInfo.FileExtension, fileData)
		//if err != nil {
		//	return status.Errorf(codes.Internal, "cannot save image to the store: %v", err)
		//}

		//log.Printf("saved file to s3 with name: %s, size: %d, FileId: %s, checkSum: %s\n", fileName, fileSize, fileInfo.FileId, fileInfo.Checksum)
	}

	fileInfo, err = dao.GetFileInfoDAO().Save(context.Background(), fileInfo)
	if err != nil {
		return status.Errorf(codes.Internal, "cannot save fileInfo: %v", err)
	}

	if mediaStreamType != grpcXVPPb.MEDIA_STREAM_TYPE_ON_DEMAND || onDemandMediaMainFile {
		news, err = dao.GetNewsDAO().AppendMedia(context.Background(), newsId, fileInfo.FileId)
		if err != nil {
			return status.Errorf(codes.Internal, "cannot update News: %v", err)
		}
	}

	newsInfo, err := GetNewsInfo(news, grpcSession.User)
	if err != nil {
		log.Println(err)
		return status.Errorf(codes.Internal, fmt.Sprintf("Internal err: %v", err))
	}

	res := &grpcXVPPb.UploadNewsMediaResponse{
		NewsInfo: newsInfo,
		FileId:   fileInfo.FileId,
	}

	err = stream.SendAndClose(res)
	if err != nil {
		return status.Errorf(codes.Unknown, "cannot send response: %v", err)
	}

	return nil
}

func (sv *XVPGRPCService) DeleteNewsMedia(ctx context.Context, req *grpcXVPPb.DeleteNewsMediaRequest) (*grpcXVPPb.DeleteNewsMediaResponse, error) {
	newsId := strings.TrimSpace(req.GetNewsId())
	fileId := strings.TrimSpace(req.GetFileId())

	if len(newsId) == 0 || len(fileId) == 0 {
		return nil, status.Errorf(codes.InvalidArgument, "Invalid Arguments")
	}

	grpcSession, ok := ctx.Value(GRPC_CTX_KEY_SESSION).(*GrpcSession)
	if !ok {
		log.Println("DeleteNewsMedia - can not cast GrpcSession")
		return nil, status.Errorf(codes.PermissionDenied, "Invalid GrpcSession")
	}

	if len(newsId) == 0 {
		return nil, status.Errorf(codes.InvalidArgument, "Invalid Arguments")
	}

	news, err := dao.GetNewsDAO().FindByNewsId(context.Background(), newsId)
	if news == nil {
		return nil, status.Errorf(codes.InvalidArgument, "Invalid News: %v", newsId)
	}

	if news.Author != grpcSession.User.Username && !grpcSession.User.IsModeratorPermission() {
		log.Println("DeleteNewsMedia - Permission Denied GrpcSession", grpcSession.User.Username)
		return nil, status.Errorf(codes.PermissionDenied, "Permission Denied")
	}

	idx := slices.IndexFunc(news.Medias, func(c string) bool { return c == fileId })
	if idx < 0 {
		return nil, status.Errorf(codes.InvalidArgument, fmt.Sprintf("Invalid fileId: %v", fileId))
	}

	news, err = dao.GetNewsDAO().RemoveMedia(ctx, newsId, fileId)
	if err != nil {
		log.Println(err)
		return nil, status.Errorf(codes.Internal, fmt.Sprintf("Internal err: %v", err))
	}

	//DELETE FILE
	go func(newsId string, fileId string) {
		fileInfo, err := dao.GetFileInfoDAO().FindByFileId(context.Background(), fileId)
		if err != nil {
			log.Println(err)
			return
		}

		if fileInfo.NewsId == newsId {
			_, err = dao.GetFileInfoDAO().DeleteByFileId(context.Background(), fileId)
			if err != nil {
				log.Println(err)
				return
			}

			onDemandFileInfos := []*model.FileInfo{}
			if fileInfo.MediaStreamType == grpcXVPPb.MEDIA_STREAM_TYPE_ON_DEMAND {
				onDemandFileInfos, err = dao.GetFileInfoDAO().FindByOnDemandMediaMainFileId(context.Background(), fileId)
				if err != nil {
					log.Println(err)
				}

				err = dao.GetFileInfoDAO().DeleteByOnDemandMediaMainFileId(context.Background(), fileId)
				if err != nil {
					log.Println(err)
				}
			}

			if len(fileInfo.FileName) > 0 {
				totalFilesWithFileNameAndNewsId, err := dao.GetFileInfoDAO().CountByFileNameAndNewsId(context.Background(), fileInfo.FileName, newsId)
				if err != nil {
					log.Println(err)
					return
				}

				if totalFilesWithFileNameAndNewsId == 0 {
					_, err = s3Handler.GetS3FileStore().DeleteObjectFromFolder(newsId, fileInfo.FileName)
					if err != nil {
						log.Println(err)
					}

					for _, onDemandFileInfo := range onDemandFileInfos {
						_, err = s3Handler.GetS3FileStore().DeleteObjectFromFolder(newsId, onDemandFileInfo.FileName)
						if err != nil {
							log.Println(err)
						}
					}
				}
			}
		}
	}(newsId, fileId)

	newsInfo, err := GetNewsInfo(news, grpcSession.User)
	if err != nil {
		log.Println(err)
		return nil, status.Errorf(codes.Internal, fmt.Sprintf("Internal err: %v", err))
	}

	return &grpcXVPPb.DeleteNewsMediaResponse{
		NewsInfo: newsInfo,
	}, nil
}

func (sv *XVPGRPCService) GetNews(ctx context.Context, req *grpcXVPPb.GetNewsRequest) (*grpcXVPPb.GetNewsResponse, error) {
	grpcSession, ok := ctx.Value(GRPC_CTX_KEY_SESSION).(*GrpcSession)
	if !ok {
		log.Println("DeleteNews - can not cast GrpcSession")
		return nil, status.Errorf(codes.PermissionDenied, "Invalid GrpcSession")
	}

	newsId := strings.TrimSpace(req.GetNewsId())

	if len(newsId) == 0 {
		return nil, status.Errorf(codes.InvalidArgument, "Invalid Arguments")
	}

	news, err := dao.GetNewsDAO().FindByNewsId(ctx, newsId)
	if news == nil {
		return nil, status.Errorf(codes.InvalidArgument, "Invalid News: %v", req.GetNewsId())
	}

	newsInfo, err := GetNewsInfo(news, grpcSession.User)
	if err != nil {
		log.Println(err)
		return nil, status.Errorf(codes.Internal, fmt.Sprintf("Internal err: %v", err))
	}

	go func(news *model.News) {
		_, err := dao.GetNewsDAO().IncreaseViews(context.Background(), news)
		if err != nil {
			log.Println(err)
		}
	}(news)

	return &grpcXVPPb.GetNewsResponse{
		NewsInfo: newsInfo,
	}, nil
}

func (sv *XVPGRPCService) GetListNews(ctx context.Context, req *grpcXVPPb.GetListNewsRequest) (*grpcXVPPb.GetListNewsResponse, error) {
	grpcSession, ok := ctx.Value(GRPC_CTX_KEY_SESSION).(*GrpcSession)
	if !ok {
		log.Println("GetListNews - can not cast GrpcSession")
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

	catIds := req.GetCatIds()
	tags := req.GetTags()
	participants := req.GetParticipants()
	searchPhrase := req.GetSearchPhrase()
	author := req.GetAuthor()

	totalItem, err := dao.GetNewsDAO().CountListNews(ctx, catIds, tags, participants, searchPhrase, author)
	if err != nil {
		log.Println(err)
		return nil, status.Errorf(codes.Internal, fmt.Sprintf("Internal err: %v", err))
	}

	sort := primitive.M{"createdAt": -1}
	listNews, err := dao.GetNewsDAO().FetchListNews(ctx, page, pageSize, catIds, tags, participants, searchPhrase, author, sort)
	if err != nil {
		log.Println(err)
		return nil, status.Errorf(codes.Internal, fmt.Sprintf("Internal err: %v", err))
	}

	if len(tags) > 0 {
		go IncreaseTagsSearchCount(tags)
	}

	if len(participants) > 0 {
		go IncreaseParticipantSearchCount(participants)
	}

	newsInfos := []*grpcXVPPb.NewsInfo{}
	for _, news := range listNews {
		newsInfo, err := GetNewsInfo(news, grpcSession.User)
		if err != nil {
			log.Println(err)
			continue
		}
		newsInfos = append(newsInfos, newsInfo)
	}

	return &grpcXVPPb.GetListNewsResponse{
		Page:      page,
		PageSize:  pageSize,
		TotalItem: totalItem,
		NewsInfos: newsInfos,
	}, nil
}

func (sv *XVPGRPCService) GetManagerListNews(ctx context.Context, req *grpcXVPPb.GetManagerListNewsRequest) (*grpcXVPPb.GetManagerListNewsResponse, error) {
	grpcSession, ok := ctx.Value(GRPC_CTX_KEY_SESSION).(*GrpcSession)
	if !ok {
		log.Println("GetListNews - can not cast GrpcSession")
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

	catIds := req.GetCatIds()
	tags := req.GetTags()
	participants := req.GetParticipants()
	searchPhrase := req.GetSearchPhrase()
	author := req.GetAuthor()

	if len(author) == 0 {
		if !grpcSession.User.IsModeratorPermission() {
			author = grpcSession.User.Username
		}
	} else {
		if author != grpcSession.User.Username && !grpcSession.User.IsModeratorPermission() {
			return nil, status.Errorf(codes.PermissionDenied, "PermissionDenied")
		}
	}

	totalItem, err := dao.GetNewsDAO().CountListNews(ctx, catIds, tags, participants, searchPhrase, author)
	if err != nil {
		log.Println(err)
		return nil, status.Errorf(codes.Internal, fmt.Sprintf("Internal err: %v", err))
	}

	sort := primitive.M{"createdAt": -1}
	listNews, err := dao.GetNewsDAO().FetchListNews(ctx, page, pageSize, catIds, tags, participants, searchPhrase, author, sort)
	if err != nil {
		log.Println(err)
		return nil, status.Errorf(codes.Internal, fmt.Sprintf("Internal err: %v", err))
	}

	if len(tags) > 0 {
		go IncreaseTagsSearchCount(tags)
	}

	if len(participants) > 0 {
		go IncreaseParticipantSearchCount(participants)
	}

	newsInfos := []*grpcXVPPb.NewsInfo{}
	for _, news := range listNews {
		newsInfo, err := GetNewsInfo(news, grpcSession.User)
		if err != nil {
			log.Println(err)
			continue
		}
		newsInfos = append(newsInfos, newsInfo)
	}

	return &grpcXVPPb.GetManagerListNewsResponse{
		Page:      page,
		PageSize:  pageSize,
		TotalItem: totalItem,
		NewsInfos: newsInfos,
	}, nil
}

func (sv *XVPGRPCService) GetListTopNews(ctx context.Context, req *grpcXVPPb.GetListTopNewsRequest) (*grpcXVPPb.GetListTopNewsResponse, error) {
	grpcSession, ok := ctx.Value(GRPC_CTX_KEY_SESSION).(*GrpcSession)
	if !ok {
		log.Println("GetListNews - can not cast GrpcSession")
		return nil, status.Errorf(codes.PermissionDenied, "Invalid GrpcSession")
	}

	limit := req.GetLimit()
	if limit <= 0 {
		limit = 10
	}
	if limit > 50 {
		limit = 50
	}

	topType := req.GetTopType()
	topTimeType := req.GetTopTimeType()

	sort := primitive.D{}
	switch topType {
	case grpcXVPPb.TOP_TYPE_VIEWS:
		switch topTimeType {
		case grpcXVPPb.TOP_TIME_TYPE_WEEK:
			sort = primitive.D{
				{"currentViewsWeek", -1},
				{"weekViews", -1},
			}
			break
		case grpcXVPPb.TOP_TIME_TYPE_MONTH:
			sort = primitive.D{
				{"currentViewsMonth", -1},
				{"monthViews", -1},
			}
			break
		default:
			sort = primitive.D{{"views", -1}}
			break
		}
		break
	case grpcXVPPb.TOP_TYPE_LIKES:
		switch topTimeType {
		case grpcXVPPb.TOP_TIME_TYPE_WEEK:
			sort = primitive.D{
				{"currentLikesWeek", -1},
				{"weekLikes", -1},
			}
			break
		case grpcXVPPb.TOP_TIME_TYPE_MONTH:
			sort = primitive.D{
				{"currentLikesMonth", -1},
				{"monthLikes", -1},
			}
			break
		default:
			sort = primitive.D{{"likes", -1}}
			break
		}
		break
	case grpcXVPPb.TOP_TYPE_RATING:
		sort = primitive.D{
			{"rating", -1},
			{"createdAt", -1},
		}
		break
	default:
		break
	}

	listNews, err := dao.GetNewsDAO().FetchListNews(ctx, 0, int64(limit), []string{}, []string{}, []string{}, "", "", sort)
	if err != nil {
		log.Println(err)
		return nil, status.Errorf(codes.Internal, fmt.Sprintf("Internal err: %v", err))
	}

	newsInfos := []*grpcXVPPb.NewsInfo{}
	for _, news := range listNews {
		newsInfo, err := GetNewsInfo(news, grpcSession.User)
		if err != nil {
			log.Println(err)
			continue
		}
		newsInfos = append(newsInfos, newsInfo)
	}

	return &grpcXVPPb.GetListTopNewsResponse{
		NewsInfos: newsInfos,
	}, nil
}

func (sv *XVPGRPCService) LikeNews(ctx context.Context, req *grpcXVPPb.LikeNewsRequest) (*grpcXVPPb.LikeNewsResponse, error) {
	grpcSession, ok := ctx.Value(GRPC_CTX_KEY_SESSION).(*GrpcSession)
	if !ok {
		log.Println("LikeNews - can not cast GrpcSession")
		return nil, status.Errorf(codes.PermissionDenied, "Invalid GrpcSession")
	}

	if grpcSession.User.Role < grpcXVPPb.USER_ROLE_NORMAL_USER {
		return nil, status.Errorf(codes.PermissionDenied, "Not Allowed Guest User")
	}

	newsId := strings.TrimSpace(req.GetNewsId())
	if len(newsId) == 0 {
		return nil, status.Errorf(codes.InvalidArgument, "Invalid Arguments")
	}

	news, err := dao.GetNewsDAO().FindByNewsId(ctx, newsId)
	if news == nil {
		return nil, status.Errorf(codes.InvalidArgument, "Invalid News: %v", req.GetNewsId())
	}

	idxLikeBy := slices.IndexFunc(news.LikedBy, func(username string) bool { return username == grpcSession.User.Username })
	if idxLikeBy >= 0 {
		return nil, status.Errorf(codes.PermissionDenied, "Already liked")
	}

	news, err = dao.GetNewsDAO().IncreaseLikes(ctx, news, grpcSession.User)
	if err != nil {
		log.Println(err)
		return nil, status.Errorf(codes.Internal, fmt.Sprintf("Internal err: %v", err))
	}

	newsInfo, err := GetNewsInfo(news, grpcSession.User)
	if err != nil {
		log.Println(err)
		return nil, status.Errorf(codes.Internal, fmt.Sprintf("Internal err: %v", err))
	}

	return &grpcXVPPb.LikeNewsResponse{
		NewsInfo: newsInfo,
	}, nil
}

func (sv *XVPGRPCService) RateNews(ctx context.Context, req *grpcXVPPb.RateNewsRequest) (*grpcXVPPb.RateNewsResponse, error) {
	grpcSession, ok := ctx.Value(GRPC_CTX_KEY_SESSION).(*GrpcSession)
	if !ok {
		log.Println("RateNews - can not cast GrpcSession")
		return nil, status.Errorf(codes.PermissionDenied, "Invalid GrpcSession")
	}

	if grpcSession.User.Role < grpcXVPPb.USER_ROLE_NORMAL_USER {
		return nil, status.Errorf(codes.PermissionDenied, "Not Allowed Guest User")
	}

	newsId := strings.TrimSpace(req.GetNewsId())
	if len(newsId) == 0 {
		return nil, status.Errorf(codes.InvalidArgument, "Invalid Arguments")
	}

	news, err := dao.GetNewsDAO().FindByNewsId(ctx, newsId)
	if news == nil {
		return nil, status.Errorf(codes.InvalidArgument, "Invalid News: %v", req.GetNewsId())
	}

	idxRatedBy := slices.IndexFunc(news.RatedBy, func(username string) bool { return username == grpcSession.User.Username })
	if idxRatedBy >= 0 {
		return nil, status.Errorf(codes.PermissionDenied, "Already Voted")
	}

	point := req.GetPoint()

	news, err = dao.GetNewsDAO().Rate(ctx, newsId, point, grpcSession.User)
	if err != nil {
		log.Println(err)
		return nil, status.Errorf(codes.Internal, fmt.Sprintf("Internal err: %v", err))
	}

	newsInfo, err := GetNewsInfo(news, grpcSession.User)
	if err != nil {
		log.Println(err)
		return nil, status.Errorf(codes.Internal, fmt.Sprintf("Internal err: %v", err))
	}

	return &grpcXVPPb.RateNewsResponse{
		NewsInfo: newsInfo,
	}, nil
}

func GetNewsInfo(news *model.News, user *model.User) (*grpcXVPPb.NewsInfo, error) {
	previewImageInfos := []*grpcXVPPb.PreviewImageInfo{}
	for _, fileId := range news.PreviewImages {
		fileInfo, err := dao.GetFileInfoDAO().FindByFileId(context.Background(), fileId)
		if err != nil {
			log.Println(err)
			continue
		}
		previewImageInfos = append(previewImageInfos, &grpcXVPPb.PreviewImageInfo{
			FileId:      fileInfo.FileId,
			FileName:    fileInfo.FileName,
			FileUrl:     fileInfo.FileUrl,
			MainPreview: fileInfo.MainPreview,
			Checksum:    fileInfo.Checksum,
			MediaType:   fileInfo.MediaType,
		})

	}

	mediaInfos := []*grpcXVPPb.MediaInfo{}
	for _, fileId := range news.Medias {
		fileInfo, err := dao.GetFileInfoDAO().FindByFileId(context.Background(), fileId)
		if err != nil {
			log.Println(err)
			continue
		}
		mediaInfos = append(mediaInfos, &grpcXVPPb.MediaInfo{
			FileId:          fileInfo.FileId,
			FileName:        fileInfo.FileName,
			FileUrl:         fileInfo.FileUrl,
			MediaType:       fileInfo.MediaType,
			MediaStreamType: fileInfo.MediaStreamType,
			Resolution:      &fileInfo.Resolution,
			Checksum:        fileInfo.Checksum,
		})
	}

	idxLikeBy := slices.IndexFunc(news.LikedBy, func(username string) bool { return username == user.Username })
	idxRatedBy := slices.IndexFunc(news.RatedBy, func(username string) bool { return username == user.Username })

	mediaEncKey := news.MediaEncKey
	mediaEncIV := news.MediaEncIV
	if news.Author != user.Username && !user.IsModeratorPermission() {
		mediaEncKey = ""
		mediaEncIV = ""
	}

	return &grpcXVPPb.NewsInfo{
		NewsId:                  news.NewsId,
		Author:                  news.Author,
		Title:                   news.Title,
		Participants:            news.Participants,
		Description:             news.Description,
		CatIds:                  news.Categories,
		Tags:                    news.Tags,
		EnableComment:           news.EnableComment,
		PreviewImageInfos:       previewImageInfos,
		MediaInfos:              mediaInfos,
		MediaEncKey:             mediaEncKey,
		MediaEncIV:              mediaEncIV,
		View:                    news.Views,
		Likes:                   news.Likes,
		LikedByRequestedSession: idxLikeBy >= 0,
		Rating:                  news.Rating,
		RatedByRequestedSession: idxRatedBy >= 0,
		Status:                  news.Status,
		WeekViews:               news.WeekViews,
		MonthViews:              news.MonthViews,
		CurrentViewsWeek:        news.CurrentViewsWeek,
		CurrentViewsMonth:       news.CurrentViewsMonth,
		WeekLikes:               news.WeekLikes,
		MonthLikes:              news.MonthLikes,
		CurrentLikesWeek:        news.CurrentLikesWeek,
		CurrentLikesMonth:       news.CurrentLikesMonth,
	}, nil
}
