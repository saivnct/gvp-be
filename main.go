package main

import (
	"context"
	"flag"
	"fmt"
	appgrpc "gbb.go/gvp/app-grpc"
	apphttp "gbb.go/gvp/app-http"
	appmail "gbb.go/gvp/app-mail"
	"gbb.go/gvp/appws"
	"gbb.go/gvp/dao"
	"gbb.go/gvp/pubsub"
	"gbb.go/gvp/utils"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"google.golang.org/grpc"
	"log"
	"math/rand"
	"net"
	"os"
	"os/signal"
	"strconv"
	"time"
)

func main() {
	// if we crash the go code, we get the file and line number
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	err := godotenv.Load()
	if err != nil {
		log.Fatal(err)
	}

	grpcPort, err := strconv.Atoi(os.Getenv("GRPC_PORT"))
	if err != nil {
		log.Fatal(err)
	}

	httpPort, err := strconv.Atoi(os.Getenv("HTTP_PORT"))
	if err != nil {
		log.Fatal(err)
	}
	httpSessionSecret := os.Getenv("HTTP_SESSION_SECRET")

	listenGRPC := flag.Int("g", grpcPort, "wait for incoming connections")
	listenHTTP := flag.Int("h", httpPort, "wait for incoming connections")
	flag.Parse()

	grpcPort = *listenGRPC
	httpPort = *listenHTTP
	fmt.Println("grpcPort", grpcPort)
	fmt.Println("httpPort", httpPort)

	//connect to mongodb
	fmt.Println("Connecting to mongodb ...")
	db := dao.GetDataBase()

	//connect to redis
	fmt.Println("Connecting to redis ...")
	cache := dao.GetCache()
	// test redis connection
	redisCtx, cancelRedis := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancelRedis()
	_, err = cache.RedisClient.Ping(redisCtx).Result()
	if err != nil {
		log.Fatalf("Failed to connect to redis: %v", err)
	}

	listener, _, grpcServer, err := initServices(httpPort, httpSessionSecret, grpcPort)
	if err != nil {
		log.Fatalf("Failed to init services: %v", err)
	}

	//give random seed
	rand.Seed(utils.UTCNowMilli())

	//Setup shutdown hook
	// Wait for Ctrl+C to exit
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt)

	//Block until a signal is received
	<-ch

	fmt.Println("Stopping grpc server")
	grpcServer.Stop()
	fmt.Println("Stopping grpc listener")
	_ = listener.Close()

	fmt.Println("Closing mongodb connection")
	mongoCtx, cancelMongo := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancelMongo()
	_ = db.MongoClient.Disconnect(mongoCtx)
	fmt.Println("End of programm")
}

func initServices(httpPort int, httpSessionSecret string, grpcPort int) (net.Listener, *gin.Engine, *grpc.Server, error) {

	listener, grpcServer, err := appgrpc.StartServer(grpcPort)
	if err != nil {
		return nil, nil, nil, err
	}

	ws := appws.GetWS()
	ws.Start()

	router, err := apphttp.StartServer(httpPort, httpSessionSecret, ws.Server)
	if err != nil {
		return nil, nil, nil, err
	}

	pubsub.StartSubscribe()

	//mail service
	smtpHost := os.Getenv("SMTP_HOST")
	smtpPort, err := strconv.Atoi(os.Getenv("SMTP_PORT"))
	if err != nil {
		log.Fatal(err)
	}
	smtpUsername := os.Getenv("SMTP_USERNAME")
	smtpPass := os.Getenv("SMTP_PASS")
	smtpAddress := os.Getenv("SMTP_ADDRESS")
	//log.Println(smtpHost, smtpPort, smtpUsername, smtpPass, smtpAddress)
	mailSv := appmail.GetMailService()
	mailSv.Start(smtpHost, smtpPort, smtpUsername, smtpPass, smtpAddress)

	return listener, router, grpcServer, nil
}
