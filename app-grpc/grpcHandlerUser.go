package appgrpc

import (
	"bytes"
	"context"
	"crypto/md5"
	"errors"
	"fmt"
	"gbb.go/gvp/dao"
	"gbb.go/gvp/model"
	"gbb.go/gvp/proto/grpcXVPPb"
	"gbb.go/gvp/s3Handler"
	"gbb.go/gvp/static"
	"gbb.go/gvp/utils"
	"github.com/gabriel-vasile/mimetype"
	"github.com/golang-jwt/jwt/v4"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"io"
	"log"
	"strings"
	"time"
)

func (sv *XVPGRPCService) CreatUser(ctx context.Context, req *grpcXVPPb.CreatAccountRequest) (*grpcXVPPb.CreatAccountResponse, error) {
	userName := strings.TrimSpace(req.GetUsername())
	email := strings.TrimSpace(req.GetEmail())
	passwordRaw := strings.TrimSpace(req.GetPassword())

	if len(userName) == 0 || len(email) == 0 || len(passwordRaw) == 0 {
		return nil, status.Errorf(codes.InvalidArgument, "Invalid argument")
	}

	if !utils.IsValidEmail(email) {
		return nil, status.Errorf(codes.InvalidArgument, "Invalid email")
	}

	if ctx.Err() == context.DeadlineExceeded {
		return nil, status.Error(codes.Canceled, "the client canceled the request")
	}

	userWithSameEmail, err := dao.GetUserDAO().FindByEmail(ctx, email)
	if err == nil && userWithSameEmail != nil {
		return nil, status.Errorf(codes.AlreadyExists, "Email existed")
	}

	userWithSameUsername, err := dao.GetUserDAO().FindByUserName(ctx, userName)
	if err == nil && userWithSameUsername != nil {
		return nil, status.Errorf(codes.AlreadyExists, "Username existed")
	}

	emailLockReg, err := dao.GetEmailLockRegDAO().FindByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) { //phoneLockReg == nil
			emailLockReg = &model.EmailLockReg{
				Email:                  email,
				NumAuthencodeSend:      0,
				LastDateAuthencodeSend: 0,
				Locked:                 false,
			}
		} else {
			log.Println(err)
			return nil, status.Errorf(codes.Internal, fmt.Sprintf("Internal err: %v", err))
		}
	}

	authenCode := utils.GenerateAuthenCode()
	log.Println("authenCode", authenCode)
	now := utils.UTCNow()

	userTmp, err := dao.GetUserTmpDAO().FindByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) { //userTmp == nil
			userTmp = &model.UserTmp{
				Email:     email,
				CreatedAt: now.UnixMilli(),
			}
		} else {
			log.Println(err)
			return nil, status.Errorf(codes.Internal, fmt.Sprintf("Internal err: %v", err))
		}
	}
	password, err := utils.HashPassword(passwordRaw)
	if err != nil {
		log.Println(err)
		return nil, status.Errorf(codes.Internal, fmt.Sprintf("Internal err: %v", err))
	}

	userTmp.Username = userName
	userTmp.Password = password
	userTmp.Authencode = authenCode
	userTmp.AuthencodeSendAt = now.UnixMilli()
	userTmp.NumberAuthenFail = 0

	if emailLockReg.NumAuthencodeSend > static.MaxResendAuthenCode {
		emailLockReg.Locked = true
	}

	if emailLockReg.Locked {
		lastDateAuthencodeSend := time.UnixMilli(emailLockReg.LastDateAuthencodeSend)
		diffMins := now.Sub(lastDateAuthencodeSend).Minutes()
		//fmt.Println("diffMins", diffMins)
		if diffMins >= static.TimeResetEmailCreateAccountLock {
			emailLockReg.Locked = false
			emailLockReg.NumAuthencodeSend = 0
		} else {
			log.Println("Over Max Resend AuthenCode", email)
			return nil, status.Errorf(codes.Aborted, "Over Max Resend AuthenCode")
		}
	}
	emailLockReg.NumAuthencodeSend++
	emailLockReg.LastDateAuthencodeSend = now.UnixMilli()

	userTmp, err = dao.GetUserTmpDAO().Save(ctx, userTmp)
	if err != nil {
		return nil, status.Errorf(
			codes.Internal,
			fmt.Sprintf("Internal err: %v", err),
		)
	}

	emailLockReg, err = dao.GetEmailLockRegDAO().Save(ctx, emailLockReg)
	if err != nil {
		return nil, status.Errorf(
			codes.Internal,
			fmt.Sprintf("Internal err: %v", err),
		)
	}

	go func(userTmp *model.UserTmp) {
		err := sv.SendEmailAuthenCode(userTmp)
		if err != nil {
			log.Println("Cannot send SMS", err)
		}
	}(userTmp)

	creatAccountResponse := grpcXVPPb.CreatAccountResponse{
		Username:          userTmp.Username,
		Email:             userTmp.Email,
		AuthenCodeTimeOut: static.AuthencodeTimeOut,
	}

	return &creatAccountResponse, nil
}

