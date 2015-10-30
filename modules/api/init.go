package api

import (
	"fmt"
	"github.com/brandfolder/gin-gorelic"
	"github.com/facebookgo/inject"
	"github.com/fernandez14/spartangeek-blacker/handle"
	"github.com/fernandez14/spartangeek-blacker/modules/api/controller"
	"github.com/gin-gonic/gin"
	"github.com/olebedev/config"
	"os"
)

type Module struct {
	Dependencies ModuleDI
	Posts        handle.PostAPI
	Votes        handle.VoteAPI
	Users        handle.UserAPI
	Categories   handle.CategoryAPI
	Elections    handle.ElectionAPI
	Comments     handle.CommentAPI
	Parts        handle.PartAPI
	Stats        handle.StatAPI
	Middlewares  handle.MiddlewareAPI
	Collector    handle.CollectorAPI
	Sitemap      handle.SitemapAPI
	Acl          handle.AclAPI
	Gaming       handle.GamingAPI
	Store        controller.StoreAPI
	BuildNotes   controller.BuildNotesAPI
	Mail         controller.MailAPI
	PostsFactory controller.PostAPI
}

type ModuleDI struct {
	Config *config.Config `inject:""`
}

func (module *Module) Populate(g inject.Graph) {

	err := g.Provide(
		&inject.Object{Value: &module.Dependencies},
		&inject.Object{Value: &module.Collector},
		&inject.Object{Value: &module.Posts},
		&inject.Object{Value: &module.PostsFactory},
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
		&inject.Object{Value: &module.Store},
		&inject.Object{Value: &module.BuildNotes},
		&inject.Object{Value: &module.Mail},
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

	var debug bool = true

	environment, err := module.Dependencies.Config.String("environment")

	if err != nil {
		panic(err)
	}

	// If development turn debug on
	if environment != "development" {
		debug = false
	}

	// Start gin classic middlewares
	router := gin.Default()

	// Start gorelic
	gorelic.InitNewrelicAgent("3e8e387fb7b29dedb924db3ba88e2790599bd0fb", "Blacker", false)

	// Middlewares setup
	router.Use(gorelic.Handler)
	router.Use(module.Middlewares.ErrorTracking(debug))
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
		v1.GET("/stats/ranking", module.Gaming.GetRanking)

		// ACL routes
		v1.GET("/permissions", module.Acl.GetRules)

		// Post routes
		v1.GET("/feed", module.Posts.FeedGet)
		v1.GET("/post", module.Posts.FeedGet)
		v1.GET("/posts/:id", module.Posts.PostsGetOne)
		v1.GET("/posts/:id/light", module.Posts.GetLightweight)
		v1.GET("/post/s/:id", module.Posts.PostsGetOne)

		// // Election routes
		v1.POST("/election/:id", module.Elections.ElectionAddOption)

		// User routes
		v1.POST("/user", module.Users.UserRegisterAction)
		//v1.GET("/user/my/notifications", users.UserNotificationsGet)
		v1.GET("/users/:id", module.Users.UserGetOne)
		v1.GET("/users/:id/:kind", module.Users.UserGetActivity)
		v1.GET("/user/search", module.Users.UserAutocompleteGet)
		v1.POST("/user/get-token/facebook", module.Users.UserGetTokenFacebook)
		v1.GET("/user/get-token", module.Users.UserGetToken)
		v1.GET("/auth/get-token", module.Users.UserGetJwtToken)
		v1.GET("/user/confirm/:code", module.Users.UserValidateEmail)

		// Categories routes
		v1.GET("/category", module.Categories.CategoriesGet)

		// Parts routes
		v1.GET("/part", module.Parts.GetPartTypes)
		v1.GET("/part/:type/manufacturers", module.Parts.GetPartManufacturers)
		v1.GET("/part/:type/models", module.Parts.GetPartManufacturerModels)

		// Stats routes
		v1.GET("/stats/board", module.Stats.BoardGet)

		// Store routes
		store := v1.Group("/store")

		store.POST("/order", module.Store.PlaceOrder)

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
			authorized.POST("/posts/:id/answer/:comment", module.PostsFactory.MarkCommentAsAnswer)

			// User routes
			authorized.POST("/user/my/avatar", module.Users.UserUpdateProfileAvatar)
			authorized.GET("/user/my", module.Users.UserGetByToken)
			authorized.PUT("/user/my", module.Users.UserUpdateProfile)
			authorized.PUT("/category/subscription/:id", module.Users.UserCategorySubscribe)
			authorized.DELETE("/category/subscription/:id", module.Users.UserCategoryUnsubscribe)

			// Gamification routes
			authorized.POST("/badges/buy/:id", module.Gaming.BuyBadge)

			// // Votes routes
			authorized.POST("/vote/comment/:id", module.Votes.VoteComment)
			authorized.POST("/vote/component/:id", module.Votes.VoteComponent)
			authorized.POST("/vote/post/:id", module.Votes.VotePost)

			// Backoffice routes
			backoffice := authorized.Group("backoffice")

			backoffice.Use(module.Middlewares.NeedAclAuthorization())
			{
				backoffice.GET("/order", module.Store.Orders)
				backoffice.GET("/order/:id", module.Store.One)
				backoffice.POST("/order/:id", module.Store.Answer)
				backoffice.POST("/order/:id/tag", module.Store.Tag)
				backoffice.POST("/order/:id/activity", module.Store.Activity)
				backoffice.POST("/order/:id/stage", module.Store.Stage)

				// Build notes routes
				backoffice.GET("/notes", module.BuildNotes.All)
				backoffice.POST("/notes", module.BuildNotes.Create)
				backoffice.GET("/notes/:id", module.BuildNotes.One)
				backoffice.PUT("/notes/:id", module.BuildNotes.Update)
				backoffice.DELETE("/notes/:id", module.BuildNotes.Delete)
			}
		}

		mail := v1.Group("/mail")
		{
			mail.HEAD("/inbound/:address", func(c *gin.Context) { c.String(200, ":)") })
			mail.POST("/inbound/:address", module.Mail.Inbound)
		}
	}

	// Run over the 3000 port
	port := os.Getenv("RUN_OVER")
	if port == "" {
		port = "3000"
	}

	router.Run(":" + port)
}
