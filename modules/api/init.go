package api

import (
	"fmt"
	"github.com/brandfolder/gin-gorelic"
	"github.com/facebookgo/inject"
	"github.com/fernandez14/spartangeek-blacker/handle"
	"github.com/gin-gonic/gin"
	"os"
)

type Module struct {
	Posts       handle.PostAPI
	Votes       handle.VoteAPI
	Users       handle.UserAPI
	Categories  handle.CategoryAPI
	Elections   handle.ElectionAPI
	Comments    handle.CommentAPI
	Parts       handle.PartAPI
	Stats       handle.StatAPI
	Middlewares handle.MiddlewareAPI
	Collector   handle.CollectorAPI
	Sitemap     handle.SitemapAPI
	Acl         handle.AclAPI
	Gaming      handle.GamingAPI
}

func (module *Module) Populate(g inject.Graph) {

	err := g.Provide(
		&inject.Object{Value: &module.Collector},
		&inject.Object{Value: &module.Posts},
		&inject.Object{Value: &module.Votes},
		&inject.Object{Value: &module.Users},
		&inject.Object{Value: &module.Categories},
		&inject.Object{Value: &module.Elections},
		&inject.Object{Value: &module.Comments},
		&inject.Object{Value: &module.Parts},
		&inject.Object{Value: &module.Stats},
		&inject.Object{Value: &module.Middlewares},
		&inject.Object{Value: &module.Acl},
		&inject.Object{Value: &module.Sitemap},
		&inject.Object{Value: &module.Gaming},
	)

	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	// Populate the DI with the instances
	if err := g.Populate(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func (module *Module) Run() {

	// Start gin classic middlewares
	router := gin.Default()

	// Start gorelic
	gorelic.InitNewrelicAgent("3e8e387fb7b29dedb924db3ba88e2790599bd0fb", "Blacker", false)

	// Middlewares setup
	router.Use(gorelic.Handler)
	router.Use(module.Middlewares.ErrorTracking())
	router.Use(module.Middlewares.CORS())
	router.Use(module.Middlewares.MongoRefresher())
	router.Use(module.Middlewares.StatsdTiming())

	// Sitemap generator
	router.GET("/sitemap.xml", module.Sitemap.GetSitemap)

	v1 := router.Group("/v1")

	v1.Use(module.Middlewares.Authorization())
	{
		v1.POST("/subscribe", module.Users.UserSubscribe)

		// Gamification routes
		v1.GET("/gamification", module.Gaming.GetRules)

		// ACL routes
		v1.GET("/permissions", module.Acl.GetRules)

		// Post routes
		v1.GET("/feed", module.Posts.FeedGet)
		v1.GET("/post", module.Posts.FeedGet)
		v1.GET("/posts/:id", module.Posts.PostsGetOne)
		v1.GET("/post/s/:id", module.Posts.PostsGetOne)

		// // Election routes
		v1.POST("/election/:id", module.Elections.ElectionAddOption)

		// User routes
		v1.POST("/user", module.Users.UserRegisterAction)
		//v1.GET("/user/my/notifications", users.UserNotificationsGet)
		v1.GET("/users/:id", module.Users.UserGetOne)
		v1.GET("/user/activity", module.Users.UserInvolvedFeedGet)
		v1.GET("/user/search", module.Users.UserAutocompleteGet)
		v1.POST("/user/get-token/facebook", module.Users.UserGetTokenFacebook)
		v1.GET("/user/get-token", module.Users.UserGetToken)
		v1.GET("/auth/get-token", module.Users.UserGetJwtToken)

		// Messaging routes
		//v1.GET("/messages", MessagesGet)
		//v1.POST("/messages", MessagePublish)
		//v1.GET("/hashtags", HashtagsGet)

		// Playlist routes
		//v1.GET("/playlist/l/:sections", PlaylistGetList)

		// Categories routes
		v1.GET("/category", module.Categories.CategoriesGet)

		// Parts routes
		v1.GET("/part", module.Parts.GetPartTypes)
		v1.GET("/part/:type/manufacturers", module.Parts.GetPartManufacturers)
		v1.GET("/part/:type/models", module.Parts.GetPartManufacturerModels)

		// Stats routes
		v1.GET("/stats/board", module.Stats.BoardGet)

		authorized := v1.Group("")

		authorized.Use(module.Middlewares.NeedAuthorization())
		{
			// Comment routes
			authorized.POST("/post/comment/:id", module.Comments.CommentAdd)
			authorized.PUT("/post/comment/:id/:index", module.Comments.CommentUpdate)
			authorized.DELETE("/post/comment/:id/:index", module.Comments.CommentDelete)

			// Post routes
			authorized.POST("/post", module.Posts.PostCreate)
			authorized.POST("/post/image", module.Posts.PostUploadAttachment)
			authorized.PUT("/posts/:id", module.Posts.PostUpdate)
			authorized.DELETE("/posts/:id", module.Posts.PostDelete)

			// User routes
			authorized.POST("/user/my/avatar", module.Users.UserUpdateProfileAvatar)
			authorized.GET("/user/my", module.Users.UserGetByToken)
			authorized.PUT("/user/my", module.Users.UserUpdateProfile)
			authorized.POST("/user/my/badge/:id", module.Gaming.BuyBadge)
			authorized.PUT("/category/subscription/:id", module.Users.UserCategorySubscribe)
			authorized.DELETE("/category/subscription/:id", module.Users.UserCategoryUnsubscribe)

			// // Votes routes
			authorized.POST("/vote/comment/:id", module.Votes.VoteComment)
			authorized.POST("/vote/component/:id", module.Votes.VoteComponent)
			authorized.POST("/vote/post/:id", module.Votes.VotePost)
		}
	}

	// Run over the 3000 port
	port := os.Getenv("RUN_OVER")
	if port == "" {
		port = "3000"
	}

	router.Run(":" + port)
}
