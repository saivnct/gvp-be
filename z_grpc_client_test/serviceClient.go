package z_grpc_client_test

import (
	"fmt"
	appgrpc "gbb.go/gvp/app-grpc"
	"gbb.go/gvp/proto/grpcXVPPb"
	"github.com/joho/godotenv"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"log"
	"os"
	"strconv"
)

func GetServiceClient(authorization string) (*grpc.ClientConn, grpcXVPPb.XVPServiceClient, error) {
	err := godotenv.Load()
	if err != nil {
		log.Fatal(err)
	}
	tls, _ := strconv.ParseBool(os.Getenv("GRPC_TLS"))
	fmt.Println("GRPC TLS", tls)

	opts := []grpc.DialOption{}

	if tls {
		certFile := "./ssl/ca.crt" // Certificate Authority Trust certificate
		creds, sslErr := credentials.NewClientTLSFromFile(certFile, "")
		if sslErr != nil {
			log.Fatalf("Error while loading CA trust certificate: %v", sslErr)
			return nil, nil, sslErr
		}
		opts = append(opts, grpc.WithTransportCredentials(creds))
	} else {
		opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	}

	if len(authorization) > 0 {
		opts = append(opts, grpc.WithPerRPCCredentials(&appgrpc.GrpcLoginCreds{
			Authorization: authorization,
		}))
	}

	clientConn, err := grpc.Dial("localhost:"+os.Getenv("GRPC_PORT"), opts...)
	//clientConn, err := grpc.Dial("grpc.dzunu.com:"+os.Getenv("GRPC_PORT"), opts...)
	if err != nil {
		log.Fatalf("could not connect: %v", err)
		return nil, nil, err
	}

	serviceClient := grpcXVPPb.NewXVPServiceClient(clientConn)

	return clientConn, serviceClient, nil
}
