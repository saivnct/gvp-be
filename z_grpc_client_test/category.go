package z_grpc_client_test

import (
	"context"
	"gbb.go/gvp/proto/grpcXVPPb"
	"github.com/davecgh/go-spew/spew"
	"google.golang.org/grpc/status"
	"log"
)

func GetAllCategories() {
	log.Println("GetAllCategories")

	cc, serviceClient, _ := GetServiceClient("")

	defer cc.Close() //this will call at very end of code

	request := grpcXVPPb.GetAllCategoryRequest{}

	response, err := serviceClient.GetAllCategory(context.Background(), &request)
	if err != nil {
		log.Fatalf("could not send GetAllCategory Request: %v - %v", status.Code(err), err)
	}

	spew.Dump(response)
}

func GetCategory() {
	log.Println("GetCategory")

	cc, serviceClient, _ := GetServiceClient("")

	defer cc.Close() //this will call at very end of code

	request := grpcXVPPb.GetCategoryRequest{
		CatId: "b70ca34e-868d-4b9f-9272-6064c5235de7",
	}

	response, err := serviceClient.GetCategory(context.Background(), &request)
	if err != nil {
		log.Fatalf("could not send GetCategory Request: %v - %v", status.Code(err), err)
	}

	spew.Dump(response)
}

func CreateCategory() {
	jwt := Login()

	cc, serviceClient, _ := GetServiceClient(jwt)

	defer cc.Close() //this will call at very end of code

	request := grpcXVPPb.CreateCategoryRequest{
		CatName:        "Category 6",
		CatDescription: "Category 6",
	}

	response, err := serviceClient.CreateCategory(context.Background(), &request)
	if err != nil {
		log.Fatalf("could not send CreateCategory Request: %v - %v", status.Code(err), err)
	}

	spew.Dump(response)
}

func UpdateCategory() {
	jwt := Login()

	cc, serviceClient, _ := GetServiceClient(jwt)

	defer cc.Close() //this will call at very end of code

	request := grpcXVPPb.UpdateCategoryRequest{
		CatId:          "7e081fbb-ee56-4f42-b598-c6574fa7a779",
		CatName:        "Category C",
		CatDescription: "Category C",
	}

	response, err := serviceClient.UpdateCategory(context.Background(), &request)
	if err != nil {
		log.Fatalf("could not send UpdateCategory Request: %v - %v", status.Code(err), err)
	}

	spew.Dump(response)
}

func DeleteCategory() {
	jwt := Login()

	cc, serviceClient, _ := GetServiceClient(jwt)

	defer cc.Close() //this will call at very end of code

	request := grpcXVPPb.DeleteCategoryRequest{
		CatId: "b7812374-2b54-4358-acc6-cdcaefc8ce9a",
	}

	response, err := serviceClient.DeleteCategory(context.Background(), &request)
	if err != nil {
		log.Fatalf("could not send DeleteCategory Request: %v - %v", status.Code(err), err)
	}

	spew.Dump(response)
}
