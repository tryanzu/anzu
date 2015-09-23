package gaming

import (
	"github.com/fernandez14/spartangeek-blacker/modules/feed"
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
