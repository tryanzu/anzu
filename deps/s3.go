package deps

import (
	"github.com/mitchellh/goamz/aws"
	"github.com/mitchellh/goamz/s3"
)

func IgniteS3(container Deps) (Deps, error) {
	cnf := container.Config()
	auth, err := aws.GetAuth(cnf.UString("amazon.access_key", ""), cnf.UString("amazon.secret", ""))
	if err != nil {
		return container, err
	}

	service := s3.New(auth, aws.USWest)
	container.S3Provider = service.Bucket(cnf.UString("amazon.s3.bucket", ""))

	return container, nil
}