func (sv *XVPGRPCService) VerifyAuthencode(ctx context.Context, req *grpcXVPPb.VerifyAuthencodeRequest) (*grpcXVPPb.VerifyAuthencodeResponse, error) {
	if len(req.GetUsername()) == 0 || len(req.GetEmail()) == 0 || len(req.GetAuthencode()) == 0 {
		return nil, status.Errorf(codes.InvalidArgument, "Invalid Arguments")
	}

	userTmp, _ := dao.GetUserTmpDAO().FindByEmail(ctx, req.GetEmail())
	if userTmp == nil {
		return nil, status.Errorf(codes.Aborted, "Not found UserTmp")
	}

	if ctx.Err() == context.DeadlineExceeded {
		return nil, status.Error(codes.Canceled, "the client canceled the request")
	}

	if userTmp.Username != req.GetUsername() {
		return nil, status.Errorf(codes.InvalidArgument, "Invalid Username")
	}

	now := utils.UTCNow()
	authencodeSendAt := time.UnixMilli(userTmp.AuthencodeSendAt)
	if now.Sub(authencodeSendAt).Seconds() > static.AuthencodeTimeOut {
		return nil, status.Errorf(codes.DeadlineExceeded, "Authencode TimeOut")
	}

	if userTmp.Authencode != req.GetAuthencode() {
		userTmp.NumberAuthenFail++
		if userTmp.NumberAuthenFail > static.MaxNumberAuthencodeFail {
			_, err := dao.GetUserTmpDAO().DeleteByEmail(ctx, userTmp.Email)
			if err != nil {
				log.Println(err)
				return nil, status.Errorf(codes.Internal, fmt.Sprintf("Internal err: %v", err))
			}
			return nil, status.Error(codes.Aborted, "Max Number Authencode Fail")
		}
		_, err := dao.GetUserTmpDAO().Save(ctx, userTmp)
		if err != nil {
			log.Println(err)
			return nil, status.Errorf(codes.Internal, fmt.Sprintf("Internal err: %v", err))
		}
		return nil, status.Error(codes.Aborted, "Invalid Authencode")
	}

	userWithSameEmail, err := dao.GetUserDAO().FindByEmail(ctx, userTmp.Email)
	if err == nil && userWithSameEmail != nil {
		return nil, status.Errorf(codes.AlreadyExists, "Email existed")
	}

	userWithSameUsername, err := dao.GetUserDAO().FindByUserName(ctx, userTmp.Username)
	if err == nil && userWithSameUsername != nil {
		return nil, status.Errorf(codes.AlreadyExists, "Username existed")
	}

	user := &model.User{
		Username:  userTmp.Username,
		Email:     userTmp.Email,
		Role:      grpcXVPPb.USER_ROLE_NORMAL_USER,
		Password:  userTmp.Password,
		CreatedAt: now.UnixMilli(),
	}

	_, err = dao.GetUserTmpDAO().DeleteByEmail(ctx, userTmp.Email)
	if err != nil {
		log.Println(err)
		return nil, status.Errorf(codes.Internal, fmt.Sprintf("Internal err: %v", err))
	}

	user, err = dao.GetUserDAO().Save(ctx, user)
	if err != nil {
		log.Println(err)
		return nil, status.Errorf(codes.Internal, fmt.Sprintf("Internal err: %v", err))
	}

	verifyAuthencodeResponse := grpcXVPPb.VerifyAuthencodeResponse{
		Username: user.Username,
		Email:    user.Email,
	}

	return &verifyAuthencodeResponse, nil
}

