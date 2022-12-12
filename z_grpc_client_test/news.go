package z_grpc_client_test

import (
	"bufio"
	"context"
	"crypto/md5"
	"fmt"
	"gbb.go/gvp/proto/grpcXVPPb"
	"github.com/davecgh/go-spew/spew"
	"google.golang.org/grpc/status"
	"io"
	"log"
	"os"
)

type NewsMediaUploadTest struct {
	NewsId string
	EncKey string
	Path   string
}

func CreateNews() {
	jwt := Login()

	cc, serviceClient, _ := GetServiceClient(jwt)

	defer cc.Close() //this will call at very end of code

	request := grpcXVPPb.CreateNewsRequest{
		Title:        "Testtttttt 222222",
		Description:  "Testtttttt 222222 Desccccccc",
		Participants: []string{},
		//CatIds:       []string{"b70ca34e-868d-4b9f-9272-6064c5235de7"},
		//CatIds:        []string{"6b6e9980-53e8-4dca-b3d9-6c53d8776bcd"},
		CatIds:        []string{"b70ca34e-868d-4b9f-9272-6064c5235de7"},
		Tags:          []string{},
		EnableComment: true,
	}

	response, err := serviceClient.CreateNews(context.Background(), &request)
	if err != nil {
		log.Fatalf("could not send CreateNews Request: %v - %v", status.Code(err), err)
	}

	spew.Dump(response)
}

func UpdateNewsInfo() {
	jwt := Login()

	cc, serviceClient, _ := GetServiceClient(jwt)

	defer cc.Close() //this will call at very end of code

	request := grpcXVPPb.UpdateNewsInfoRequest{
		NewsId:      "b8e72884-d70e-4fb4-ae14-9e891f7df570",
		Title:       "clip 2",
		Description: "clip 2",
		//CatIds:        []string{"7863bd09-8141-443b-9ad1-17bbfdd2e03f", "7e081fbb-ee56-4f42-b598-c6574fa7a779"},
		Participants:  []string{"act 3"},
		Tags:          []string{"tag 1", "tag 2"},
		EnableComment: true,
	}

	response, err := serviceClient.UpdateNewsInfo(context.Background(), &request)
	if err != nil {
		log.Fatalf("could not send UpdateNewsInfo Request: %v - %v", status.Code(err), err)
	}

	spew.Dump(response)
}

func DeleteNews() {
	jwt := Login()

	cc, serviceClient, _ := GetServiceClient(jwt)

	defer cc.Close() //this will call at very end of code

	request := grpcXVPPb.DeleteNewsRequest{
		NewsId: "17fce397-9d28-41cf-8c38-322067fc5845",
	}

	response, err := serviceClient.DeleteNews(context.Background(), &request)
	if err != nil {
		log.Fatalf("could not send DeleteNews Request: %v - %v", status.Code(err), err)
	}

	spew.Dump(response)
}

