package appgrpc

import (
	"context"
	"fmt"
	"gbb.go/gvp/dao"
	"gbb.go/gvp/model"
	"gbb.go/gvp/proto/grpcXVPPb"
	"gbb.go/gvp/utils"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"log"
)

type NewsFileTest struct {
	NewsId   string
	FileName string
}

func (sv *XVPGRPCService) Test(ctx context.Context, req *grpcXVPPb.TestRequest) (*grpcXVPPb.TestResponse, error) {
	//region TEST S3
	//url, err := s3Handler.GetS3FileStore().GenObjectPresignedUrlFromUrl("1e724844-3968-4ed4-89df-ce7a08d949fd", "livestream.m3u8", 6*time.Hour)
	//if err != nil {
	//	log.Println(err)
	//	return &grpcXVPPb.TestResponse{}, nil
	//}
	//
	//log.Println(url)
	//endregion

	//region TEST TIME
	//log.Println("timeNow", time.Now().UnixMilli())
	//log.Println("timeNowUTC", utils.UTCNowMilli())
	//log.Println("beginningOfWeek", utils.UTCNowBeginningOfWeek())
	//log.Println("beginningOfMonth", utils.UTCNowBeginningOfMonth())
	//endregion

	//region TEST MAIL
	//mailSv := appmail.GetMailService()
	//mailSv.SendMail([]string{"saivnct@gmail.com"}, "hello", "text/plain", "hello world")
	//endregion

	//region TEST NEWS
	go func() {

		allNews := []*model.News{}

		filter := primitive.M{}

		findOptions := options.Find()
		sort := primitive.M{"createdAt": -1}
		findOptions.SetSort(sort)

		err := dao.GetNewsDAO().FindAll(context.Background(), filter, findOptions, &allNews)
		if err != nil {
			log.Println(err)
			return
		}

		for _, news := range allNews {
			log.Println(news.NewsId, news.MediaEncKey, news.Title)
		}

		//newsFiles := []NewsFileTest{
		//	NewsFileTest{
		//		NewsId:   "c43785a1-0021-4907-a310-4b5943b7bfc5",
		//		FileName: "walk_the_line.mov",
		//	},
		//	NewsFileTest{
		//		NewsId:   "4b1df48b-af19-4cbe-817c-65cd09aef613",
		//		FileName: "tomorrowland.mov",
		//	},
		//	NewsFileTest{
		//		NewsId:   "4aae6bc4-dc2c-4bd8-8831-3e4737d3258e",
		//		FileName: "thelittledeath.mov",
		//	},
		//	NewsFileTest{
		//		NewsId:   "8f47ddc6-8626-411d-b629-b7de52fc5dc1",
		//		FileName: "thelastwitchhunter.mov",
		//	},
		//	NewsFileTest{
		//		NewsId:   "6c22c038-1781-47d5-862c-cb5fcfb385a6",
		//		FileName: "swepvii.mov",
		//	},
		//	NewsFileTest{
		//		NewsId:   "908d616b-6d19-4b87-a0a4-34a7bd9e8a99",
		//		FileName: "sanandres.mov",
		//	},
		//	NewsFileTest{
		//		NewsId:   "959ac445-e06f-41a3-9989-95c7a9271cc0",
		//		FileName: "rickiandtheflash.mov",
		//	},
		//	NewsFileTest{
		//		NewsId:   "3e470f56-4fb6-4aa8-a0d5-8867fe59431c",
		//		FileName: "poltergeist.mov",
		//	},
		//	NewsFileTest{
		//		NewsId:   "57067c9d-5505-42bc-afd8-0ae8e2e669d5",
		//		FileName: "missionimpossibleroguenation.mov",
		//	},
		//	NewsFileTest{
		//		NewsId:   "b87c04e8-5c33-4d31-a7e6-e11427f21a6c",
		//		FileName: "maggie.mov",
		//	},
		//	NewsFileTest{
		//		NewsId:   "d353e537-2db4-4a68-a87e-093407689e46",
		//		FileName: "madmax.mov",
		//	},
		//	NewsFileTest{
		//		NewsId:   "3b2cb90a-d5e2-44bb-8c27-087635401dcd",
		//		FileName: "jurassicworld.mov",
		//	},
		//	NewsFileTest{
		//		NewsId:   "dce4464a-f779-4468-a2be-f5a6a575d2f0",
		//		FileName: "heavenknowswhat.mov",
		//	},
		//	NewsFileTest{
		//		NewsId:   "04590414-d689-4f9f-9d11-0fb07711c749",
		//		FileName: "freedom.mov",
		//	},
		//	NewsFileTest{
		//		NewsId:   "fb263cbd-cd3e-49af-b423-36f585a673d3",
		//		FileName: "fantasticfour.mov",
		//	},
		//	NewsFileTest{
		//		NewsId:   "b25c7acf-fbd5-4839-a5bd-968366ddb6d4",
		//		FileName: "dishonesty.mov",
		//	},
		//	NewsFileTest{
		//		NewsId:   "83f03963-0cb8-4708-93a3-eee0da1b3846",
		//		FileName: "darkwasthenight.mov",
		//	},
		//	NewsFileTest{
		//		NewsId:   "f2053860-5390-4c71-a23d-d25968a9efea",
		//		FileName: "batmanvssuperman.mov",
		//	},
		//	NewsFileTest{
		//		NewsId:   "ff8c01a4-d213-4314-ae57-f568495fb03f",
		//		FileName: "avengersageofultron.mov",
		//	},
		//	NewsFileTest{
		//		NewsId:   "d2044b9f-1c0d-49e5-9ddc-7c22ac63b97b",
		//		FileName: "Avatar.mp4",
		//	},
		//}
		//count := 0
		//for _, news := range allNews {
		//	newsId := news.NewsId
		//	mediaEncKey, err := utils.GenerateRandomHex(16)
		//	if err != nil {
		//		log.Println(err)
		//		continue
		//	}
		//
		//	mediaEncIV, err := utils.GenerateRandomHex(16)
		//	if err != nil {
		//		log.Println(err)
		//		continue
		//	}
		//
		//	updateFields := primitive.M{}
		//	updateFields["mediaEncKey"] = mediaEncKey
		//	updateFields["mediaEncIV"] = mediaEncIV
		//	update := primitive.M{"$set": updateFields}
		//
		//	newsUpdate, err := dao.GetNewsDAO().UpdateByNewsId(context.Background(), newsId, update, []interface{}{}, false)
		//	if err != nil {
		//		log.Println(err)
		//		continue
		//	}
		//
		//	idx := slices.IndexFunc(newsFiles, func(newsFile NewsFileTest) bool { return newsFile.NewsId == newsId })
		//	if idx >= 0 {
		//		log.Println("Executing Convert for", newsUpdate.Title, newsFiles[idx].FileName, "...")
		//		shPath := "/Users/giangbb/giangbb/workspace/golang/xvp/xvp-hls/gen-hls-v2.sh"
		//		args := []string{
		//			"-f", fmt.Sprintf("/Users/giangbb/Downloads/hls/%v", newsFiles[idx].FileName),
		//			"-n", newsId,
		//			"-i", newsUpdate.MediaEncIV,
		//			"-k", newsUpdate.MediaEncKey,
		//			"-r", "720",
		//		}
		//		result, err := utils.BashExecute(shPath, args)
		//		if err != nil {
		//			log.Println(err)
		//			return
		//		}
		//		log.Println(result)
		//	}
		//
		//	count++
		//	log.Println(newsId, news.Title)
		//}
		//
		//log.Println(count)
	}()
	//endregion

	//region RANDOM TEST
	//randHex, _ := utils.GenerateRandomHex(16)
	//randHex2, _ := utils.GenerateRandomHex(16)
	//authenCode := utils.GenerateAuthenCode()
	//log.Println("randHex", randHex)
	//log.Println("randHex", randHex2)
	//log.Println("authenCode", authenCode)
	//endregion

	//region BASH TEST
	//go func() {
	//
	//	log.Println("Execute ...")
	//	shPath := "/Users/giangbb/giangbb/workspace/golang/xvp/xvp-hls/gen-hls-v2.sh"
	//	args := []string{
	//		"-f", "/Users/giangbb/Downloads/hls/avatar.mp4",
	//		"-n", "17fce397-9d28-41cf-8c38-322067fc5845",
	//		"-i", "9566c74d10037c4d7bbb0407d1e2c649",
	//		"-k", "52fdfc072182654f163f5f0f9a621d72",
	//		"-r", "480",
	//	}
	//	result, err := utils.BashExecute(shPath, args)
	//	if err != nil {
	//		log.Println(err)
	//		return
	//	}
	//	log.Println(result)
	//}()
	//endregion

	return &grpcXVPPb.TestResponse{}, nil
}

