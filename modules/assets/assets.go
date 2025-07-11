package assets

import (
	"context"
	"encoding/base64"
	"net/http"
	"path/filepath"
	"time"

	"github.com/mitchellh/goamz/s3"
	"github.com/tryanzu/core/deps"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func Boot() *Module {

	module := &Module{}

	return module
}

type Module struct {
	S3 *s3.Bucket `inject:""`
}

func (module *Module) UploadBase64(content, filename, related string, related_id primitive.ObjectID, meta interface{}) error {
	ctx := context.Background()
	data, err := base64.StdEncoding.DecodeString(content)
	if err != nil {
		return err
	}

	extension := filepath.Ext(filename)
	random := primitive.NewObjectID().Hex()

	// Detect the downloaded file type
	dataType := http.DetectContentType(data)

	// S3 path
	path := related + "/" + random + extension

	// Upload binary to s3
	err = module.S3.Put(path, data, dataType, s3.ACL("public-read"))
	if err != nil {
		return err
	}

	database := deps.Container.Mgo()
	asset := &Asset{
		Related:   related,
		RelatedId: related_id,
		Path:      path,
		Meta:      meta,
		Created:   time.Now(),
	}

	collection := database.Collection("assets")
	_, err = collection.InsertOne(ctx, asset)
	if err != nil {
		return err
	}

	return nil
}
