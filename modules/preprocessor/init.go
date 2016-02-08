package preprocessor

import (
	"github.com/fernandez14/spartangeek-blacker/handle"
	"github.com/facebookgo/inject"
	"github.com/fernandez14/spartangeek-blacker/modules/preprocessor/controller"
	"github.com/gin-gonic/gin"
	"github.com/olebedev/config"
	"os"
	"fmt"
	"io/ioutil"
)

type Module struct {
	Dependencies ModuleDI
	Middlewares  handle.MiddlewareAPI
	Posts        controller.PostAPI
	Components   controller.ComponentAPI
	General      controller.GeneralAPI
}

type ModuleDI struct {
	Config *config.Config `inject:""`
}

func (module *Module) Populate(g inject.Graph) {

	err := g.Provide(
		&inject.Object{Value: &module.Dependencies},
		&inject.Object{Value: &module.Middlewares},
		&inject.Object{Value: &module.Posts},
		&inject.Object{Value: &module.Components},
		&inject.Object{Value: &module.General},
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

	page, err := module.Dependencies.Config.String("application.page")

	if err != nil {
		panic(err)
	}

	data, err := ioutil.ReadFile(page)
	base := string(data)

	// Share base html page to the controllers
	module.Posts.Page = base
	module.Components.Page = base
	module.General.Page = base

	// Start gin classic middlewares
	router := gin.Default()

	// Middlewares setup
	router.Use(module.Middlewares.ErrorTracking(debug))
	router.Use(module.Middlewares.MongoRefresher())

	router.GET("/p/:slug/:id", module.Posts.Get)
	router.GET("/p/:slug/:id/:comment", module.Posts.Get)
	router.GET("/componentes", module.Components.Landing)
	router.GET("/componentes/:type", module.Components.Get)
	router.GET("/componentes/:type/:slug", module.Components.Get)
	router.GET("/componente/:slug", module.Components.MigrateOld)
	router.GET("/", module.General.Landing)

	// Run over the 3014 port
	port := os.Getenv("RUN_OVER")
	if port == "" {
		port = "3014"
	}

	router.Run(":" + port)
}
