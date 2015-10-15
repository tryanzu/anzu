package feed

type LightPost struct {
	di   *FeedModule
	data LightPostModel
}

func (post *LightPost) Data() LightPostModel {

	return post.data
}
