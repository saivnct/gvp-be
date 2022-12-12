package main

import (
	"fmt"
	"log"
)

func main() {
	fmt.Println("Hello, I'm a client")

	//testing
	//z_grpc_client_test.CreatTestUser() //for testing only
	//z_grpc_client_test.Test()

	//user
	//z_grpc_client_test.CreatUser()
	//z_grpc_client_test.VerifyAuthencode()
	//z_grpc_client_test.Login()
	//z_grpc_client_test.UpdateProfile()
	//z_grpc_client_test.UploadUserAvatar()
	//z_grpc_client_test.GetUserInfoV1() //Guest Call
	//z_grpc_client_test.GetUserInfoV2() //User Call

	//category
	//z_grpc_client_test.GetAllCategories()
	//z_grpc_client_test.GetCategory()
	//z_grpc_client_test.CreateCategory()
	//z_grpc_client_test.UpdateCategory()
	//z_grpc_client_test.DeleteCategory()

	//news
	//z_grpc_client_test.CreateNews()
	//z_grpc_client_test.UpdateNewsInfo()
	//z_grpc_client_test.DeleteNews()

	//z_grpc_client_test.UploadNewsPreviewImage()
	//z_grpc_client_test.DeleteNewsPreviewImage()

	//z_grpc_client_test.UploadNewsMedia()
	//z_grpc_client_test.UploadNewsOndemandMedia()
	//z_grpc_client_test.DeleteNewsMedia()

	//z_grpc_client_test.GetNews()
	//z_grpc_client_test.GetListNews()
	//z_grpc_client_test.GetManagerListNews()
	//z_grpc_client_test.GetListTopViewNews()

	//z_grpc_client_test.LikeNews()
	//z_grpc_client_test.RateNews()

	//files
	//z_grpc_client_test.DownloadFile()
	//z_grpc_client_test.GetFilePresignedUrl()
	//z_grpc_client_test.GetFileInfo()

	//news comment
	//z_grpc_client_test.CreateNewsComment()
	//z_grpc_client_test.UpdateNewsComment()
	//z_grpc_client_test.DeleteNewsComment()
	//z_grpc_client_test.GetNewsComments()

	//news tags
	//z_grpc_client_test.GetNewsTags()
	//z_grpc_client_test.GetNewsParticipants()

	//newsMediaUploadTests := []z_grpc_client_test.NewsMediaUploadTest{
	//	z_grpc_client_test.NewsMediaUploadTest{
	//		NewsId: "c43785a1-0021-4907-a310-4b5943b7bfc5",
	//		EncKey: "91cfe3605712a677c9f4b3f3ac7dc4e2",
	//		Path:   "walk_the_line",
	//	},
	//	z_grpc_client_test.NewsMediaUploadTest{
	//		NewsId: "4b1df48b-af19-4cbe-817c-65cd09aef613",
	//		EncKey: "c0fb06018b89584f37c8943cabc21e3c",
	//		Path:   "tomorrowland",
	//	},
	//	z_grpc_client_test.NewsMediaUploadTest{
	//		NewsId: "4aae6bc4-dc2c-4bd8-8831-3e4737d3258e",
	//		EncKey: "f336597743b9cf1c85d3458c4411abbb",
	//		Path:   "thelittledeath",
	//	},
	//	z_grpc_client_test.NewsMediaUploadTest{
	//		NewsId: "8f47ddc6-8626-411d-b629-b7de52fc5dc1",
	//		EncKey: "27ee173a7e126ffc174e647e27cda7b7",
	//		Path:   "thelastwitchhunter",
	//	},
	//	z_grpc_client_test.NewsMediaUploadTest{
	//		NewsId: "6c22c038-1781-47d5-862c-cb5fcfb385a6",
	//		EncKey: "8adec0443c9f9cbeead41d2368a3a875",
	//		Path:   "swepvii",
	//	},
	//	z_grpc_client_test.NewsMediaUploadTest{
	//		NewsId: "908d616b-6d19-4b87-a0a4-34a7bd9e8a99",
	//		EncKey: "84cd195a7b357b87b4dc406496183b11",
	//		Path:   "sanandres",
	//	},
	//	z_grpc_client_test.NewsMediaUploadTest{
	//		NewsId: "959ac445-e06f-41a3-9989-95c7a9271cc0",
	//		EncKey: "f2c52144823999fe606d3b0e09f77bac",
	//		Path:   "rickiandtheflash",
	//	},
	//	z_grpc_client_test.NewsMediaUploadTest{
	//		NewsId: "3e470f56-4fb6-4aa8-a0d5-8867fe59431c",
	//		EncKey: "993c5714e4614cb0627858e401c4e07e",
	//		Path:   "poltergeist",
	//	},
	//	z_grpc_client_test.NewsMediaUploadTest{
	//		NewsId: "57067c9d-5505-42bc-afd8-0ae8e2e669d5",
	//		EncKey: "5733a111eedc59885938f7c46353d565",
	//		Path:   "missionimpossibleroguenation",
	//	},
	//	z_grpc_client_test.NewsMediaUploadTest{
	//		NewsId: "b87c04e8-5c33-4d31-a7e6-e11427f21a6c",
	//		EncKey: "60959794a530ff5c0682b88d5b96a375",
	//		Path:   "maggie",
	//	},
	//	z_grpc_client_test.NewsMediaUploadTest{
	//		NewsId: "d353e537-2db4-4a68-a87e-093407689e46",
	//		EncKey: "592608803dca0847af73f31cbfd43ee1",
	//		Path:   "madmax",
	//	},
	//	z_grpc_client_test.NewsMediaUploadTest{
	//		NewsId: "3b2cb90a-d5e2-44bb-8c27-087635401dcd",
	//		EncKey: "479f87ff59761436b09f1a86ae627928",
	//		Path:   "jurassicworld",
	//	},
	//	z_grpc_client_test.NewsMediaUploadTest{
	//		NewsId: "dce4464a-f779-4468-a2be-f5a6a575d2f0",
	//		EncKey: "e722d6288e756c588ca5944909a21a4a",
	//		Path:   "heavenknowswhat",
	//	},
	//	z_grpc_client_test.NewsMediaUploadTest{
	//		NewsId: "04590414-d689-4f9f-9d11-0fb07711c749",
	//		EncKey: "2e064dd56669896a0735ece920cb019d",
	//		Path:   "freedom",
	//	},
	//	z_grpc_client_test.NewsMediaUploadTest{
	//		NewsId: "fb263cbd-cd3e-49af-b423-36f585a673d3",
	//		EncKey: "7d435bfcd8965ade892c4fde38cfe094",
	//		Path:   "fantasticfour",
	//	},
	//	z_grpc_client_test.NewsMediaUploadTest{
	//		NewsId: "b25c7acf-fbd5-4839-a5bd-968366ddb6d4",
	//		EncKey: "8dbdf2d35dd357eeabff146a66c3fd85",
	//		Path:   "dishonesty",
	//	},
	//	z_grpc_client_test.NewsMediaUploadTest{
	//		NewsId: "83f03963-0cb8-4708-93a3-eee0da1b3846",
	//		EncKey: "a77d0c20535d9b700cc6a3980f8cfd2b",
	//		Path:   "darkwasthenight",
	//	},
	//	z_grpc_client_test.NewsMediaUploadTest{
	//		NewsId: "f2053860-5390-4c71-a23d-d25968a9efea",
	//		EncKey: "84c848c1b1d3675b2519d55962ebffdb",
	//		Path:   "batmanvssuperman",
	//	},
	//	z_grpc_client_test.NewsMediaUploadTest{
	//		NewsId: "ff8c01a4-d213-4314-ae57-f568495fb03f",
	//		EncKey: "f24180e9c2270345fd4ea5ef52b22be6",
	//		Path:   "avengersageofultron",
	//	},
	//	z_grpc_client_test.NewsMediaUploadTest{
	//		NewsId: "d2044b9f-1c0d-49e5-9ddc-7c22ac63b97b",
	//		EncKey: "783968ba0373e3a5058b882b360af3ac",
	//		Path:   "Avatar",
	//	},
	//}
	//
	//for _, newsMediaUploadTest := range newsMediaUploadTests {
	//	z_grpc_client_test.UploadNewsOndemandMedia(newsMediaUploadTest)
	//}

	log.Println("Done!")
}
