package client

import (
	"context"
	"fmt"
	"time"

	pb "bff-service/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type VMClient struct {
	client pb.VMServiceClient
	conn *grpc.ClientConn
}

func NewClient(grpcAddr string) (*VMClient, error) {
	conn, err := grpc.NewClient(grpcAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("main-service gRPC bağlantı hatası: %w", err)
	}

	return &VMClient{
		client: pb.NewVMServiceClient(conn),
		conn: conn,
	}, nil
}

func (c *VMClient) Close() {
	if c.conn != nil {
		c.conn.Close()
	}
}

func (c *VMClient) ListVMs(ctx context.Context) (*pb.ListVMsResponse, error) {
	ctxTimeout, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	return c.client.ListVMs(ctxTimeout, &pb.ListVMsRequest{})
}

func (c *VMClient) VMAction(ctx context.Context, vmID int32) (*pb.VMActionResponse, error) {
	ctxTimeout, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	return c.client.StartVM(ctxTimeout, &pb.VMActionRequest{VmId: vmID})
}
