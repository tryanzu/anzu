package api

import (
	"fmt"
	"github.com/brandfolder/gin-gorelic"
	"github.com/facebookgo/inject"
	"github.com/fernandez14/spartangeek-blacker/handle"
	"github.com/fernandez14/spartangeek-blacker/modules/api/controller"
	"github.com/fernandez14/spartangeek-blacker/modules/api/controller/cart"
	"github.com/fernandez14/spartangeek-blacker/modules/api/controller/checkout"
	"github.com/fernandez14/spartangeek-blacker/modules/api/controller/comments"
	"github.com/fernandez14/spartangeek-blacker/modules/api/controller/components"
	"github.com/fernandez14/spartangeek-blacker/modules/api/controller/deals"
	"github.com/fernandez14/spartangeek-blacker/modules/api/controller/massdrop"
	"github.com/fernandez14/spartangeek-blacker/modules/api/controller/posts"
	"github.com/fernandez14/spartangeek-blacker/modules/api/controller/products"
	"github.com/fernandez14/spartangeek-blacker/modules/api/controller/users"
	"github.com/gin-gonic/contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/olebedev/config"
	"os"
)

type Module struct {
	Dependencies      ModuleDI
	Posts             handle.PostAPI
	Votes             handle.VoteAPI
	Users             handle.UserAPI
	Categories        handle.CategoryAPI
	Elections         handle.ElectionAPI
	CommentsFactory   comments.API
	Parts             handle.PartAPI
	Stats             handle.StatAPI
	Middlewares       handle.MiddlewareAPI
	Collector         handle.CollectorAPI
	Sitemap           handle.SitemapAPI
	Acl               handle.AclAPI
	Gaming            handle.GamingAPI
	Store             controller.StoreAPI
	BuildNotes        controller.BuildNotesAPI
	Mail              controller.MailAPI
	PostsFactory      posts.API
	Components        controller.ComponentAPI
	CartFactory       cart.API
	Checkout          checkout.API
	Products          products.API
	Massdrop          massdrop.API
	Deals             deals.API
	Customer          controller.CustomerAPI
	Orders            controller.OrdersAPI
	Owners            controller.OwnersAPI
	Lead              controller.LeadAPI
	ComponentsFactory components.API
	UsersFactory      users.API
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
		&inject.Object{Value: &module.ComponentsFactory},
		&inject.Object{Value: &module.UsersFactory},
		&inject.Object{Value: &module.CartFactory},
		&inject.Object{Value: &module.Checkout},
		&inject.Object{Value: &module.Products},
		&inject.Object{Value: &module.Votes},
		&inject.Object{Value: &module.Users},
		&inject.Object{Value: &module.Categories},
		&inject.Object{Value: &module.Elections},
		&inject.Object{Value: &module.CommentsFactory},
		&inject.Object{Value: &module.Parts},
		&inject.Object{Value: &module.Stats},
		&inject.Object{Value: &module.Middlewares},
		&inject.Object{Value: &module.Acl},
		&inject.Object{Value: &module.Sitemap},
		&inject.Object{Value: &module.Gaming},
		&inject.Object{Value: &module.Store},
		&inject.Object{Value: &module.BuildNotes},
		&inject.Object{Value: &module.Mail},
		&inject.Object{Value: &module.Components},
		&inject.Object{Value: &module.Massdrop},
		&inject.Object{Value: &module.Customer},
		&inject.Object{Value: &module.Orders},
		&inject.Object{Value: &module.Owners},
		&inject.Object{Value: &module.Deals},
		&inject.Object{Value: &module.Lead},
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

	// Session storage
	secret, err := module.Dependencies.Config.String("application.secret")

	if err != nil {
		panic(err)
	}

	redis_server, err := module.Dependencies.Config.String("cache.redis")

	if err != nil {
		panic(err)
	}

	store, err := sessions.NewRedisStore(10, "tcp", redis_server, "", []byte(secret))

	if err != nil {
		panic(err)
	}

	// Start gin classic middlewares
	router := gin.Default()

	// Start gorelic
	gorelic.InitNewrelicAgent("3e8e387fb7b29dedb924db3ba88e2790599bd0fb", "Blacker", false)

	// Middlewares setup
	router.Use(sessions.Sessions("session", store))
	router.Use(gorelic.Handler)
	router.Use(module.Middlewares.ErrorTracking(debug))
	router.Use(module.Middlewares.CORS())
	router.Use(module.Middlewares.MongoRefresher())
	router.Use(module.Middlewares.StatsdTiming())

	// Sitemap generator
	router.GET("/sitemap.xml", module.Sitemap.GetSitemap)
	router.GET("/sitemap_components.xml", module.Sitemap.GetComponentsSitemap)

	v1 := router.Group("/v1")

	v1.Use(module.Middlewares.Authorization())
	{
		v1.GET("/panic-now", func(c *gin.Context) {

			panic("panic is just an illusion")
		})

		v1.GET("/whoami", func(c *gin.Context) {

			c.String(200, c.ClientIP())
		})

		v1.POST("/subscribe", module.Users.UserSubscribe)
		v1.POST("/leads", module.Lead.Post)

		// Gamification routes
		v1.GET("/gamification", module.Gaming.GetRules)
		v1.GET("/stats/ranking", module.Gaming.GetRanking)

		// ACL routes
		v1.GET("/permissions", module.Acl.GetRules)

		// Post routes
		v1.GET("/feed", module.Posts.FeedGet)
		v1.GET("/post", module.Posts.FeedGet)
		v1.GET("/posts/:id", module.PostsFactory.Get)
		v1.GET("/postss/:id", module.Posts.PostsGetOne)
		v1.GET("/posts/:id/comments", module.PostsFactory.GetPostComments)
		v1.GET("/posts/:id/light", module.Posts.GetLightweight)
		v1.GET("/post/s/:id", module.Posts.PostsGetOne)

		// Search routes
		v1.GET("/search/posts", module.PostsFactory.Search)
		v1.GET("/search/products", module.Products.Search)
		v1.GET("/search/components", module.ComponentsFactory.Search)

		// Massdrop routes
		v1.GET("/massdrop", module.Massdrop.Get)

		// // Election routes
		v1.POST("/election/:id", module.Elections.ElectionAddOption)

		// User routes
		v1.POST("/user", module.Users.UserRegisterAction)
		//v1.GET("/user/my/notifications", users.UserNotificationsGet)
		v1.GET("/users/:id", module.Users.UserGetOne)
		v1.GET("/users/:id/:kind", module.Users.UserGetActivity)
		v1.GET("/user/search", module.Users.UserAutocompleteGet)
		v1.POST("/user/get-token/facebook", module.Users.UserGetTokenFacebook)
		v1.GET("/auth/get-token", module.Users.UserGetJwtToken)
		v1.GET("/auth/lost-password", module.UsersFactory.RequestPasswordRecovery)
		v1.GET("/auth/recovery-token/:token", module.UsersFactory.ValidatePasswordRecovery)
		v1.PUT("/auth/recovery-token/:token", module.UsersFactory.UpdatePasswordFromToken)
		v1.GET("/user/confirm/:code", module.Users.UserValidateEmail)

		// Categories routes
		v1.GET("/category", module.Categories.CategoriesGet)

		// Parts routes
		v1.GET("/part", module.Parts.GetPartTypes)
		v1.GET("/part/:type/manufacturers", module.Parts.GetPartManufacturers)
		v1.GET("/part/:type/models", module.Parts.GetPartManufacturerModels)
		v1.GET("/component/:id", module.Components.Get)
		v1.GET("/component/:id/posts", module.Components.GetPosts)

		// Stats routes
		v1.GET("/stats/board", module.Stats.BoardGet)

		// Store routes
		store := v1.Group("/store")
		{
			store.POST("/order", module.Store.PlaceOrder)

			// Cart routes
			store.GET("/cart", module.CartFactory.Get)
			store.POST("/cart", module.CartFactory.Add)
			store.DELETE("/cart/:id", module.CartFactory.Delete)

			// Products routes
			store.GET("/product/:id", module.Products.Get)

			// Store routes with auth
			astore := store.Group("")

			astore.Use(module.Middlewares.NeedAuthorization())
			{
				astore.POST("/checkout", module.Checkout.Place)
				astore.POST("/checkout/massdrop", module.Checkout.Massdrop)

				// Massdrop insterested
				store.PUT("/product/:id/massdrop", module.Products.Massdrop)

				// Customer routes
				astore.GET("/customer", module.Customer.Get)
				astore.POST("/customer/address", module.Customer.CreateAddress)
				astore.DELETE("/customer/address/:id", module.Customer.DeleteAddress)
				astore.PUT("/customer/address/:id", module.Customer.UpdateAddress)
			}
		}

		authorized := v1.Group("")

		authorized.Use(module.Middlewares.NeedAuthorization())
		{
			// Auth routes
			v1.GET("/auth/logout", module.Users.UserLogout)

			// Comment routes
			authorized.POST("/post/comment/:id", module.CommentsFactory.Add)
			authorized.PUT("/post/comment/:id", module.CommentsFactory.Update)
			authorized.DELETE("/post/comment/:id", module.CommentsFactory.Delete)

			// Post routes
			authorized.POST("/post", module.PostsFactory.Create)
			authorized.POST("/post/image", module.Posts.PostUploadAttachment)
			authorized.PUT("/posts/:id", module.PostsFactory.Update)
			authorized.DELETE("/posts/:id", module.Posts.PostDelete)
			authorized.POST("/posts/:id/answer/:comment", module.PostsFactory.MarkCommentAsAnswer)
			authorized.POST("/posts/:id/relate/:related_id", module.PostsFactory.Relate)

			// User routes
			authorized.POST("/user/my/avatar", module.Users.UserUpdateProfileAvatar)
			authorized.GET("/user/my", module.Users.UserGetByToken)
			authorized.PUT("/user/my", module.Users.UserUpdateProfile)
			authorized.PUT("/category/subscription/:id", module.Users.UserCategorySubscribe)
			authorized.DELETE("/category/subscription/:id", module.Users.UserCategoryUnsubscribe)
			authorized.POST("/user/own/:kind/:id", module.Owners.Post)
			authorized.DELETE("/user/own/:kind/:id", module.Owners.Delete)

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
				store := backoffice.Group("/store")
				{
					store.GET("/order", module.Orders.Get)
					store.GET("/order/:id", module.Orders.GetOne)
					store.POST("/order/:id/send-confirmation", module.Orders.SendOrderConfirmation)
					store.PUT("/order/:id/status", module.Orders.ChangeStatus)
				}

				backoffice.POST("/deals/invoice", module.Deals.GenerateInvoice)

				backoffice.GET("/order", module.Store.Orders)
				backoffice.GET("/order/:id", module.Store.One)
				backoffice.DELETE("/order/:id", module.Store.Ignore)
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

				// Components routes
				backoffice.PUT("/component/:slug/price", module.Components.UpdatePrice)
				backoffice.DELETE("/component/:slug/price", module.Components.DeletePrice)
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
