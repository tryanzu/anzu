package deps

import (
	"io"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

var (
	// The default value is a legacy bucket used in spartangeek.
	AwsS3Bucket = "spartan-board"
	// S3Endpoint for configurable endpoint (AWS S3, DigitalOcean Spaces, etc.)
	S3Endpoint = ""
	// S3Region for configurable region
	S3Region = "us-west-1"
)

func IgniteS3(container Deps) (Deps, error) {
	// Get configuration from environment variables
	if bucket := os.Getenv("AWS_S3_BUCKET"); bucket != "" {
		AwsS3Bucket = bucket
	}
	if endpoint := os.Getenv("AWS_S3_ENDPOINT"); endpoint != "" {
		S3Endpoint = endpoint
	}
	if region := os.Getenv("AWS_S3_REGION"); region != "" {
		S3Region = region
	}

	// Configure AWS session
	config := &aws.Config{
		Region: aws.String(S3Region),
	}

	// If custom endpoint is provided (for DigitalOcean Spaces, MinIO, etc.)
	if S3Endpoint != "" {
		config.Endpoint = aws.String(S3Endpoint)
		config.S3ForcePathStyle = aws.Bool(true)
	}

	// Use default credential chain (env vars, IAM roles, etc.)
	// AWS_ACCESS_KEY_ID and AWS_SECRET_ACCESS_KEY will be picked up automatically
	if accessKey := os.Getenv("AWS_ACCESS_KEY_ID"); accessKey != "" {
		secretKey := os.Getenv("AWS_SECRET_ACCESS_KEY")
		config.Credentials = credentials.NewStaticCredentials(accessKey, secretKey, "")
	}

	sess, err := session.NewSession(config)
	if err != nil {
		return container, err
	}

	service := s3.New(sess)
	container.S3Provider = &S3Service{
		client: service,
		bucket: AwsS3Bucket,
	}
	return container, nil
}

// S3Service wraps AWS SDK S3 client to provide similar interface to goamz
type S3Service struct {
	client *s3.S3
	bucket string
}

// GetBucketName returns the configured bucket name
func (s *S3Service) GetBucketName() string {
	return s.bucket
}

// PutObject uploads an object to S3
func (s *S3Service) PutObject(key string, data []byte, contentType string) error {
	_, err := s.client.PutObject(&s3.PutObjectInput{
		Bucket:      aws.String(s.bucket),
		Key:         aws.String(key),
		Body:        aws.ReadSeekCloser(strings.NewReader(string(data))),
		ContentType: aws.String(contentType),
	})
	return err
}

// GetObject retrieves an object from S3
func (s *S3Service) GetObject(key string) ([]byte, error) {
	result, err := s.client.GetObject(&s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, err
	}
	defer result.Body.Close()
	
	return io.ReadAll(result.Body)
}

// DeleteObject deletes an object from S3
func (s *S3Service) DeleteObject(key string) error {
	_, err := s.client.DeleteObject(&s3.DeleteObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})
	return err
}

// PutReader uploads an object to S3 from a reader
func (s *S3Service) PutReader(key string, reader io.ReadSeeker, size int64, contentType string) error {
	_, err := s.client.PutObject(&s3.PutObjectInput{
		Bucket:        aws.String(s.bucket),
		Key:           aws.String(key),
		Body:          reader,
		ContentType:   aws.String(contentType),
		ContentLength: aws.Int64(size),
	})
	return err
}

// GetURL returns the public URL for an object
func (s *S3Service) GetURL(key string) string {
	if S3Endpoint != "" {
		// For custom endpoints like DigitalOcean Spaces
		return S3Endpoint + "/" + s.bucket + "/" + key
	}
	// Default AWS S3 URL format
	return "https://s3-" + S3Region + ".amazonaws.com/" + s.bucket + "/" + key
}
