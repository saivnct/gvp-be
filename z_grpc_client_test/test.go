package z_grpc_client_test

import (
	"context"
	"gbb.go/gvp/proto/grpcXVPPb"
	"github.com/davecgh/go-spew/spew"
	"google.golang.org/grpc/status"
	"log"
)

func Test() {
	jwt := Login()

	cc, serviceClient, _ := GetServiceClient(jwt)

	defer cc.Close() //this will call at very end of code

	request := grpcXVPPb.TestRequest{}

	response, err := serviceClient.Test(context.Background(), &request)
	if err != nil {
		log.Fatalf("could not send Test Request: %v - %v", status.Code(err), err)
	}

	spew.Dump(response)
}
