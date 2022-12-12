package z_grpc_client_test

import (
	"context"
	"gbb.go/gvp/proto/grpcXVPPb"
	"github.com/davecgh/go-spew/spew"
	"google.golang.org/grpc/status"
	"log"
)

func GetNewsParticipants() {
	cc, serviceClient, _ := GetServiceClient("")

	defer cc.Close() //this will call at very end of code

	request := grpcXVPPb.GetNewsParticipantsRequest{
		Page:     0,
		PageSize: 50,
	}

	response, err := serviceClient.GetNewsParticipants(context.Background(), &request)
	if err != nil {
		log.Fatalf("could not send GetNewsParticipants Request: %v - %v", status.Code(err), err)
	}

	spew.Dump(response)
}
