package grpc_client

import (
	"context"
	"log"
	"time"

	"github.com/kisanlink/protobuf/pb-aaa"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	
)

var UserClient pb.UserServiceClient

func ClientInterceptor(token string) grpc.UnaryClientInterceptor {
	return func(
		ctx context.Context,
		method string,
		req interface{},
		reply interface{},
		cc *grpc.ClientConn,
		invoker grpc.UnaryInvoker,
		opts ...grpc.CallOption,
	) error {
			if token != "" {
			md := metadata.Pairs("aaa-auth-token", token)
			ctx = metadata.NewOutgoingContext(ctx, md)
		}

		// Log the request
		log.Printf("Sending request to method: %s", method)
		start := time.Now()
		err := invoker(ctx, method, req, reply, cc, opts...)
		if err != nil {
			log.Printf("Request failed: %v (duration: %s)", err, time.Since(start))
			return err
		}
		log.Printf("Request succeeded to method: %s (duration: %s)", method, time.Since(start))
		return nil
	}
}


func GrpcClient(token string) (*grpc.ClientConn, error) {
var clientInterceptor grpc.UnaryClientInterceptor

    clientInterceptor = ClientInterceptor(token)


conn, err := grpc.Dial("localhost:50052", grpc.WithInsecure(), grpc.WithBlock(), grpc.WithUnaryInterceptor(clientInterceptor))
if err != nil {
    log.Fatalf("failed to connect to gRPC server: %v", err)
}

	
	UserClient = pb.NewUserServiceClient(conn)
	client := pb.NewGreeterClient(conn)

	request := &pb.HelloRequest{
		Name: "World",
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	response, err := client.SayHello(ctx, request)
	if err != nil {
		log.Fatalf("Failed to call SayHello: %v", err)
	}
	log.Printf("Response from Greeter service: %s", response.GetMessage())

	return conn, nil
}