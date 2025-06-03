package grpc_client

import (
	"context"
	"log"
	"time"

	"github.com/Kisanlink/farmers-module/config"
	"github.com/kisanlink/protobuf/pb-aaa"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// UserClient is your global UserServiceClient
var UserClient pb.UserServiceClient

// unary interceptor to attach the AAA auth token and log
func ClientInterceptor(token string) grpc.UnaryClientInterceptor {
	return func(
		ctx context.Context,
		method string,
		req, reply interface{},
		cc *grpc.ClientConn,
		invoker grpc.UnaryInvoker,
		opts ...grpc.CallOption,
	) error {
		if token != "" {
			md := metadata.Pairs("aaa-auth-token", token)
			ctx = metadata.NewOutgoingContext(ctx, md)
		}
		log.Printf("→ gRPC %s", method)
		start := time.Now()
		err := invoker(ctx, method, req, reply, cc, opts...)
		log.Printf("← gRPC %s %v (took %s)", method,
			errOrOK(err), time.Since(start))
		return err
	}
}

func errOrOK(err error) string {
	if err != nil {
		return err.Error()
	}
	return "OK"
}

// GrpcClient dials the AAA host:port, installs interceptor, and
// initializes UserClient. No Greeter calls here.
func GrpcClient(token string) (*grpc.ClientConn, error) {
	// load .env or env vars
	config.LoadEnv()
	host := config.GetEnv("AAA_HOST")
	port := config.GetEnv("AAA_GRPC_PORT")
	addr := host + ":" + port

	conn, err := grpc.Dial(
		addr,
		grpc.WithInsecure(),
		grpc.WithBlock(),
		grpc.WithUnaryInterceptor(ClientInterceptor(token)),
	)
	if err != nil {
		return nil, err
	}

	// only initialize UserService client
	UserClient = pb.NewUserServiceClient(conn)
	return conn, nil
}

// Helper to fetch a user by ID (in your services/GetUserByIdClient)
func GetUserById(ctx context.Context, id string) (*pb.GetUserByIdResponse, error) {
	if UserClient == nil {
		return nil, status.Errorf(codes.FailedPrecondition, "gRPC client not initialized")
	}
	return UserClient.GetUserById(ctx, &pb.GetUserByIdRequest{Id: id})
}
