package preprocessor

import (
	"github.com/fernandez14/spartangeek-blacker/handle"
	"github.com/gin-gonic/gin"
	"github.com/olebedev/config"
	"os"
)

type Module struct {
	Dependencies ModuleDI
	Middlewares  handle.MiddlewareAPI
	Posts        controller.PostAPI
}

type ModuleDI struct {
	Config *config.Config `inject:""`
}

func (module *Module) Populate(g inject.Graph) {

	err := g.Provide(
		&inject.Object{Value: &module.Dependencies},
		&inject.Object{Value: &module.Middlewares},
		&inject.Object{Value: &module.Posts},
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

	// Middlewares setup
	router.Use(module.Middlewares.ErrorTracking(debug))
	router.Use(module.Middlewares.MongoRefresher())

	router.GET("/p/:slug/:id", module.Posts.Get)

	// Run over the 3014 port
	port := os.Getenv("RUN_OVER")
	if port == "" {
		port = "3014"
	}

	router.Run(":" + port)
}
