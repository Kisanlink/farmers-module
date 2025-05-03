package grpc_client

import (
	"context"
	"sync"
	"time"

	"github.com/Kisanlink/farmers-module/config"
	"github.com/Kisanlink/farmers-module/utils"
	"github.com/kisanlink/protobuf/pb-aaa"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/metadata"
)

var (
	UserClient pb.UserServiceClient
	conn       *grpc.ClientConn
	once       sync.Once
)

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

		utils.Log.Debugf("Sending request to method: %s", method)
		start := time.Now()
		err := invoker(ctx, method, req, reply, cc, opts...)
		if err != nil {
			utils.Log.Errorf("Request failed: %v (duration: %s)", err, time.Since(start))
			return err
		}
		utils.Log.Debugf("Request succeeded to method: %s (duration: %s)", method, time.Since(start))
		return nil
	}
}

func InitGrpcClient(token string) (pb.UserServiceClient, error) {
	var err error
	once.Do(func() {
		config.LoadEnv()
		aaa_host := config.GetEnv("AAA_HOST")
		aaa_grpc_port := config.GetEnv("AAA_GRPC_PORT")
		connection := aaa_host + ":" + aaa_grpc_port

		clientInterceptor := ClientInterceptor(token)
		creds := credentials.NewClientTLSFromCert(nil, "") // Use system CA pool
		conn, err = grpc.Dial(connection,
			grpc.WithTransportCredentials(creds),
			grpc.WithBlock(),
			grpc.WithUnaryInterceptor(clientInterceptor),
		)
		if err != nil {
			utils.Log.Errorf("Failed to connect to gRPC server: %v", err)
			return
		}
		utils.Log.Infof("Connected to gRPC server at %s", connection)
		UserClient = pb.NewUserServiceClient(conn)
	})
	return UserClient, err
}
