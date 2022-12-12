package z_grpc_client_test

import (
	"context"
	"gbb.go/gvp/proto/grpcXVPPb"
	"github.com/davecgh/go-spew/spew"
	"google.golang.org/grpc/status"
	"log"
)

func CreatTestUser() {
	log.Println("CreatTestUser")

	cc, serviceClient, _ := GetServiceClient("")

	defer cc.Close() //this will call at very end of code

	//admin
	//request := grpcXVPPb.TestCreateUserRequest{
	//	Username: "admin",
	//	Email:    "admin@admin.com",
	//	Password: "123465",
	//	Role:     grpcXVPPb.USER_ROLE_ADMIN,
	//}

	//normal user
	//request := grpcXVPPb.TestCreateUserRequest{
	//	Username: "test",
	//	Email:    "test@test.com",
	//	Password: "123465",
	//	Role:     grpcXVPPb.USER_ROLE_NORMAL_USER,
	//}

	//content user
	request := grpcXVPPb.TestCreateUserRequest{
		Username: "content2",
		Email:    "content2@test.com",
		Password: "123465aA@",
		Role:     grpcXVPPb.USER_ROLE_CONTENT_USER,
	}

	response, err := serviceClient.TestCreateUser(context.Background(), &request)
	if err != nil {
		log.Fatalf("could not send CreatUser Request: %v - %v", status.Code(err), err)
	}

	spew.Dump(response)
}

func CreatUser() {
	log.Println("CreatUser")

	cc, serviceClient, _ := GetServiceClient("")

	defer cc.Close() //this will call at very end of code

	//content user
	request := grpcXVPPb.CreatAccountRequest{
		Username: "saivnct",
		Email:    "saivnct@gmail.com",
		Password: "123465",
	}

	response, err := serviceClient.CreatUser(context.Background(), &request)
	if err != nil {
		log.Fatalf("could not send UpdateProfile Request: %v - %v", status.Code(err), err)
	}

	spew.Dump(response)
}

func VerifyAuthencode() {
	log.Println("VerifyAuthencode")

	cc, serviceClient, _ := GetServiceClient("")

	defer cc.Close() //this will call at very end of code

	//content user
	request := grpcXVPPb.VerifyAuthencodeRequest{
		Username:   "test2",
		Email:      "test2@gmail.com",
		Authencode: "177918",
	}

	response, err := serviceClient.VerifyAuthencode(context.Background(), &request)
	if err != nil {
		log.Fatalf("could not send UpdateProfile Request: %v - %v", status.Code(err), err)
	}

	spew.Dump(response)
}

func Login() string {
	log.Println("Login")

	cc, serviceClient, _ := GetServiceClient("")

	defer cc.Close() //this will call at very end of code

	//admin
	//request := grpcXVPPb.LoginRequest{
	//	Username: "admin",
	//	Password: "2911@Saivnct",
	//}

	//normal user
	//request := grpcXVPPb.LoginRequest{
	//	Username: "test",
	//	Password: "123465",
	//}

	//content user
	request := grpcXVPPb.LoginRequest{
		Username: "content",
		Password: "123465aA@",
	}

	//content2 user
	//request := grpcXVPPb.LoginRequest{
	//	Username: "content2",
	//	Password: "123465aA@",
	//}

	response, err := serviceClient.Login(context.Background(), &request)
	if err != nil {
		log.Fatalf("could not send Login Request: %v - %v", status.Code(err), err)
	}

	spew.Dump(response)
	return response.GetJwt()
}

func UpdateProfile() {
	log.Println("UpdateProfile")

	jwt := Login()

	cc, serviceClient, _ := GetServiceClient(jwt)

	defer cc.Close() //this will call at very end of code

	//content user
	request := grpcXVPPb.UpdateProfileRequest{
		FirstName:   "tttt",
		LastName:    "eeesstt",
		PhoneNumber: "0123456789",
		Gender:      grpcXVPPb.USER_GENDER_MALE,
		Birthday:    1669084669000, //unix milli second
	}

	response, err := serviceClient.UpdateProfile(context.Background(), &request)
	if err != nil {
		log.Fatalf("could not send UpdateProfile Request: %v - %v", status.Code(err), err)
	}

	spew.Dump(response)
}

