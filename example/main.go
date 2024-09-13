package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"

	"cloud.google.com/go/storage"
	downscoped "github.com/mariiatuzovska/downscoped-gcloud-token"
	"golang.org/x/oauth2"
	"google.golang.org/api/option"
)

var bucketName = os.Getenv("BUCKET_NAME")

func main() {
	ctx := context.Background()

	token, err := downscoped.NewDownscopedToken(ctx, []downscoped.AccessBoundaryRule{
		{
			AvailableResource: "//storage.googleapis.com/projects/_/buckets/" + bucketName,
			AvailablePermissions: []string{
				"inRole:roles/storage.objectViewer",
			},
			AvailabilityCondition: downscoped.Condition{
				Title:      "obj-prefixes",
				Expression: fmt.Sprintf("resource.name.startsWith(\"projects/_/buckets/%s/objects/f291d062-e376-4cf9-99dd-19d33176f158/\")", bucketName),
			},
		},
	})

	sts := oauth2.StaticTokenSource(token)

	storageClient, err := storage.NewClient(ctx, option.WithTokenSource(sts))
	if err != nil {
		log.Fatalf("Could not create storage Client: %v", err)
	}
	defer storageClient.Close()

	var objectName string = "f291d062-e376-4cf9-99dd-19d33176f158/txt.txt"
	var newObjectName string = "f291d062-e376-4cf9-99dd-19d33176f158/txt-1.txt"

	bkt := storageClient.Bucket(bucketName)
	obj := bkt.Object(objectName)
	r, err := obj.NewReader(ctx)
	if err != nil {
		log.Fatalf("NewReader: %v", err)
	}
	defer r.Close()

	data, err := io.ReadAll(r)
	if err != nil {
		log.Fatalf("Reading from object failed: %v", err)
	}
	log.Printf("Reading from %v object: %v\n", objectName, string(data))

	// try to write to the object
	w := bkt.Object(newObjectName).NewWriter(ctx)
	_, err = w.Write([]byte("Hello World"))
	if err != nil {
		log.Fatalf("Write an object %v failed: %v", newObjectName, err)
	}
	if err := w.Close(); err != nil {
		// expected error
		log.Println("Save an object %v failed: %v", newObjectName, err)
	}
}