func (sv *XVPGRPCService) TestCreateUser(ctx context.Context, req *grpcXVPPb.TestCreateUserRequest) (*grpcXVPPb.TestCreateUserResponse, error) {
	if len(req.GetUsername()) == 0 || len(req.GetEmail()) == 0 || len(req.GetPassword()) == 0 {
		return nil, status.Errorf(codes.InvalidArgument, "Invalid Arguments")
	}

	userWithSameEmail, err := dao.GetUserDAO().FindByEmail(ctx, req.GetEmail())
	if err == nil && userWithSameEmail != nil {
		return nil, status.Errorf(codes.AlreadyExists, "Email existed")
	}

	userWithSameUsername, err := dao.GetUserDAO().FindByUserName(ctx, req.GetUsername())
	if err == nil && userWithSameUsername != nil {
		return nil, status.Errorf(codes.AlreadyExists, "Username existed")
	}

	password, err := utils.HashPassword(req.GetPassword())
	if err != nil {
		log.Println(err)
		return nil, status.Errorf(codes.Internal, fmt.Sprintf("Internal err: %v", err))
	}

	user := &model.User{
		Username:  req.GetUsername(),
		Email:     req.GetEmail(),
		Role:      req.GetRole(),
		Password:  password,
		CreatedAt: utils.UTCNowMilli(),
	}

	user, err = dao.GetUserDAO().Save(ctx, user)
	if err != nil {
		log.Println(err)
		return nil, status.Errorf(codes.Internal, fmt.Sprintf("Internal err: %v", err))
	}

	return &grpcXVPPb.TestCreateUserResponse{
		Username: user.Username,
		Email:    user.Email,
		Role:     user.Role,
	}, nil
}
