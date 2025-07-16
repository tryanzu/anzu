package posts

import (
	"github.com/olebedev/config"
	"github.com/tryanzu/core/deps"
	"github.com/tryanzu/core/modules/acl"
	"github.com/tryanzu/core/modules/exceptions"
	"github.com/tryanzu/core/modules/feed"
	"github.com/tryanzu/core/modules/gaming"

	"regexp"
)

var legalSlug = regexp.MustCompile(`^([a-zA-Z0-9\-\.|/]+)$`)

type API struct {
	Feed   *feed.FeedModule             `inject:""`
	Acl    *acl.Module                  `inject:""`
	Gaming *gaming.Module               `inject:""`
	Errors *exceptions.ExceptionsModule `inject:""`
	Config *config.Config               `inject:""`
	S3     *deps.S3Service              `inject:""`
}