func UploadNewsPreviewImage() {
	jwt := Login()

	cc, serviceClient, _ := GetServiceClient(jwt)

	defer cc.Close() //this will call at very end of code

	//ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	//defer cancel()

	stream, err := serviceClient.UploadNewsPreviewImage(context.Background())
	if err != nil {
		log.Fatal("cannot upload file: ", err)
	}

	newsId := "1389fea1-6b11-4f41-bbe4-9160b884bab2"

	//region upload  local file
	imagePath := "/Users/solgo/kingsman/xvp/ffmpeg/thump-Big-Buck-Bunny-1080p.mp4"
	//imagePath := "/Users/giangbb/Downloads/img2.jpg"
	file, err := os.Open(imagePath)
	if err != nil {
		log.Fatal("cannot open file: ", err)
	}
	defer file.Close()

	hash := md5.New()
	if _, err := io.Copy(hash, file); err != nil {
		log.Fatal(err)
	}
	checksum := fmt.Sprintf("%x", hash.Sum(nil))

	fileInfo, err := file.Stat()
	if err != nil {
		log.Fatal("cannot obtain file stat: ", err)
	}

	log.Printf("file size: %d, checksum %v", fileInfo.Size(), checksum)

	req := &grpcXVPPb.UploadNewsPreviewImageRequest{
		Data: &grpcXVPPb.UploadNewsPreviewImageRequest_UploadNewsPreviewImageInfo{
			UploadNewsPreviewImageInfo: &grpcXVPPb.UploadNewsPreviewImageInfo{
				NewsId:      newsId,
				FileName:    "img.png",
				Checksum:    checksum,
				MainPreview: false,
				MediaType:   grpcXVPPb.MEDIA_TYPE_VIDEO,
			},
		},
	}

	err = stream.Send(req)
	if err != nil {
		log.Fatal("cannot send file info to server: ", err, stream.RecvMsg(nil))
	}

	file.Seek(0, io.SeekStart)
	reader := bufio.NewReader(file)
	buffer := make([]byte, 1024)

	for {
		n, err := reader.Read(buffer)
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal("cannot read chunk to buffer: ", err)
		}

		req := &grpcXVPPb.UploadNewsPreviewImageRequest{
			Data: &grpcXVPPb.UploadNewsPreviewImageRequest_ChunkData{
				ChunkData: buffer[:n],
			},
		}

		err = stream.Send(req)
		if err != nil {
			//To get the real error that contains the gRPC status code, we must call stream.RecvMsg() with a nil parameter. The nil parameter basically means that we don't expect to receive any message, but we just want to get the error that function returns
			log.Fatal("cannot send chunk to server: ", err, stream.RecvMsg(nil))
		}
	}
	//endregion

	////region upload url file
	//req := &grpcXVPPb.UploadNewsPreviewImageRequest{
	//	Data: &grpcXVPPb.UploadNewsPreviewImageRequest_UploadNewsPreviewImageInfo{
	//		UploadNewsPreviewImageInfo: &grpcXVPPb.UploadNewsPreviewImageInfo{
	//			NewsId:      newsId,
	//			FileUrl:     "https://2.bp.blogspot.com/-8plGmHtcDU8/U4s7GVilkbI/AAAAAAAACMc/uyRPksGp2dg/s1600/troll.jpeg",
	//			MainPreview: true,
	//		},
	//	},
	//}
	//
	//err = stream.Send(req)
	//if err != nil {
	//	log.Fatal("cannot send file info to server: ", err, stream.RecvMsg(nil))
	//}
	////endregion

	res, err := stream.CloseAndRecv()
	if err != nil {
		log.Fatal("cannot receive response: ", err)
	}

	log.Println("res:", res.GetNewsInfo())
}

func DeleteNewsPreviewImage() {
	jwt := Login()

	cc, serviceClient, _ := GetServiceClient(jwt)

	defer cc.Close() //this will call at very end of code

	request := grpcXVPPb.DeleteNewsPreviewImageRequest{
		NewsId: "1389fea1-6b11-4f41-bbe4-9160b884bab2",
		FileId: "0378e2f6-3c1f-4542-bcf9-e7311396aa03",
	}

	response, err := serviceClient.DeleteNewsPreviewImage(context.Background(), &request)
	if err != nil {
		log.Fatalf("could not send DeleteNewsPreviewImage Request: %v - %v", status.Code(err), err)
	}

	spew.Dump(response)
}

