package api

import (
	"github.com/desertbit/glue"
	"github.com/facebookgo/inject"
	chttp "github.com/fernandez14/spartangeek-blacker/core/http"
	"github.com/fernandez14/spartangeek-blacker/handle"
	"github.com/fernandez14/spartangeek-blacker/modules/api/controller"
	"github.com/fernandez14/spartangeek-blacker/modules/api/controller/comments"
	"github.com/fernandez14/spartangeek-blacker/modules/api/controller/oauth"
	"github.com/fernandez14/spartangeek-blacker/modules/api/controller/posts"
	"github.com/fernandez14/spartangeek-blacker/modules/api/controller/users"
	"github.com/fernandez14/spartangeek-blacker/modules/api/controller/votes"
	"github.com/gin-gonic/contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/olebedev/config"

	"fmt"
	"log"
	"net/http"
	"os"
)

type Module struct {
	Dependencies    ModuleDI
	Posts           handle.PostAPI
	Votes           handle.VoteAPI
	VotesFactory    votes.API
	Oauth           oauth.API
	Users           handle.UserAPI
	Categories      handle.CategoryAPI
	CommentsFactory comments.API
	Stats           handle.StatAPI
	Middlewares     handle.MiddlewareAPI
	Collector       handle.CollectorAPI
	Sitemap         handle.SitemapAPI
	Acl             handle.AclAPI
	Gaming          handle.GamingAPI
	PostsFactory    posts.API
	UsersFactory    users.API
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
		&inject.Object{Value: &module.UsersFactory},
		&inject.Object{Value: &module.Votes},
		&inject.Object{Value: &module.VotesFactory},
		&inject.Object{Value: &module.Users},
		&inject.Object{Value: &module.Categories},
		&inject.Object{Value: &module.CommentsFactory},
		&inject.Object{Value: &module.Stats},
		&inject.Object{Value: &module.Middlewares},
		&inject.Object{Value: &module.Acl},
		&inject.Object{Value: &module.Sitemap},
		&inject.Object{Value: &module.Gaming},
		&inject.Object{Value: &module.Oauth},
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
		gin.SetMode(gin.ReleaseMode)
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

	templates, err := module.Dependencies.Config.String("application.templates")
	if err != nil {
		panic(err)
	}

	router := gin.Default()
	router.LoadHTMLGlob(templates)

	// Middlewares setup
	router.Use(sessions.Sessions("session", store))
	router.Use(module.Middlewares.ErrorTracking(debug))
	router.Use(module.Middlewares.CORS())
	router.Use(module.Middlewares.MongoRefresher())
	router.Use(module.Middlewares.StatsdTiming())
	router.Use(module.Middlewares.TrustIP())
	router.Use(chttp.SiteMiddleware())

	/**
	 * Routes section.
	 * - All route definitions will go below this point.
	 */
	router.Static("/assets", "./static/frontend/public")
	router.Static("/js", "./static/frontend/public/js")
	router.Static("/css", "./static/frontend/public/css")
	router.Static("/images", "./static/frontend/public/images")
	router.Static("/app", "./static/frontend/public/app")

	router.GET("/", controller.HomePage)
	router.GET("/chat", chttp.TitleMiddleware("Chat oficial"), controller.HomePage)
	router.GET("/reglamento", chttp.TitleMiddleware("Reglamento y código de conducta"), controller.HomePage)
	router.GET("/about", chttp.TitleMiddleware("Acerca de"), controller.HomePage)
	router.GET("/terminos-y-condiciones", chttp.TitleMiddleware("Terminos y condiciones"), controller.HomePage)
	router.GET("/p/:slug/:id", controller.PostPage)
	router.GET("/u/:username/:id", controller.UserPage)
	router.GET("/sitemap.xml", module.Sitemap.GetSitemap)

	v1 := router.Group("/v1")
	v1.Use(module.Middlewares.Authorization())
	{
		// Authentication routes
		v1.GET("/oauth/:provider", module.Oauth.GetAuthRedirect)
		v1.GET("/oauth/:provider/callback", module.Oauth.CompleteAuth)

		v1.POST("/subscribe", module.Users.UserSubscribe)

		// Gamification routes
		v1.GET("/gamification", module.Gaming.GetRules)
		v1.GET("/stats/ranking", module.Gaming.GetRanking)

		// ACL routes
		v1.GET("/permissions", module.Acl.GetRules)

		// Post routes
		v1.GET("/feed", module.Posts.FeedGet)
		v1.GET("/post", module.Posts.FeedGet)
		v1.GET("/posts/:id", module.PostsFactory.Get)
		v1.GET("/posts/:id/comments", module.PostsFactory.GetPostComments)
		v1.GET("/posts/:id/light", module.Posts.GetLightweight)
		v1.GET("/comments/:post_id", controller.Comments)

		// Search routes
		v1.GET("/search/posts", module.PostsFactory.Search)

		// User routes
		v1.POST("/user", module.Users.UserRegisterAction)
		v1.GET("/users/:id", module.Users.UserGetOne)
		v1.GET("/users/:id/:kind", module.Users.UserGetActivity)
		v1.GET("/user/search", module.Users.UserAutocompleteGet)
		v1.POST("/auth/get-token", module.Users.UserGetJwtToken)
		v1.GET("/auth/lost-password", module.UsersFactory.RequestPasswordRecovery)
		v1.GET("/auth/recovery-token/:token", module.UsersFactory.ValidatePasswordRecovery)
		v1.PUT("/auth/recovery-token/:token", module.UsersFactory.UpdatePasswordFromToken)
		v1.GET("/user/confirm/:code", module.Users.UserValidateEmail)

		// Categories routes
		v1.GET("/category", module.Categories.CategoriesGet)

		// Stats routes
		v1.GET("/stats/board", module.Stats.BoardGet)

		authorized := v1.Group("")
		authorized.Use(module.Middlewares.NeedAuthorization())
		{
			authorized.GET("/notifications", chttp.UserMiddleware(), controller.Notifications)
			authorized.POST("/build", module.PostsFactory.Create)

			// Auth routes
			authorized.GET("/auth/resend-confirmation", module.UsersFactory.ResendConfirmation)

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

			// User routes
			authorized.POST("/user/my/avatar", module.Users.UserUpdateProfileAvatar)
			authorized.GET("/user/my", module.Users.UserGetByToken)
			authorized.PUT("/user/my", module.Users.UserUpdateProfile)
			authorized.PATCH("/me/:field", module.UsersFactory.Patch)
			authorized.PUT("/category/subscription/:id", module.Users.UserCategorySubscribe)
			authorized.DELETE("/category/subscription/:id", module.Users.UserCategoryUnsubscribe)

			// Gamification routes
			authorized.POST("/badges/buy/:id", module.Gaming.BuyBadge)

			// Votes routes
			authorized.POST("/vote/comment/:id", module.VotesFactory.Comment)
			authorized.POST("/vote/component/:id", module.Votes.VoteComponent)
			authorized.POST("/vote/post/:id", module.Votes.VotePost)
		}
	}

	// Run over the 3000 port
	port := os.Getenv("RUN_OVER")
	if port == "" {
		port = "3000"
	}

	glues := glue.NewServer(glue.Options{HTTPSocketType: glue.HTTPSocketTypeNone})
	defer glues.Release()
	glues.OnNewSocket(onNewSocket)

	h := http.NewServeMux()
	h.HandleFunc("/glue/", glues.ServeHTTP)
	h.HandleFunc("/", router.ServeHTTP)

	err = http.ListenAndServe(":"+port, h)
	log.Fatal(err)
}

func onNewSocket(s *glue.Socket) {
	// Set a function which is triggered as soon as the socket is closed.
	s.OnClose(func() {
		log.Printf("socket closed with remote address: %s", s.RemoteAddr())
	})

	// Set a function which is triggered during each received message.
	s.OnRead(func(data string) {
		// Echo the received data back to the client.
		s.Write(data)
	})

	// Send a welcome string to the client.
	s.Write("Hello Client")
}