func (sv *XVPGRPCService) Login(ctx context.Context, req *grpcXVPPb.LoginRequest) (*grpcXVPPb.LoginResponse, error) {

	if len(req.GetUsername()) == 0 || len(req.GetPassword()) == 0 {
		return nil, status.Errorf(codes.InvalidArgument, "Invalid Arguments")
	}

	//log.Println("Call Login", req.GetUsername())

	user, err := dao.GetUserDAO().FindByUserName(context.Background(), req.GetUsername())
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) { //NotFound user
			return nil, status.Errorf(codes.NotFound, "NotFound user")
		} else {
			return nil, status.Errorf(codes.Internal, fmt.Sprintf("Internal err: %v", err))
		}
	}

	isValidPassword := utils.CheckPasswordHash(req.GetPassword(), user.Password)
	if !isValidPassword {
		return nil, status.Errorf(codes.Unauthenticated, "Invalid credential")
	}

	expirationTime := jwt.NumericDate{Time: time.Now().Add(static.JWTTTL * time.Minute)}
	// Create the JWT claims, which includes the username and expiry time
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{
		ExpiresAt: &expirationTime,
		Subject:   user.Username,
	})

	responseJWT, err := token.SignedString(static.JWTKey())
	if err != nil {
		return nil, status.Errorf(codes.Internal, fmt.Sprintf("Internal err: %v", err))
	}

	loginResponse := &grpcXVPPb.LoginResponse{
		Jwt:    responseJWT,
		JwtTTL: static.JWTTTL,
	}

	return loginResponse, nil
}

func (sv *XVPGRPCService) UpdateProfile(ctx context.Context, req *grpcXVPPb.UpdateProfileRequest) (*grpcXVPPb.UpdateProfileResponse, error) {
	grpcSession, ok := ctx.Value(GRPC_CTX_KEY_SESSION).(*GrpcSession)
	if !ok {
		log.Println("UpdateProfile - can not cast GrpcSession")
		return nil, status.Errorf(codes.PermissionDenied, "Invalid GrpcSession")
	}

	//log.Println("UpdateProfile", req.GetFirstName(), req.GetLastName())

	firstName := strings.TrimSpace(req.GetFirstName())
	lastName := strings.TrimSpace(req.GetLastName())
	phoneNumber := strings.TrimSpace(req.GetPhoneNumber())
	gender := req.GetGender()
	birthday := req.GetBirthday()
	//log.Println("UpdateProfile", firstName, lastName)

	updateFields := primitive.M{}
	if len(firstName) > 0 {
		updateFields["firstName"] = firstName
	}

	if len(lastName) > 0 {
		updateFields["lastName"] = lastName
	}

	if len(phoneNumber) > 0 {
		updateFields["phoneNumber"] = phoneNumber
	}

	if birthday > 0 {
		updateFields["birthday"] = birthday
	}

	log.Println("gender", gender)
	if gender == grpcXVPPb.USER_GENDER_MALE || gender == grpcXVPPb.USER_GENDER_FEMALE {
		updateFields["gender"] = gender
	}

	update := primitive.M{"$set": updateFields}

	user, err := dao.GetUserDAO().UpdateByUserName(ctx, grpcSession.User.Username, update, []interface{}{}, false)
	if err != nil {
		return nil, status.Errorf(codes.Internal, fmt.Sprintf("Internal err: %v", err))
	}

	userInfo, err := GetUserInfo(user, grpcSession.User)
	if err != nil {
		log.Println(err)
		return nil, status.Errorf(codes.Internal, fmt.Sprintf("Internal err: %v", err))
	}

	return &grpcXVPPb.UpdateProfileResponse{
		UserInfo: userInfo,
	}, nil
}

