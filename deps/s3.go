package deps

import (
	"github.com/mitchellh/goamz/aws"
	"github.com/mitchellh/goamz/s3"
)

var (
	// The default value is a legacy bucket used in spartangeek.
	AwsS3Bucket = "spartan-board"
)

func IgniteS3(container Deps) (Deps, error) {
	auth, err := aws.GetAuth("", "")
	if err != nil {
		return container, err
	}

	service := s3.New(auth, aws.USWest)
	container.S3Provider = service.Bucket(AwsS3Bucket)
	return container, nil
}
