package appgrpc

import (
	"context"
	"fmt"
	"gbb.go/gvp/dao"
	"gbb.go/gvp/proto/grpcXVPPb"
	"gbb.go/gvp/s3Handler"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"io"
	"log"
	"strings"
	"time"
)

func (sv *XVPGRPCService) DownloadFile(req *grpcXVPPb.DownloadFileRequest, stream grpcXVPPb.XVPService_DownloadFileServer) error {
	fileId := strings.TrimSpace(req.GetFileId())

	fileInfo, err := dao.GetFileInfoDAO().FindByFileId(context.Background(), fileId)
	if err != nil {
		log.Println(err)
		return status.Errorf(codes.PermissionDenied, fmt.Sprintf("Not Found FileInfo"))
	}

	if len(fileInfo.FileUrl) > 0 {
		return status.Errorf(codes.PermissionDenied, fmt.Sprintf("Invalid FileInfo download destination"))
	}

	res := &grpcXVPPb.DownloadFileResponse{
		Data: &grpcXVPPb.DownloadFileResponse_FileInfo{
			FileInfo: &grpcXVPPb.FileInfo{
				FileId:     fileInfo.FileId,
				FileName:   fileInfo.FileName,
				MediaType:  fileInfo.MediaType,
				Resolution: &fileInfo.Resolution,
				Checksum:   fileInfo.Checksum,
			},
		},
	}
	err = stream.Send(res)
	if err != nil {
		log.Println("cannot send FileInfo to client: ", err)
		return err
	}

	s3Object, err := s3Handler.GetS3FileStore().DownloadObject(fileInfo.FileName)
	if err != nil {
		log.Println("Get s3 Object err", err)
		return status.Errorf(codes.NotFound, fmt.Sprintf("Not found S3 object"))
	}

	if s3Object.Body != nil {
		defer s3Object.Body.Close()
	}
	// If there is no content length, it is a directory
	if s3Object.ContentLength == nil {
		return nil
	}

	//log.Println("DownloadFile:", *s3Object.ETag, *s3Object.ContentLength)

	buffer := make([]byte, 1024)
	for {
		n, errRead := s3Object.Body.Read(buffer)

		if errRead != nil && errRead != io.EOF {
			return status.Errorf(codes.Internal, fmt.Sprintf("Cannot read S3 object"))
		}

		if n > 0 {
			res := &grpcXVPPb.DownloadFileResponse{
				Data: &grpcXVPPb.DownloadFileResponse_ChunkData{
					ChunkData: buffer[:n],
				},
			}
			err = stream.Send(res)
			if err != nil {
				//To get the real error that contains the gRPC status code, we must call stream.RecvMsg() with a nil parameter. The nil parameter basically means that we don't expect to receive any message, but we just want to get the error that function returns
				log.Println("cannot send chunk to client: ", err)
				return status.Errorf(codes.Internal, fmt.Sprintf("Cannot send chunk to client"))
			}
		}

		if errRead == io.EOF {
			//log.Println("DownloadMediaMsg - reach EOF")
			break
		}
	}

	log.Println("DownloadFile - Done")
	return nil

}

func (sv *XVPGRPCService) GetFilePresignedUrl(ctx context.Context, req *grpcXVPPb.GetFilePresignedUrlRequest) (*grpcXVPPb.GetFilePresignedUrlResponse, error) {
	fileId := strings.TrimSpace(req.GetFileId())
	if len(fileId) == 0 {
		return nil, status.Errorf(codes.InvalidArgument, "Invalid Arguments")
	}

	fileInfo, _ := dao.GetFileInfoDAO().FindByFileId(ctx, fileId)
	if fileInfo == nil {
		return nil, status.Errorf(codes.InvalidArgument, "Invalid Arguments")
	}

	url, err := s3Handler.GetS3FileStore().GenObjectPresignedUrlFromUrl(fileInfo.NewsId, fileInfo.FileName, 6*time.Hour)
	if err != nil {
		log.Println(err)
		return nil, status.Errorf(codes.NotFound, "Not Dound")
	}

	//log.Println(url)

	return &grpcXVPPb.GetFilePresignedUrlResponse{
		Url: url,
	}, nil
}

func (sv *XVPGRPCService) GetFileInfo(ctx context.Context, req *grpcXVPPb.GetFileInfolRequest) (*grpcXVPPb.GetFileInfolResponse, error) {
	fileId := strings.TrimSpace(req.GetFileId())
	if len(fileId) == 0 {
		return nil, status.Errorf(codes.InvalidArgument, "Invalid Arguments")
	}

	fileInfo, _ := dao.GetFileInfoDAO().FindByFileId(ctx, fileId)
	if fileInfo == nil {
		return nil, status.Errorf(codes.InvalidArgument, "Invalid Arguments")
	}

	return &grpcXVPPb.GetFileInfolResponse{
		FileInfo: &grpcXVPPb.FileInfo{
			FileId:          fileInfo.FileId,
			FileName:        fileInfo.FileName,
			MediaType:       fileInfo.MediaType,
			MediaStreamType: fileInfo.MediaStreamType,
			Resolution:      &fileInfo.Resolution,
			Checksum:        fileInfo.Checksum,
		},
	}, nil
}
