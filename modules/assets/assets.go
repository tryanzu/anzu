package assets

import(
	"github.com/mitchellh/goamz/s3"
	"github.com/fernandez14/spartangeek-blacker/mongo"
	"gopkg.in/mgo.v2/bson"

	"net/http"
	"encoding/base64"
	"time"
	"path/filepath"
)

func Boot() *Module {

	module := &Module{}

	return module
}

type Module struct {
	Mongo *mongo.Service `inject:""`
	S3    *s3.Bucket `inject:""`
}

func (module *Module) UploadBase64(content, filename, related string, related_id bson.ObjectId, meta interface{}) error {

	data, err := base64.StdEncoding.DecodeString(content)

	if err != nil {
		return err
	}

	extension := filepath.Ext(filename)
	random := bson.NewObjectId().Hex()

	// Detect the downloaded file type
	dataType := http.DetectContentType(data)

	// S3 path
	path := related + "/" + random + extension

	// Upload binary to s3
	err = module.S3.Put(path, data, dataType, s3.ACL("public-read"))

	if err != nil {
		return err
	}

	database := module.Mongo.Database
	asset := &Asset{
		Related: related,
		RelatedId: related_id,
		Path: path,
		Meta: meta,
		Created: time.Now(),
	}

	err = database.C("assets").Insert(asset)

	if err != nil {
		return err
	}

	return nil
}