// upload local HLS files
func UploadNewsOndemandMedia(newsMediaUploadTest NewsMediaUploadTest) {
	jwt := Login()

	cc, serviceClient, _ := GetServiceClient(jwt)

	defer cc.Close() //this will call at very end of code

	newsId := newsMediaUploadTest.NewsId
	mediaEncKey := newsMediaUploadTest.EncKey
	mediaPath := fmt.Sprintf("/Users/giangbb/Downloads/hls/%v", newsMediaUploadTest.Path)

	mediaFile, err := os.Open(mediaPath)
	files, err := mediaFile.Readdir(0)
	if err != nil {
		log.Fatal(err)
	}

	foundOnDemandMediaMainFile := false

	fileNames := []string{"livestream.m3u8"}
	for _, f := range files {
		if !f.IsDir() {

			if f.Name() == "livestream.m3u8" {
				foundOnDemandMediaMainFile = true
			} else {
				fileNames = append(fileNames, f.Name())
			}
		}
	}

	if !foundOnDemandMediaMainFile {
		log.Fatal("not found OnDemand Media MainFile")
	}

	onDemandMediaMainFileId := ""
	for _, fileName := range fileNames {
		filePath := fmt.Sprintf("%v/%v", mediaPath, fileName)

		stream, err := serviceClient.UploadNewsMedia(context.Background())
		if err != nil {
			log.Fatal("cannot upload file: ", err)
		}

		file, err := os.Open(filePath)
		if err != nil {
			log.Fatal("cannot open file: ", err)
		}

		hash := md5.New()
		if _, err := io.Copy(hash, file); err != nil {
			log.Fatal(err)
		}
		checksum := fmt.Sprintf("%x", hash.Sum(nil))

		fileInfo, err := file.Stat()
		if err != nil {
			log.Fatal("cannot obtain file stat: ", err)
		}

		log.Printf("fileName: %v, file size: %d, checksum %v", fileName, fileInfo.Size(), checksum)

		req := &grpcXVPPb.UploadNewsMediaRequest{
			Data: &grpcXVPPb.UploadNewsMediaRequest_UploadNewsMediaInfo{
				UploadNewsMediaInfo: &grpcXVPPb.UploadNewsMediaInfo{
					NewsId:                  newsId,
					FileName:                fileName,
					MediaType:               grpcXVPPb.MEDIA_TYPE_VIDEO,
					Checksum:                checksum,
					MediaStreamType:         grpcXVPPb.MEDIA_STREAM_TYPE_ON_DEMAND,
					MediaEncKey:             mediaEncKey,
					OnDemandMediaMainFile:   fileName == "livestream.m3u8",
					OnDemandMediaMainFileId: onDemandMediaMainFileId,
				},
			},
		}

		err = stream.Send(req)
		if err != nil {
			log.Fatal("cannot send file info to server: ", err, stream.RecvMsg(nil))
		}

		_, err = file.Seek(0, io.SeekStart)
		if err != nil {
			log.Fatal("cannot send file info to server: ", err, stream.RecvMsg(nil))
		}
		reader := bufio.NewReader(file)
		buffer := make([]byte, 1024)

		for {
			n, err := reader.Read(buffer)
			if err == io.EOF {
				break
			}
			if err != nil {
				log.Fatal("cannot read chunk to buffer: ", err)
			}

			req := &grpcXVPPb.UploadNewsMediaRequest{
				Data: &grpcXVPPb.UploadNewsMediaRequest_ChunkData{
					ChunkData: buffer[:n],
				},
			}

			err = stream.Send(req)
			if err != nil {
				//To get the real error that contains the gRPC status code, we must call stream.RecvMsg() with a nil parameter. The nil parameter basically means that we don't expect to receive any message, but we just want to get the error that function returns
				log.Fatal("cannot send chunk to server: ", err, stream.RecvMsg(nil))
			}
		}

		res, err := stream.CloseAndRecv()
		if err != nil {
			log.Fatal("cannot receive response: ", err)
		}

		log.Println("res:", res.GetNewsInfo())
		log.Println("fileId:", res.GetFileId())

		if fileName == "livestream.m3u8" {
			onDemandMediaMainFileId = res.GetFileId()
		}

		err = file.Close()
		if err != nil {
			log.Fatal("error close file: ", err)
		}
	}
}

