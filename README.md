# downscoped-gcloud-token
Downscoped Google Cloud's Secure Token

This program demonstrates how to downscope permissions for accessing Google Cloud Storage objects using a root access token. 
It generates a downscoped token with restricted permissions (such as read-only access to specific objects) based on an access boundary and then uses this token to interact with the Cloud Storage bucket, allowing you to read from and write to specific objects while maintaining limited access.

## Run the example program

The example program checks if write (create object) access has been revoked while ensuring that read access is still available. 
It uses downscoped tokens to enforce fine-grained access control, limiting permissions for reading from a specific object, while also attempting to write to the object to verify if that permission has been properly revoked.

1. Create bucket (make sure you have read and write permissions) and export bucket name:

```
$ export BUCKET_NAME="<my_bucket_name>"
```

2. Create object in the bucket:

```
$ echo "my txt file" >> txt.txt
$ gcloud storage cp ./txt.txt gs://<my_bucket_name>/f291d062-e376-4cf9-99dd-19d33176f158/txt.txt
```

3. Run the program:

```
go run . 
```