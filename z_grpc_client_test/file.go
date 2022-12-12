package z_grpc_client_test

import (
	"bytes"
	"context"
	"crypto/md5"
	"fmt"
	"gbb.go/gvp/fileHandler"
	"gbb.go/gvp/proto/grpcXVPPb"
	"github.com/davecgh/go-spew/spew"
	"google.golang.org/grpc/status"
	"io"
	"log"
)

func DownloadFile() {
	jwt := Login()

	cc, serviceClient, _ := GetServiceClient(jwt)

	defer cc.Close() //this will call at very end of code

	req := &grpcXVPPb.DownloadFileRequest{
		FileId: "5f5f1b64-608e-46e3-b8b6-7fbedf43156a",
	}

	stream, err := serviceClient.DownloadFile(context.Background(), req)
	if err != nil {
		log.Fatal("cannot download file: ", err)
	}

	res, err := stream.Recv()
	fileInfo := res.GetFileInfo()

	fileData := bytes.Buffer{}
	fileSize := 0

	for {
		//log.Println("waiting to receive more media data")

		res, err := stream.Recv()
		if err == io.EOF {
			log.Println("received file data from client")
			break
		}
		if err != nil {
			log.Fatal("Cannot receive chunk data: ", err)
		}

		chunk := res.GetChunkData()
		size := len(chunk)

		//log.Printf("received a chunk with size: %d\n", size)

		fileSize += size
		_, err = fileData.Write(chunk)
		if err != nil {
			log.Fatal("Cannot write chunk data: ", err)
		}
	}

	checksum := fmt.Sprintf("%x", md5.Sum(fileData.Bytes()))
	if checksum != fileInfo.Checksum {
		log.Fatal("Invalid checksum download: ", checksum, fileInfo.Checksum)
	}

	diskFileStore := fileHandler.NewDiskFileStore("./z_grpc_client_test/media")
	_, err = diskFileStore.SaveToDiskV2(fileInfo.FileName, fileData)

	log.Println("Done download file", fileSize, checksum)
	if err != nil {
		log.Fatal("cannot save file to the store: ", err)
	}
}

func GetFilePresignedUrl() {
	cc, serviceClient, _ := GetServiceClient("")

	defer cc.Close() //this will call at very end of code

	request := grpcXVPPb.GetFilePresignedUrlRequest{
		FileId: "4b1df48b-af19-4cbe-817c-65cd09aef613-3e9d2073f3719a083eaef04997432b43",
	}

	response, err := serviceClient.GetFilePresignedUrl(context.Background(), &request)
	if err != nil {
		log.Fatalf("could not send GetFilePresignedUrl Request: %v - %v", status.Code(err), err)
	}

	spew.Dump(response)
}

func GetFileInfo() {
	jwt := Login()

	cc, serviceClient, _ := GetServiceClient(jwt)

	defer cc.Close() //this will call at very end of code

	request := grpcXVPPb.GetFileInfolRequest{
		FileId: "d2044b9f-1c0d-49e5-9ddc-7c22ac63b97b-3e9d2073f3719a083eaef04997432b43",
	}

	response, err := serviceClient.GetFileInfo(context.Background(), &request)
	if err != nil {
		log.Fatalf("could not send GetFilePresignedUrl Request: %v - %v", status.Code(err), err)
	}

	spew.Dump(response)
}
