package apphttp

import (
	"bufio"
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"gbb.go/gvp/dao"
	"gbb.go/gvp/model"
	"gbb.go/gvp/proto/grpcXVPPb"
	"gbb.go/gvp/s3Handler"
	"gbb.go/gvp/static"
	"gbb.go/gvp/utils"
	"github.com/gabriel-vasile/mimetype"
	"github.com/gin-gonic/gin"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const (
	FILE_ID = "fileId"
)
const DOWNLOADS_PATH = "./downloads/"

type NewsController struct {
}

func (sv *NewsController) GetPresignedUrl(ctx *gin.Context) {
	fileId := strings.TrimSpace(ctx.Param("fileId"))

	fileInfo, _ := dao.GetFileInfoDAO().FindByFileId(ctx, fileId)
	if fileInfo == nil {
		ctx.Writer.WriteHeader(http.StatusBadRequest)
		return
	}

	url, err := s3Handler.GetS3FileStore().GenObjectPresignedUrlFromUrl(fileInfo.NewsId, fileInfo.FileName, 6*time.Hour)
	if err != nil {
		log.Println(err)
		ctx.Writer.WriteHeader(http.StatusBadRequest)
		return
	}

	//session := sessions.Default(ctx)
	//session.Set("mytoken", "xxx")
	//session.Save()

	ctx.String(http.StatusOK, url)
}

func (sv *NewsController) GetKeyV2(ctx *gin.Context) {
	newsId := strings.TrimSpace(ctx.Param("newsId"))

	news, _ := dao.GetNewsDAO().FindByNewsId(ctx, newsId)
	if news == nil {
		ctx.Writer.WriteHeader(http.StatusBadRequest)
		return
	}

	data, err := hex.DecodeString(news.MediaEncKey)
	if err != nil {
		fmt.Println(err)
	}

	//log.Println("call GetKeyV2")
	//data, err := os.ReadFile("./downloads/enc.key")
	//if err != nil {
	//	fmt.Println(err)
	//}

	//log.Println("data", len(data))
	//log.Println("data", hex.EncodeToString(data))

	ctx.Header("Content-Description", "File Transfer")
	ctx.Header("Content-Transfer-Encoding", "binary")
	ctx.Header("Content-Disposition", "attachment; filename=enc.key")
	ctx.Data(http.StatusOK, "application/octet-stream", data)

}

func (sv *NewsController) GetKey(ctx *gin.Context) {
	fileId := strings.TrimSpace(ctx.GetHeader(FILE_ID))

	if len(fileId) == 0 {
		log.Println("Invalid fileId", fileId)
		ctx.Writer.WriteHeader(http.StatusBadRequest)
		return
	}

	//userSession := ctx.Value(CTX_KEY_USER).(*model.User)
	//log.Printf("GetKey - userSession %v\n", userSession.Username)

	fileInfo, _ := dao.GetFileInfoDAO().FindByFileId(ctx, fileId)
	if fileInfo == nil {
		ctx.Writer.WriteHeader(http.StatusBadRequest)
		return
	}

	data, err := hex.DecodeString(fileInfo.MediaEncKey)
	if err != nil {
		fmt.Println(err)
	}

	//log.Println("data", len(data))
	//log.Println("data", hex.EncodeToString(data))

	ctx.Header("Content-Description", "File Transfer")
	ctx.Header("Content-Transfer-Encoding", "binary")
	ctx.Header("Content-Disposition", "attachment; filename=enc.key")
	ctx.Data(http.StatusOK, "application/octet-stream", data)
}

func (sv *NewsController) UploadSinglePreviewImage(ctx *gin.Context) {
	// single file
	userSession := ctx.Value(CTX_KEY_USER).(*model.User)

	newsId, existed := ctx.GetPostForm("newsId")
	if !existed {
		ctx.String(http.StatusForbidden, "Invalid form")
		return
	}
	log.Println("newsId", newsId)

	fileUrl, _ := ctx.GetPostForm("fileUrl")
	log.Println("fileUrl", fileUrl)

	mainPreview, existed := ctx.GetPostForm("mainPreview")
	if !existed {
		ctx.String(http.StatusForbidden, "Invalid form")
		return
	}
	mainPreviewBool, err := strconv.ParseBool(mainPreview)
	if err != nil {
		ctx.String(http.StatusForbidden, "Invalid form")
		return
	}
	log.Println("mainPreview", mainPreviewBool)

	news, err := dao.GetNewsDAO().FindByNewsId(ctx, newsId)
	if news == nil {
		ctx.String(http.StatusForbidden, "Invalid News: %v", newsId)
		return
	}
	if news.Author != userSession.Username && !userSession.IsModeratorPermission() {
		ctx.String(http.StatusForbidden, "Permission Denied")
		return
	}

	var fileInfo *model.FileInfo
	if len(fileUrl) > 0 {
		fileInfo = &model.FileInfo{
			FileId:      utils.GenerateUUID(),
			FileUrl:     fileUrl,
			MediaType:   grpcXVPPb.MEDIA_TYPE_IMAGE,
			NewsId:      newsId,
			MainPreview: mainPreviewBool,
			CreatedAt:   time.Now().UnixMilli(),
		}
	} else {
		fileUpload, err := ctx.FormFile("file")
		if err != nil {
			log.Println(err)
			ctx.String(http.StatusForbidden, "Invalid form")
			return
		}

		file, err := fileUpload.Open()
		if err != nil {
			log.Println(err)
			ctx.String(http.StatusForbidden, fmt.Sprintf("error: %v", err))
			return
		}

		defer file.Close()

		file.Seek(0, io.SeekStart)
		reader := bufio.NewReader(file)
		buffer := make([]byte, 1024)

		fileData := bytes.Buffer{}
		var fileSize int64 = 0

		for {
			n, err := reader.Read(buffer)
			if err == io.EOF {
				break
			}
			if err != nil {
				ctx.String(http.StatusForbidden, fmt.Sprintf("error: %v", err))
				return
			}

			fileSize += int64(n)
			if fileSize > static.MaxImageFileSize {
				ctx.String(http.StatusForbidden, fmt.Sprintf("image file is too large: %d > %d", fileSize, static.MaxImageFileSize))
				return
			}
			_, err = fileData.Write(buffer[:n])

		}

		checksum := fmt.Sprintf("%x", md5.Sum(fileData.Bytes()))

		mimetype := mimetype.Detect(fileData.Bytes())
		mimetypeString := mimetype.String()
		if !strings.HasPrefix(mimetypeString, "image") {
			ctx.String(http.StatusForbidden, "Invalid image file")
			return
		}

		//log.Println("mimetype", mimetype.String())
		//log.Println("extension", mimetype.Extension())

		fileName := fmt.Sprintf("%s%s%s", static.S3NamePrefxix, checksum, mimetype.Extension())

		//upload to S3
		s3FileStore := s3Handler.GetS3FileStore()
		uploadOutput, err := s3FileStore.UploadFileToFolder(newsId, fileName, fileData, mimetype)
		if err != nil {
			ctx.String(http.StatusForbidden, fmt.Sprintf("cannot save file to the s3: %v", err))
			return
		}

		log.Println("Done Upload File to S3... etag:", *uploadOutput.ETag)

		fileInfo = &model.FileInfo{
			FileId:      utils.GenerateUUID(),
			FileName:    fileName,
			FileSize:    fileSize,
			Checksum:    checksum,
			MediaType:   grpcXVPPb.MEDIA_TYPE_IMAGE,
			NewsId:      newsId,
			MainPreview: mainPreviewBool,
			CreatedAt:   time.Now().UnixMilli(),
		}

		//diskFileStore := fileHandler.NewDiskFileStore("media")
		//savedName, err := diskFileStore.SaveToDisk(fileInfo.FileName, mediaMsgInfo.FileExtension, fileData)
		//if err != nil {
		//	return status.Errorf(codes.Internal, "cannot save image to the store: %v", err)
		//}

		log.Printf("saved file to s3 with name: %s, size: %d, FileId: %s, checkSum: %s\n", fileName, fileSize, fileInfo.FileId, fileInfo.Checksum)
	}

	fileInfo, err = dao.GetFileInfoDAO().Save(ctx, fileInfo)
	if err != nil {
		ctx.String(http.StatusForbidden, fmt.Sprintf("cannot save fileInfo: %v", err))
		return
	}

	news, err = dao.GetNewsDAO().AppendPreviewImage(ctx, newsId, fileInfo.FileId)
	if err != nil {
		ctx.String(http.StatusForbidden, fmt.Sprintf("cannot update News: %v", err))
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"status": "success"})
}

func (sv *NewsController) UploadMultiplePreviewImages(ctx *gin.Context) {
	//TODO

	// Multipart form
	form, _ := ctx.MultipartForm()
	files := form.File["upload[]"]
	for _, file := range files {
		log.Println(file.Filename)
	}

	ctx.JSON(http.StatusOK, gin.H{"status": "success"})
}

func (sv *NewsController) UploadSingleMedia(ctx *gin.Context) {
	//TODO

	ctx.JSON(http.StatusOK, gin.H{"status": "success"})
}

func (sv *NewsController) UploadMultipleMedias(ctx *gin.Context) {
	//TODO

	ctx.JSON(http.StatusOK, gin.H{"status": "success"})
}