func UploadNewsMedia() {
	jwt := Login()

	cc, serviceClient, _ := GetServiceClient(jwt)

	defer cc.Close() //this will call at very end of code

	//ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	//defer cancel()

	stream, err := serviceClient.UploadNewsMedia(context.Background())
	if err != nil {
		log.Fatal("cannot upload file: ", err)
	}

	newsId := "fa533fe9-e78f-4870-89a6-69c303353c75"

	//region upload local MP4 file
	filePath := "/home/gbb/gbb/xvp/test/mad-max/big-buck-bunny-1080p.mp4"
	fileName := "big-buck-bunny-1080p.mp4"

	file, err := os.Open(filePath)
	if err != nil {
		log.Fatal("cannot open file: ", err)
	}
	defer file.Close()

	hash := md5.New()
	if _, err := io.Copy(hash, file); err != nil {
		log.Fatal(err)
	}
	checksum := fmt.Sprintf("%x", hash.Sum(nil))

	fileInfo, err := file.Stat()
	if err != nil {
		log.Fatal("cannot obtain file stat: ", err)
	}

	log.Printf("fileName: %v, file size: %d, checksum %v", fileName, fileInfo.Size(), checksum)

	resolution := grpcXVPPb.VIDEO_RESOLUTION_HD
	req := &grpcXVPPb.UploadNewsMediaRequest{
		Data: &grpcXVPPb.UploadNewsMediaRequest_UploadNewsMediaInfo{
			UploadNewsMediaInfo: &grpcXVPPb.UploadNewsMediaInfo{
				NewsId:          newsId,
				FileName:        fileName,
				MediaType:       grpcXVPPb.MEDIA_TYPE_VIDEO,
				Resolution:      &resolution,
				Checksum:        checksum,
				MediaStreamType: grpcXVPPb.MEDIA_STREAM_TYPE_BUFFERING,
			},
		},
	}

	err = stream.Send(req)
	if err != nil {
		log.Fatal("cannot send file info to server: ", err, stream.RecvMsg(nil))
	}

	file.Seek(0, io.SeekStart)
	reader := bufio.NewReader(file)
	buffer := make([]byte, 1024)

	for {
		n, err := reader.Read(buffer)
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal("cannot read chunk to buffer: ", err)
		}

		req := &grpcXVPPb.UploadNewsMediaRequest{
			Data: &grpcXVPPb.UploadNewsMediaRequest_ChunkData{
				ChunkData: buffer[:n],
			},
		}

		err = stream.Send(req)
		if err != nil {
			//To get the real error that contains the gRPC status code, we must call stream.RecvMsg() with a nil parameter. The nil parameter basically means that we don't expect to receive any message, but we just want to get the error that function returns
			log.Fatal("cannot send chunk to server: ", err, stream.RecvMsg(nil))
		}
	}
	//endregion

	////region upload url file
	//resolution := grpcXVPPb.VIDEO_RESOLUTION_SD
	//req := &grpcXVPPb.UploadNewsMediaRequest{
	//	Data: &grpcXVPPb.UploadNewsMediaRequest_UploadNewsMediaInfo{
	//		UploadNewsMediaInfo: &grpcXVPPb.UploadNewsMediaInfo{
	//			NewsId:     newsId,
	//			FileUrl:    "https://test-videos.co.uk/vids/bigbuckbunny/mp4/h264/360/Big_Buck_Bunny_360_10s_1MB.mp4",
	//			MediaType:  grpcXVPPb.MEDIA_TYPE_VIDEO,
	//			Resolution: &resolution,
	//			MediaStreamType: grpcXVPPb.MEDIA_STREAM_TYPE_BUFFERING,
	//		},
	//	},
	//}
	//
	//err = stream.Send(req)
	//if err != nil {
	//	log.Fatal("cannot send file info to server: ", err, stream.RecvMsg(nil))
	//}
	////endregion

	res, err := stream.CloseAndRecv()
	if err != nil {
		log.Fatal("cannot receive response: ", err)
	}

	log.Println("res:", res.GetNewsInfo())
}

func DeleteNewsMedia() {
	jwt := Login()

	cc, serviceClient, _ := GetServiceClient(jwt)

	defer cc.Close() //this will call at very end of code

	request := grpcXVPPb.DeleteNewsMediaRequest{
		NewsId: "c43785a1-0021-4907-a310-4b5943b7bfc5",
		FileId: "c43785a1-0021-4907-a310-4b5943b7bfc5-5a3da73525c6aad43a79d8847af5e87a",
	}

	response, err := serviceClient.DeleteNewsMedia(context.Background(), &request)
	if err != nil {
		log.Fatalf("could not send DeleteNewsMedia Request: %v - %v", status.Code(err), err)
	}

	spew.Dump(response)
}

func GetNews() {
	cc, serviceClient, _ := GetServiceClient("")

	//jwt := Login()
	//cc, serviceClient, _ := GetServiceClient(jwt)

	defer cc.Close() //this will call at very end of code

	request := grpcXVPPb.GetNewsRequest{
		NewsId: "17fce397-9d28-41cf-8c38-322067fc5845",
	}

	response, err := serviceClient.GetNews(context.Background(), &request)
	if err != nil {
		log.Fatalf("could not send GetNews Request: %v - %v", status.Code(err), err)
	}

	spew.Dump(response)
}

