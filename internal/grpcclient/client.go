package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	pb "github.com/developerc/GophKeeper/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// main запускает клиента gRPC
func main() {
	addr := "localhost:5000"
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	// создадим клиент grpc с перехватчиком
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Println("could not connect to grpc server: ", err)
		os.Exit(1)
	}
	defer conn.Close()

	grpcClient := pb.NewGrpcServiceClient(conn)

	authorizedResponse, err := grpcClient.CreateUser(ctx, &pb.UserRegisterRequest{Login: "myLogin", Password: "myPassword"})
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(authorizedResponse.Token)

}
