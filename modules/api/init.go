package api

import (
	"context"
	"os/signal"
	"time"

	"github.com/facebookgo/inject"
	"github.com/gin-gonic/contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/olebedev/config"
	handle "github.com/tryanzu/core/board/legacy"
	"github.com/tryanzu/core/board/realtime"
	chttp "github.com/tryanzu/core/core/http"
	"github.com/tryanzu/core/modules/api/controller"
	"github.com/tryanzu/core/modules/api/controller/oauth"
	"github.com/tryanzu/core/modules/api/controller/posts"
	"github.com/tryanzu/core/modules/api/controller/users"

	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
)

type Module struct {
	Dependencies ModuleDI
	Posts        handle.PostAPI
	Oauth        oauth.API
	Users        handle.UserAPI
	Middlewares  handle.MiddlewareAPI
	Acl          handle.AclAPI
	Gaming       handle.GamingAPI
	PostsFactory posts.API
	UsersFactory users.API
}

type ModuleDI struct {
	Config *config.Config `inject:""`
}

func (module *Module) Run(bindTo string) {
	debug := true
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

	store := sessions.NewCookieStore([]byte(secret))
	templates, err := module.Dependencies.Config.String("application.templates")
	if err != nil {
		panic(err)
	}

	router := gin.Default()
	router.SetFuncMap(template.FuncMap{
		"config": func(key string) string {
			return module.Dependencies.Config.UString(key, fmt.Sprintf("[Fatal error] Cannot get config with key %s", key))
		},
	})
	router.LoadHTMLGlob(templates)

	// Middlewares setup
	router.Use(sessions.Sessions("session", store))
	router.Use(module.Middlewares.ErrorTracking(debug))
	router.Use(module.Middlewares.CORS())
	router.Use(module.Middlewares.MongoRefresher())
	router.Use(chttp.SiteMiddleware())

	// Production only middlewares
	if debug == false {
		router.Use(module.Middlewares.TrustIP())
		router.Use(chttp.MaxAllowed(5))
	}

	/**
	 * Routes section.
	 * - All route definitions will go below this point.
	 */
	router.Static("/assets", "./static/frontend/public")
	router.Static("/js", "./static/frontend/public/js")
	router.Static("/css", "./static/frontend/public/css")
	router.Static("/images", "./static/frontend/public/images")
	router.Static("/app", "./static/frontend/public/app")
	router.Static("/dist", "./static/frontend/public/dist")

	router.GET("/", controller.HomePage)
	router.GET("/publicar", chttp.TitleMiddleware("Nueva publicación"), controller.HomePage)
	router.GET("/c/:slug", chttp.TitleMiddleware("Categoria"), controller.HomePage)
	router.GET("/chat", chttp.TitleMiddleware("Chat"), controller.HomePage)
	router.GET("/p/:slug/:id", controller.PostPage)
	router.GET("/u/:username/:id", controller.UserPage)
	router.GET("/validate/:code", module.Users.UserValidateEmail)

	v1 := router.Group("/v1")
	v1.Use(module.Middlewares.Authorization())

	// Authentication routes
	v1.GET("/oauth/:provider", module.Oauth.GetAuthRedirect)
	v1.GET("/oauth/:provider/callback", module.Oauth.CompleteAuth)

	// Gamification routes
	v1.GET("/gamification", module.Gaming.GetRules)

	// ACL routes
	v1.GET("/permissions", module.Acl.GetRules)

	// Post routes
	v1.GET("/feed", module.Posts.FeedGet)
	v1.GET("/posts/:id", module.PostsFactory.Get)
	v1.GET("/comments/:post_id", controller.Comments)

	// User routes
	v1.POST("/user", module.Users.UserRegisterAction)
	v1.GET("/users/:id", module.Users.UserGetOne)
	v1.GET("/users/:id/:kind", module.Users.UserGetActivity)
	v1.GET("/user/search", module.Users.UserAutocompleteGet)
	v1.POST("/auth/get-token", module.Users.UserGetJwtToken)
	v1.GET("/auth/lost-password", module.UsersFactory.RequestPasswordRecovery)
	v1.GET("/auth/recovery-token/:token", module.UsersFactory.ValidatePasswordRecovery)
	v1.PUT("/auth/recovery-token/:token", module.UsersFactory.UpdatePasswordFromToken)

	// Categories routes
	v1.GET("/category", controller.Categories)

	authorized := v1.Group("")
	authorized.Use(module.Middlewares.NeedAuthorization())

	authorized.PUT("/config", chttp.UserMiddleware(), chttp.Can("board-config"), controller.UpdateConfig)
	authorized.GET("/notifications", chttp.UserMiddleware(), controller.Notifications)

	// Auth routes
	authorized.GET("/auth/resend-confirmation", module.UsersFactory.ResendConfirmation)

	// Comment routes
	authorized.POST("/comments/:id", chttp.UserMiddleware(), chttp.Can("comment"), controller.NewComment)
	authorized.PUT("/comments/:id", chttp.UserMiddleware(), chttp.Can("comment"), controller.UpdateComment)
	authorized.DELETE("/comments/:id", chttp.UserMiddleware(), chttp.Can("comment"), controller.DeleteComment)

	// Flag routes
	authorized.POST("/flags", chttp.UserMiddleware(), controller.NewFlag)
	authorized.GET("/flags/:related/:id", chttp.UserMiddleware(), controller.Flag)

	// Post routes
	authorized.POST("/post", module.PostsFactory.Create)
	authorized.POST("/post/image", module.Posts.PostUploadAttachment)
	authorized.PUT("/posts/:id", module.PostsFactory.Update)
	authorized.DELETE("/posts/:id", module.Posts.PostDelete)

	// User routes
	authorized.GET("/users", chttp.UserMiddleware(), chttp.Can("users:admin"), controller.Users)
	authorized.POST("/user/my/avatar", module.Users.UserUpdateProfileAvatar)
	authorized.GET("/user/my", module.Users.UserGetByToken)
	authorized.PUT("/user/my", module.Users.UserUpdateProfile)
	authorized.PATCH("/me/:field", module.UsersFactory.Patch)
	authorized.GET("/reasons/ban", chttp.UserMiddleware(), chttp.Can("users:admin"), controller.BanReasons)
	authorized.GET("/reasons/flag", chttp.UserMiddleware(), controller.FlagReasons)
	authorized.POST("/ban", chttp.UserMiddleware(), chttp.Can("users:admin"), controller.Ban)

	// Votes routes
	authorized.POST("/react/:type/:id", chttp.UserMiddleware(), controller.UpsertReaction)

	h := http.NewServeMux()
	h.HandleFunc("/glue/", realtime.ServeHTTP())
	h.HandleFunc("/", router.ServeHTTP)

	// Start the http server as an isolated goroutine.
	srv := &http.Server{
		Addr:    bindTo,
		Handler: h,
	}
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Listen: %s\n", err)
		}
	}()
	// Wait for interrupt signal to gracefully shutdown the server with
	// a timeout of 5 seconds.
	quit := make(chan os.Signal)
	signal.Notify(quit, os.Interrupt)
	<-quit
	log.Println("Shutdown Server ...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server Shutdown:", err)
	}
	log.Println("Server exiting")
}

func (module *Module) Populate(g inject.Graph) {
	err := g.Provide(
		&inject.Object{Value: &module.Dependencies},
		&inject.Object{Value: &module.Posts},
		&inject.Object{Value: &module.PostsFactory},
		&inject.Object{Value: &module.UsersFactory},
		&inject.Object{Value: &module.Users},
		&inject.Object{Value: &module.Middlewares},
		&inject.Object{Value: &module.Acl},
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
