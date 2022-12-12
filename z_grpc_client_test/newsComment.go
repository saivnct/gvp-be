package z_grpc_client_test

import (
	"context"
	"gbb.go/gvp/proto/grpcXVPPb"
	"github.com/davecgh/go-spew/spew"
	"google.golang.org/grpc/status"
	"log"
)

func CreateNewsComment() {
	jwt := Login()

	cc, serviceClient, _ := GetServiceClient(jwt)

	defer cc.Close() //this will call at very end of code

	request := grpcXVPPb.CreateNewsCommentRequest{
		NewsId:          "bb88c4c2-da21-427c-8137-57e967d06591",
		ParentCommentId: "b634c423-813f-4748-9d49-69846cdcfda8",
		Content:         "comment 0 - 2 - 3",
	}

	response, err := serviceClient.CreateNewsComment(context.Background(), &request)
	if err != nil {
		log.Fatalf("could not send CreateNews Request: %v - %v", status.Code(err), err)
	}

	spew.Dump(response)
}

func UpdateNewsComment() {
	jwt := Login()

	cc, serviceClient, _ := GetServiceClient(jwt)

	defer cc.Close() //this will call at very end of code

	request := grpcXVPPb.UpdateNewsCommentRequest{
		CommentId: "3687674e-f36c-495c-80a8-aecfb73d0490",
		Content:   "comment 0 updated",
	}

	response, err := serviceClient.UpdateNewsComment(context.Background(), &request)
	if err != nil {
		log.Fatalf("could not send UpdateNewsComment Request: %v - %v", status.Code(err), err)
	}

	spew.Dump(response)
}

func DeleteNewsComment() {
	jwt := Login()

	cc, serviceClient, _ := GetServiceClient(jwt)

	defer cc.Close() //this will call at very end of code

	request := grpcXVPPb.DeleteNewsCommentRequest{
		CommentId: "3687674e-f36c-495c-80a8-aecfb73d0490",
	}

	response, err := serviceClient.DeleteNewsComment(context.Background(), &request)
	if err != nil {
		log.Fatalf("could not send DeleteNewsComment Request: %v - %v", status.Code(err), err)
	}

	spew.Dump(response)
}

func GetNewsComments() {
	cc, serviceClient, _ := GetServiceClient("")

	defer cc.Close() //this will call at very end of code

	request := grpcXVPPb.GetNewsCommentsRequest{
		NewsId: "d2044b9f-1c0d-49e5-9ddc-7c22ac63b97b",
		//ParentCommentId: "b634c423-813f-4748-9d49-69846cdcfda8",
		Page:     0,
		PageSize: 10,
	}

	response, err := serviceClient.GetNewsComments(context.Background(), &request)
	if err != nil {
		log.Fatalf("could not send DeleteNewsComment Request: %v - %v", status.Code(err), err)
	}

	spew.Dump(response)
}
