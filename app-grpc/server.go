package appgrpc

import (
	"context"
	"fmt"
	"gbb.go/gvp/dao"
	"gbb.go/gvp/model"
	"gbb.go/gvp/proto/grpcXVPPb"
	"gbb.go/gvp/utils"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
	"log"
	"net"
	"os"
	"strconv"
)

const (
	GRPC_CRED_AUTH       = "authorization"
	GRPC_CTX_KEY_SESSION = "GRPC_CTX_KEY_SESSION"
)

var (
	GRPCAccessDeniedErr        = status.Errorf(codes.PermissionDenied, "Access Denied")
	GRPCUnauthenticateUserdErr = status.Errorf(codes.PermissionDenied, "Invalid User")
	GRPCInvalidJWTErr          = status.Errorf(codes.PermissionDenied, "Invalid JWT")
)

func StartServer(grpcPort int) (net.Listener, *grpc.Server, error) {
	listener, err := net.Listen("tcp", "0.0.0.0:"+strconv.Itoa(grpcPort))
	if err != nil {
		log.Fatalf("Failed to listen grpc server: %v", err)
	}

	opts := []grpc.ServerOption{}
	grpcAuthentication, _ := strconv.ParseBool(os.Getenv("GRPC_AUTHENTICATION"))
	fmt.Println("GRPC Authentication", grpcAuthentication)
	if grpcAuthentication {
		opts = append(opts, grpc.StreamInterceptor(streamInterceptor), grpc.UnaryInterceptor(unaryInterceptor))
	}

	tls, _ := strconv.ParseBool(os.Getenv("GRPC_TLS"))

	fmt.Println("GRPC TLS", tls)

	if tls {
		certFile := "./ssl/server.crt"
		keyFile := "./ssl/server.pem"
		creds, sslErr := credentials.NewServerTLSFromFile(certFile, keyFile)
		if sslErr != nil {
			log.Fatalf("Failed to loading grpc certificate: %v", err)
		}
		//spew.Dump(creds)
		opts = append(opts, grpc.Creds(creds))
	}
	grpcServer := grpc.NewServer(opts...)

	xvpGRPCService := XVPGRPCService{}
	grpcXVPPb.RegisterXVPServiceServer(grpcServer, &xvpGRPCService)

	//Register reflection service on gRPC server, so that evan CLI can interact with server without .proto files
	reflection.Register(grpcServer)

	go func() {
		log.Printf("Starting grpc server on: %v\n", grpcPort)
		if err := grpcServer.Serve(listener); err != nil {
			log.Fatalf("Failed to serve grpc: %v", err)
		}
	}()

	return listener, grpcServer, nil
}

func authorize(ctx context.Context) (*model.User, error) {
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		//spew.Dump(md)
		if len(md[GRPC_CRED_AUTH]) > 0 {
			jwtToken := md[GRPC_CRED_AUTH][0]

			if len(jwtToken) > 0 {

				claims, err := utils.ParseJWTToken(jwtToken)

				if err != nil {
					log.Println("GRPC - Invalid jwtToken", err)
					return nil, GRPCInvalidJWTErr
				}

				jwtUsername := claims.Subject

				user, err := dao.GetUserDAO().FindByUserName(ctx, jwtUsername)
				if err != nil {
					return nil, GRPCUnauthenticateUserdErr
				}

				return user, nil
			}
		}
	}

	grpcDummyTestMode, _ := strconv.ParseBool(os.Getenv("GRPC_DUMMY_TEST_MODE"))
	grpcDummyUserName := os.Getenv("GRPC_DUMMY_USER")
	if grpcDummyTestMode && len(grpcDummyUserName) > 0 {
		log.Println("Using GRPC DummyUser", grpcDummyUserName)

		user, err := dao.GetUserDAO().FindByUserName(ctx, grpcDummyUserName)
		if err != nil {
			log.Println("Using GRPC DummyUser Err", err)
			return nil, err
		}

		return user, nil
	}

	p, success := peer.FromContext(ctx)
	if !success {
		log.Println("Cannot check client ip")
		return nil, GRPCAccessDeniedErr
	}

	return &model.User{
		Username:  p.Addr.String(),
		Email:     "",
		Role:      grpcXVPPb.USER_ROLE_GUEST,
		Password:  "",
		CreatedAt: utils.UTCNowMilli(),
	}, nil

}

func unaryInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	user, err := authorize(ctx)
	if err != nil {
		return nil, err
	}

	//if user.Role == grpcXVPPb.USER_ROLE_GUEST {
	//	log.Println("Call grpc as guest", user.Username, reflect.TypeOf(req))
	//}
	//spew.Dump(req)

	switch req.(type) {
	case
		//user
		*grpcXVPPb.CreatAccountRequest,
		*grpcXVPPb.VerifyAuthencodeRequest,
		*grpcXVPPb.LoginRequest,
		*grpcXVPPb.GetUserInfoRequest,

		//category
		*grpcXVPPb.GetAllCategoryRequest,
		*grpcXVPPb.GetCategoryRequest,

		//news
		*grpcXVPPb.GetListNewsRequest,
		*grpcXVPPb.GetNewsRequest,
		*grpcXVPPb.GetListTopNewsRequest,

		//files info
		*grpcXVPPb.GetFilePresignedUrlRequest,

		//comments
		*grpcXVPPb.GetNewsCommentsRequest,

		//news tags
		*grpcXVPPb.GetNewsTagsRequest,

		//news participants
		*grpcXVPPb.GetNewsParticipantsRequest,

		//testing
		*grpcXVPPb.TestCreateUserRequest:
		//log.Println("unaryInterceptor: bypass authorization check")
	case
		*grpcXVPPb.CreateCategoryRequest,
		*grpcXVPPb.UpdateCategoryRequest,
		*grpcXVPPb.DeleteCategoryRequest:
		if !user.IsModeratorPermission() {
			log.Println("NOT MODERATOR PERMISSION!")
			return nil, GRPCAccessDeniedErr
		}
	case
		*grpcXVPPb.CreateNewsRequest,
		*grpcXVPPb.UpdateNewsInfoRequest,
		*grpcXVPPb.DeleteNewsRequest,
		grpcXVPPb.XVPService_UploadNewsPreviewImageServer,
		*grpcXVPPb.DeleteNewsPreviewImageRequest,
		grpcXVPPb.XVPService_UploadNewsMediaServer,
		*grpcXVPPb.DeleteNewsMediaRequest:
		if !user.IsContentUserPermission() {
			log.Println("NOT CONTENT PERMISSION!")
			return nil, GRPCAccessDeniedErr
		}
	default:
		if !user.IsNotGuest() {
			log.Println("NOT ACCEPT GUEST USER!")
			return nil, GRPCAccessDeniedErr
		}
	}

	grpcSession := &GrpcSession{
		User: user,
	}

	ctxInterceptor := context.WithValue(ctx, GRPC_CTX_KEY_SESSION, grpcSession)

	return handler(ctxInterceptor, req)
}

func streamInterceptor(srv interface{}, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
	user, err := authorize(stream.Context())
	if err != nil {
		return err
	}

	//log.Println("streamInterceptor", info.FullMethod)

	switch info.FullMethod {
	case "/grpcXVPPb.XVPService/UploadNewsPreviewImage",
		"/grpcXVPPb.XVPService/UploadNewsMedia":
		if !user.IsContentUserPermission() {
			log.Println("NOT CONTENT PROVIDER PERMISSION!")
			return GRPCAccessDeniedErr
		}
	default:
		if !user.IsNotGuest() {
			log.Println("NOT ACCEPT GUEST USER!")
			return GRPCAccessDeniedErr
		}
	}

	return handler(srv, newWrappedStream(stream, user))
}

// wrappedStream wraps around the embedded grpc.ServerStream, and intercepts the RecvMsg and
// SendMsg method call.
type wrappedStream struct {
	grpc.ServerStream
	WrappedContext context.Context
}

func (w *wrappedStream) RecvMsg(m interface{}) error {
	//log.Println("Receive a message (Type: %T) at %s", m, time.Now().Format(time.RFC3339))
	return w.ServerStream.RecvMsg(m)
}

func (w *wrappedStream) SendMsg(m interface{}) error {
	//log.Println("Send a message (Type: %T) at %v", m, time.Now().Format(time.RFC3339))
	return w.ServerStream.SendMsg(m)
}

func (w *wrappedStream) Context() context.Context {
	return w.WrappedContext
}

func newWrappedStream(stream grpc.ServerStream, user *model.User) grpc.ServerStream {
	grpcSession := &GrpcSession{
		User: user,
	}
	wrappedContext := context.WithValue(stream.Context(), GRPC_CTX_KEY_SESSION, grpcSession)
	return &wrappedStream{stream, wrappedContext}
}