func (sv *XVPGRPCService) UploadUserAvatar(stream grpcXVPPb.XVPService_UploadUserAvatarServer) error {
	req, err := stream.Recv()
	if err != nil {
		log.Println(err)
		return err
	}

	uploadUserAvatarInfo := req.GetUploadUserAvatarInfo()
	fileUrl := strings.TrimSpace(uploadUserAvatarInfo.GetFileUrl())

	grpcSession, ok := stream.Context().Value(GRPC_CTX_KEY_SESSION).(*GrpcSession)
	if !ok {
		log.Println("UploadNewsPreviewImage - can not cast GrpcSession")
		return status.Errorf(codes.PermissionDenied, "Invalid GrpcSession")
	}

	currentAvatar := grpcSession.User.Avatar
	var currentAvatarFileInfo *model.FileInfo
	if len(currentAvatar) > 0 {
		currentAvatarFileInfo, _ = dao.GetFileInfoDAO().FindByFileId(context.Background(), currentAvatar)
	}

	var fileInfo *model.FileInfo
	if len(fileUrl) > 0 {
		fileInfo = &model.FileInfo{
			FileId:    utils.GenerateUUID(),
			FileUrl:   fileUrl,
			MediaType: grpcXVPPb.MEDIA_TYPE_IMAGE,
			NewsId:    static.NewsID_Avatar,
			CreatedAt: utils.UTCNowMilli(),
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
		if uploadUserAvatarInfo.GetChecksum() != checksum {
			log.Println("Invalid Checksum", checksum, uploadUserAvatarInfo.GetChecksum())
			return status.Errorf(codes.InvalidArgument, "Invalid Checksum: %v - %v", checksum, uploadUserAvatarInfo.GetChecksum())
		}

		mimetype := mimetype.Detect(fileData.Bytes())
		mimetypeString := mimetype.String()

		if !strings.HasPrefix(mimetypeString, "image") {
			return status.Errorf(codes.InvalidArgument, "Invalid media image file")
		}

		fileName := fmt.Sprintf("%s%s%s", static.S3NamePrefxix, checksum, mimetype.Extension())

		//upload to S3
		s3FileStore := s3Handler.GetS3FileStore()
		_, err := s3FileStore.UploadFileToFolder(static.NewsID_Avatar, fileName, fileData, mimetype)
		if err != nil {
			return status.Errorf(codes.Internal, "cannot save file to the s3: %v", err)
		}

		//log.Println("Done Upload File to S3... etag:", *uploadOutput.ETag)
		fileInfo = &model.FileInfo{
			FileId:    fmt.Sprintf("%v-%v", static.NewsID_Avatar, checksum),
			FileName:  fileName,
			FileSize:  fileSize,
			Checksum:  checksum,
			MediaType: grpcXVPPb.MEDIA_TYPE_IMAGE,
			NewsId:    static.NewsID_Avatar,
			CreatedAt: utils.UTCNowMilli(),
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

	updateFields := primitive.M{}
	updateFields["avatar"] = fileInfo.FileId
	update := primitive.M{"$set": updateFields}

	user, err := dao.GetUserDAO().UpdateByUserName(context.Background(), grpcSession.User.Username, update, []interface{}{}, false)
	if err != nil {
		return status.Errorf(codes.Internal, fmt.Sprintf("cannot save user: %v", err))
	}

	userInfo, err := GetUserInfo(user, grpcSession.User)
	if err != nil {
		log.Println(err)
		return status.Errorf(codes.Internal, fmt.Sprintf("Internal err: %v", err))
	}

	if currentAvatarFileInfo != nil {
		go func(fileId string) {
			fileInfo, err := dao.GetFileInfoDAO().FindByFileId(context.Background(), fileId)
			if err != nil {
				log.Println(err)
				return
			}

			if fileInfo.NewsId == static.NewsID_Avatar {
				_, err = dao.GetFileInfoDAO().DeleteByFileId(context.Background(), fileId)
				if err != nil {
					log.Println(err)
					return
				}

				if len(fileInfo.FileName) > 0 {
					totalFilesWithFileNameAndNewsId, err := dao.GetFileInfoDAO().CountByFileNameAndNewsId(context.Background(), fileInfo.FileName, static.NewsID_Avatar)
					if err != nil {
						log.Println(err)
						return
					}

					if totalFilesWithFileNameAndNewsId == 0 {
						_, err = s3Handler.GetS3FileStore().DeleteObjectFromFolder(static.NewsID_Avatar, fileInfo.FileName)
						if err != nil {
							log.Println(err)
						}
					}
				}
			}
		}(currentAvatarFileInfo.FileId)
	}

	res := &grpcXVPPb.UploadUserAvatarResponse{
		UserInfo: userInfo,
	}

	err = stream.SendAndClose(res)
	if err != nil {
		return status.Errorf(codes.Unknown, "cannot send response: %v", err)
	}

	return nil
}

func (sv *XVPGRPCService) GetUserInfo(ctx context.Context, req *grpcXVPPb.GetUserInfoRequest) (*grpcXVPPb.GetUserInfoResponse, error) {
	grpcSession, ok := ctx.Value(GRPC_CTX_KEY_SESSION).(*GrpcSession)
	if !ok {
		log.Println("UpdateProfile - can not cast GrpcSession")
		return nil, status.Errorf(codes.PermissionDenied, "Invalid GrpcSession")
	}

	userName := strings.TrimSpace(req.GetUsername())
	if len(userName) == 0 {
		return nil, status.Errorf(codes.InvalidArgument, "Invalid Arguments")
	}

	user, err := dao.GetUserDAO().FindByUserName(context.Background(), userName)
	if user == nil {
		return nil, status.Errorf(codes.NotFound, "Not found user: %v", userName)
	}

	userInfo, err := GetUserInfo(user, grpcSession.User)
	if err != nil {
		log.Println(err)
		return nil, status.Errorf(codes.Internal, fmt.Sprintf("Internal err: %v", err))
	}

	return &grpcXVPPb.GetUserInfoResponse{
		UserInfo: userInfo,
	}, nil
}

func GetUserInfo(user *model.User, userSession *model.User) (*grpcXVPPb.UserInfo, error) {
	avatarFileId := ""
	avatarFileUrl := ""

	if len(user.Avatar) > 0 {
		fileInfo, err := dao.GetFileInfoDAO().FindByFileId(context.Background(), user.Avatar)
		if err != nil {
			log.Println(err)
			return nil, err
		}

		avatarFileId = fileInfo.FileId
		avatarFileUrl = fileInfo.FileUrl
	}

	userAvatarInfo := &grpcXVPPb.UserAvatarInfo{
		FileId:  avatarFileId,
		FileUrl: avatarFileUrl,
	}

	role := grpcXVPPb.USER_ROLE_GUEST
	phoneNumber := ""
	var birthday int64 = 0
	if userSession.Username == user.Username || userSession.IsModeratorPermission() {
		role = user.Role
		phoneNumber = user.PhoneNumber
		birthday = user.Birthday
	}

	return &grpcXVPPb.UserInfo{
		Username:       user.Username,
		Email:          user.Email,
		Role:           role,
		FirstName:      user.FirstName,
		LastName:       user.LastName,
		UserAvatarInfo: userAvatarInfo,
		Gender:         user.Gender,
		PhoneNumber:    phoneNumber,
		Birthday:       birthday,
	}, nil
}