func UploadUserAvatar() {
	log.Println("UploadUserAvatar")

	jwt := Login()

	cc, serviceClient, _ := GetServiceClient(jwt)

	defer cc.Close() //this will call at very end of code

	stream, err := serviceClient.UploadUserAvatar(context.Background())
	if err != nil {
		log.Fatal("cannot upload file: ", err)
	}

	////region upload  local file
	//imagePath := "/Users/giangbb/Downloads/img.jpg"
	//
	//file, err := os.Open(imagePath)
	//if err != nil {
	//	log.Fatal("cannot open file: ", err)
	//}
	//defer file.Close()
	//
	//hash := md5.New()
	//if _, err := io.Copy(hash, file); err != nil {
	//	log.Fatal(err)
	//}
	//checksum := fmt.Sprintf("%x", hash.Sum(nil))
	//
	//fileInfo, err := file.Stat()
	//if err != nil {
	//	log.Fatal("cannot obtain file stat: ", err)
	//}
	//
	//log.Printf("file size: %d, checksum %v", fileInfo.Size(), checksum)
	//
	//req := &grpcXVPPb.UploadUserAvatarRequest{
	//	Data: &grpcXVPPb.UploadUserAvatarRequest_UploadUserAvatarInfo{
	//		UploadUserAvatarInfo: &grpcXVPPb.UploadUserAvatarInfo{
	//			//FileUrl:  "",
	//			Checksum: checksum,
	//		},
	//	},
	//}
	//
	//err = stream.Send(req)
	//if err != nil {
	//	log.Fatal("cannot send file info to server: ", err, stream.RecvMsg(nil))
	//}
	//
	//file.Seek(0, io.SeekStart)
	//reader := bufio.NewReader(file)
	//buffer := make([]byte, 1024)
	//
	//for {
	//	n, err := reader.Read(buffer)
	//	if err == io.EOF {
	//		break
	//	}
	//	if err != nil {
	//		log.Fatal("cannot read chunk to buffer: ", err)
	//	}
	//
	//	req := &grpcXVPPb.UploadUserAvatarRequest{
	//		Data: &grpcXVPPb.UploadUserAvatarRequest_ChunkData{
	//			ChunkData: buffer[:n],
	//		},
	//	}
	//
	//	err = stream.Send(req)
	//	if err != nil {
	//		//To get the real error that contains the gRPC status code, we must call stream.RecvMsg() with a nil parameter. The nil parameter basically means that we don't expect to receive any message, but we just want to get the error that function returns
	//		log.Fatal("cannot send chunk to server: ", err, stream.RecvMsg(nil))
	//	}
	//}
	////endregion

	//region upload url file
	req := &grpcXVPPb.UploadUserAvatarRequest{
		Data: &grpcXVPPb.UploadUserAvatarRequest_UploadUserAvatarInfo{
			UploadUserAvatarInfo: &grpcXVPPb.UploadUserAvatarInfo{
				FileUrl: "https://cdn.vntrip.vn/cam-nang/wp-content/uploads/2020/10/meme-hot-1.jpg",
			},
		},
	}

	err = stream.Send(req)
	if err != nil {
		log.Fatal("cannot send file info to server: ", err, stream.RecvMsg(nil))
	}
	//endregion

	res, err := stream.CloseAndRecv()
	if err != nil {
		log.Fatal("cannot receive response: ", err)
	}

	spew.Dump(res)

}

func GetUserInfoV1() {
	log.Println("GetUserInfo")

	cc, serviceClient, _ := GetServiceClient("")

	defer cc.Close() //this will call at very end of code

	//content user
	request := grpcXVPPb.GetUserInfoRequest{
		Username: "content",
	}

	response, err := serviceClient.GetUserInfo(context.Background(), &request)
	if err != nil {
		log.Fatalf("could not send GetUserInfo Request: %v - %v", status.Code(err), err)
	}

	spew.Dump(response)
}

func GetUserInfoV2() {
	log.Println("GetMyInfo")

	jwt := Login()

	cc, serviceClient, _ := GetServiceClient(jwt)

	defer cc.Close() //this will call at very end of code

	//content user
	request := grpcXVPPb.GetUserInfoRequest{
		Username: "test",
	}

	response, err := serviceClient.GetUserInfo(context.Background(), &request)
	if err != nil {
		log.Fatalf("could not send GetUserInfo Request: %v - %v", status.Code(err), err)
	}

	spew.Dump(response)
}
