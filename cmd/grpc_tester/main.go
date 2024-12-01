package main

import (
	"context"
	cryptov1 "github.com/1337Bart/smol-crypto-api/api/proto/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"log"
	"time"
)

// todo - move it to a test
func main() {
	// Connect to the server
	conn, err := grpc.Dial("localhost:50051",
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer conn.Close()

	// Create a client
	client := cryptov1.NewCryptoServiceClient(conn)

	// Context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	// Test case 1: Default pagination
	resp1, err := client.ListCryptos(ctx, &cryptov1.ListCryptosRequest{})
	if err != nil {
		log.Fatalf("ListCryptos failed: %v", err)
	}
	log.Printf("Default pagination response: %+v\n", resp1)
	log.Printf("Number of cryptos: %d\n", len(resp1.Cryptos))
	log.Printf("Current page: %d\n", resp1.CurrentPage)
	log.Printf("Total count: %d\n", resp1.TotalCount)

	// Test case 2: Custom pagination
	resp2, err := client.ListCryptos(ctx, &cryptov1.ListCryptosRequest{
		Pagination: &cryptov1.PaginationRequest{
			Page:  2,
			Limit: 5,
		},
	})
	if err != nil {
		log.Fatalf("ListCryptos failed: %v", err)
	}
	log.Printf("\nCustom pagination response: %+v\n", resp2)
	log.Printf("Number of cryptos: %d\n", len(resp2.Cryptos))
	log.Printf("Current page: %d\n", resp2.CurrentPage)
	log.Printf("Total count: %d\n", resp2.TotalCount)
}
