package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"testing"

	"google.golang.org/grpc"
	"google.golang.org/grpc/test/bufconn"

	pb "example.com/m/helloworld"
)

func dialer() func(context.Context, string) (net.Conn, error) {
	const bufSize = 1024 * 1024
	var lis *bufconn.Listener = bufconn.Listen(bufSize)

	s := grpc.NewServer()
	pb.RegisterGreeterServer(s, &server{})
	go func() {
		if err := s.Serve(lis); err != nil {
			log.Fatalf("Server exited with error: %v", err)
		}
	}()
	return func(context.Context, string) (net.Conn, error) {
		return lis.Dial()
	}
}

func TestSayHello(t *testing.T) {
	ctx := context.Background()
	conn, err := grpc.DialContext(ctx, "bufnet", grpc.WithContextDialer(dialer()), grpc.WithInsecure())
	if err != nil {
		t.Fatalf("Failed to dial bufnet: %v", err)
	}
	defer conn.Close()
	client := pb.NewGreeterClient(conn)

	stream, err := client.SayHello(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	waitc := make(chan struct{})
	go func() {
		to_receive := 10
		received := 0
		for {
			in, err := stream.Recv()
			received += 1
			if err == io.EOF {
				// read done.
				if received != to_receive {
					log.Fatal("received ", received, to_receive)
				}
				close(waitc)
				return
			}
			if err != nil {
				log.Fatalf("Failed to receive a note : %v", err)
			}
			log.Printf("Got message %s", in.Name)
		}
	}()
	for i := 0; i < 10; i++ {
		if err := stream.Send(&pb.HelloRequest{Name: "aus" + fmt.Sprint(i)}); err != nil {
			log.Fatalf("Failed to send a note: %v", err)
		}
	}
	stream.CloseSend()
	<-waitc

	// Test for output here.
}