func GetListNews() {

	cc, serviceClient, _ := GetServiceClient("")

	defer cc.Close() //this will call at very end of code

	request := grpcXVPPb.GetListNewsRequest{
		//CatIds:       []string{"8047ba41-03fe-45ca-b8b6-aef4bd5c3015"},
		//Tags:         []string{"tag 1"},
		//SearchPhrase: "1",
		//Page:         0,
		//PageSize:     10,
	}

	response, err := serviceClient.GetListNews(context.Background(), &request)
	if err != nil {
		log.Fatalf("could not send GetListNews Request: %v - %v", status.Code(err), err)
	}

	spew.Dump(response)
}

func GetManagerListNews() {
	jwt := Login()

	cc, serviceClient, _ := GetServiceClient(jwt)

	defer cc.Close() //this will call at very end of code

	request := grpcXVPPb.GetManagerListNewsRequest{
		//CatIds:       []string{"8047ba41-03fe-45ca-b8b6-aef4bd5c3015"},
		//Tags:         []string{"tag 1"},
		//SearchPhrase: "1",
		//Page:         0,
		//PageSize:     10,
		//Author: "content",
	}

	response, err := serviceClient.GetManagerListNews(context.Background(), &request)
	if err != nil {
		log.Fatalf("could not send GetListNews Request: %v - %v", status.Code(err), err)
	}

	spew.Dump(response)
}

func GetListTopViewNews() {
	cc, serviceClient, _ := GetServiceClient("")

	defer cc.Close() //this will call at very end of code

	request := grpcXVPPb.GetListTopNewsRequest{
		Limit:   100,
		TopType: grpcXVPPb.TOP_TYPE_VIEWS,
		//TopType: grpcXVPPb.TOP_TYPE_LIKES,
		//TopType: grpcXVPPb.TOP_TYPE_RATING,

		//TopTimeType: grpcXVPPb.TOP_TIME_TYPE_WEEK,
		TopTimeType: grpcXVPPb.TOP_TIME_TYPE_MONTH,
		//TopTimeType: grpcXVPPb.TOP_TIME_TYPE_ALL,
	}

	response, err := serviceClient.GetListTopNews(context.Background(), &request)
	if err != nil {
		log.Fatalf("could not send GetListTopViewNews Request: %v - %v", status.Code(err), err)
	}

	for _, newsInfo := range response.GetNewsInfos() {
		//VIEWS
		log.Printf("id: %v - name: %v - total views: %v - week: %v - weekViews: %v - month: %v - monthViews: %v",
			newsInfo.GetNewsId(), newsInfo.GetTitle(), newsInfo.GetView(),
			newsInfo.GetCurrentViewsWeek(), newsInfo.GetWeekViews(),
			newsInfo.GetCurrentViewsMonth(), newsInfo.GetMonthViews())

		//LIKES
		//log.Printf("id: %v - name: %v - total likes: %v - week: %v - weekLikes: %v - month: %v - monthLikes: %v",
		//	newsInfo.GetNewsId(), newsInfo.GetTitle(), newsInfo.GetLikes(),
		//	newsInfo.GetCurrentLikesWeek(), newsInfo.GetWeekLikes(),
		//	newsInfo.GetCurrentLikesMonth(), newsInfo.GetMonthLikes())

		//RATING
		//log.Printf("id: %v - name: %v - rating: %v",
		//	newsInfo.GetNewsId(), newsInfo.GetTitle(), newsInfo.GetRating())
	}

}

func LikeNews() {
	jwt := Login()

	cc, serviceClient, _ := GetServiceClient(jwt)

	defer cc.Close() //this will call at very end of code

	request := grpcXVPPb.LikeNewsRequest{
		NewsId: "1e724844-3968-4ed4-89df-ce7a08d949fd",
	}

	response, err := serviceClient.LikeNews(context.Background(), &request)
	if err != nil {
		log.Fatalf("could not send GetNews Request: %v - %v", status.Code(err), err)
	}

	spew.Dump(response)
}

func RateNews() {
	jwt := Login()

	cc, serviceClient, _ := GetServiceClient(jwt)

	defer cc.Close() //this will call at very end of code

	request := grpcXVPPb.RateNewsRequest{
		NewsId: "8e7a3494-a772-4a58-a75a-b6063aa50294",
		Point:  5,
	}

	response, err := serviceClient.RateNews(context.Background(), &request)
	if err != nil {
		log.Fatalf("could not send GetNews Request: %v - %v", status.Code(err), err)
	}

	spew.Dump(response)
}
