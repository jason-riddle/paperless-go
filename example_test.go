package paperless_test

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/jason-riddle/paperless-go"
)

func ExampleNewClient() {
	// Create a basic client
	client := paperless.NewClient("http://localhost:8000", "your-api-token")
	fmt.Printf("Client created: %T\n", client)
	// Output: Client created: *paperless.Client
}

func ExampleNewClient_withOptions() {
	// Create a client with custom timeout
	client := paperless.NewClient(
		"http://localhost:8000",
		"your-api-token",
		paperless.WithTimeout(10*time.Second),
	)
	fmt.Printf("Client created: %T\n", client)
	// Output: Client created: *paperless.Client
}

func ExampleClient_GetDocument() {
	client := paperless.NewClient("http://localhost:8000", "your-api-token")

	doc, err := client.GetDocument(context.Background(), 1)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Document: %s\n", doc.Title)
}

func ExampleClient_ListDocuments() {
	client := paperless.NewClient("http://localhost:8000", "your-api-token")

	// List first page of documents
	docs, err := client.ListDocuments(context.Background(), nil)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Found %d documents\n", docs.Count)
}

func ExampleClient_ListDocuments_withOptions() {
	client := paperless.NewClient("http://localhost:8000", "your-api-token")

	// Search for documents with filtering
	opts := &paperless.ListOptions{
		Query:    "invoice",
		Ordering: "-created",
		PageSize: 10,
	}

	docs, err := client.ListDocuments(context.Background(), opts)
	if err != nil {
		log.Fatal(err)
	}

	for _, doc := range docs.Results {
		fmt.Printf("Document: %s\n", doc.Title)
	}
}

func ExampleClient_GetTag() {
	client := paperless.NewClient("http://localhost:8000", "your-api-token")

	tag, err := client.GetTag(context.Background(), 1)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Tag: %s (%d documents)\n", tag.Name, tag.DocumentCount)
}

func ExampleClient_ListTags() {
	client := paperless.NewClient("http://localhost:8000", "your-api-token")

	tags, err := client.ListTags(context.Background(), nil)
	if err != nil {
		log.Fatal(err)
	}

	for _, tag := range tags.Results {
		fmt.Printf("Tag: %s\n", tag.Name)
	}
}

func ExampleIsNotFound() {
	client := paperless.NewClient("http://localhost:8000", "your-api-token")

	doc, err := client.GetDocument(context.Background(), 999)
	if paperless.IsNotFound(err) {
		fmt.Println("Document not found")
		return
	}
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Document: %s\n", doc.Title)
}
