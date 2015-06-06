package handle

import (	
	"github.com/getsentry/raven-go"
	"github.com/gin-gonic/gin"
	"os"
	"fmt"
	"errors"
)

type MiddlewareAPI struct {
	ErrorService *raven.Client `inject:""`
}

func (di *MiddlewareAPI) CORS() gin.HandlerFunc {
	return func(c *gin.Context) {

		c.Writer.Header().Set("Content-Type", "application/json")
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET,POST,OPTIONS,PUT")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Auth-Token,Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")

		if c.Request.Method == "OPTIONS" {

			c.AbortWithStatus(200)
			return
		}
		c.Next()
	}
}

func (di *MiddlewareAPI) ErrorTracking() gin.HandlerFunc {
	return func(c *gin.Context) {

		envfile := os.Getenv("ENV_FILE")

		if envfile == "" {

			envfile = "./env.json"
		}

		tags := map[string]string {
			"config_file": envfile,
		}

		defer func() {

			var packet *raven.Packet

			switch rval := recover().(type) {
			case nil:
				return
			case error:
				packet = raven.NewPacket(rval.Error(), raven.NewException(rval, raven.NewStacktrace(2, 3, nil)))
			default:
				rvalStr := fmt.Sprint(rval)
				packet = raven.NewPacket(rvalStr, raven.NewException(errors.New(rvalStr), raven.NewStacktrace(2, 3, nil)))
			}

			// Grab the error and send it to sentry
			di.ErrorService.Capture(packet, tags)

			// Also abort the request with 500
			c.AbortWithStatus(500)
		}()

		c.Next()
	}
}