package gaming

import (
	"github.com/tryanzu/core/modules/feed"
)

type Post struct {
	post *feed.Post
	di   *Module
}

// Review gaming facts of a post
func (self *Post) Review() {

	//data := self.post.Data()
	//reached, viewed := self.post.GetReachViews()

}
