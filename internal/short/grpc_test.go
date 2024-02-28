package short

import (
	"context"
	pb "github.com/1tn-pw/protobufs/generated/short_service/v1"
	"testing"
)

// Initialize our server to be available for all tests
var server = Server{}

// TestCreateURL tests the CreateURL method of our grpc server
func TestCreateURL(t *testing.T) {
	// Define the input and expected outcome
	input := &pb.CreateURLRequest{Url: "testUrl"}
	expected := "expectedShortUrl"

	// Call the function with our input
	response, err := server.CreateURL(context.Background(), input)

	// Perform test for error
	if err != nil {
		t.Errorf("Error was expected to be nil but got %v", err)
	}

	// Perform test for returned result
	if response.ShortUrl != expected {
		t.Errorf("Expected %s but got %s", expected, response.ShortUrl)
	}
}

// TestGetURL tests the GetURL method of our grpc server
func TestGetURL(t *testing.T) {
	// Define the input and expected outcome
	input := &pb.GetURLRequest{ShortUrl: "testShortUrl"}
	expected := "expectedUrl"

	// Call the function with our input
	response, err := server.GetURL(context.Background(), input)

	// Perform test for error
	if err != nil {
		t.Errorf("Error was expected to be nil but got %v", err)
	}

	// Perform test for returned result
	if response.Url != expected {
		t.Errorf("Expected %s but got %s", expected, response.Url)
	}
}
