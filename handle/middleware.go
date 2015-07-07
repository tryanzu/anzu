package handle

import (
	"errors"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"github.com/getsentry/raven-go"
	"github.com/gin-gonic/gin"
	"github.com/olebedev/config"
	"github.com/fernandez14/spartangeek-blacker/mongo"
	"github.com/cactus/go-statsd-client/statsd"
	"os"
	"time"
	"strings"
)

type MiddlewareAPI struct {
	ErrorService  *raven.Client  `inject:""`
	ConfigService *config.Config `inject:""`
	DataService *mongo.Service `inject:""`
	StatsService *statsd.Client `inject:""`
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

func (di *MiddlewareAPI) StatsdTiming() gin.HandlerFunc {

	return func(c *gin.Context) {

		t := time.Now()

		c.Next()

		latency := time.Since(t)
		name := c.HandlerName()
		name  = strings.Replace(name, "github.com/fernandez14/spartangeek-blacker/handle.*", "", -1)
		name  = name[0:len(name)-4]

		// Send the latency information about the handler
		di.StatsService.TimingDuration(name, latency, 1.0)
	}
}

func (di *MiddlewareAPI) MongoRefresher() gin.HandlerFunc {
	return func(c *gin.Context) {

		c.Next()

		// Refresh the session after the request is done (mongo gets tooooo hot after a while)
		go di.DataService.Session.Refresh()
	}
}

func (di *MiddlewareAPI) Authorization() gin.HandlerFunc {
	return func(c *gin.Context) {

		// Check whether the token is present
		token := c.Request.Header.Get("Authorization")

		if token != "" {

			// Check for the JWT inside the header
			if token[0:6] == "Bearer" {

				jtw_token := token[7:len(token)]
				secret, err := di.ConfigService.String("application.secret")

				if err != nil {
					panic(err)
				}

				signed, err := jwt.Parse(jtw_token, func(passed_token *jwt.Token) (interface{}, error) {
					// since we only use the one private key to sign the tokens,
					// we also only use its public counter part to verify
					return []byte(secret), nil
				})

				// Branch out into the possible error from signing
				switch err.(type) {
				case nil:

					if !signed.Valid { // but may still be invalid

						c.JSON(401, gin.H{"status": "error", "message": "Token compromised, will be notified"})

						// Abort the request
						c.AbortWithStatus(401)
						return
					}

				case *jwt.ValidationError: // Something went wrong during validation

					signingError := err.(*jwt.ValidationError)

					switch signingError.Errors {
					case jwt.ValidationErrorExpired:
						c.JSON(401, gin.H{"status": "error", "message": "Token expired, request new one"})

						// Abort the request
						c.AbortWithStatus(401)
						return

					default:
						c.JSON(401, gin.H{"status": "error", "message": "Error parsing token, will be notified"})

						// Abort the request
						c.AbortWithStatus(401)
						return
					}

				default: // Something weird went wrong
					c.JSON(401, gin.H{"status": "error", "message": "Error parsing token, will be notified"})

					// Abort the request
					c.AbortWithStatus(401)
					return
				}

				// Set the token for further usage
				c.Set("token", jtw_token)
				c.Set("user_id", signed.Claims["user_id"].(string))
			}
		}

		c.Next()
	}
}

func (di *MiddlewareAPI) NeedAuthorization() gin.HandlerFunc {
	return func(c *gin.Context) {

		// Check whether the token is present
		_, token_exists := c.Get("token")

		if token_exists == false {

			c.JSON(401, gin.H{"status": "error", "message": "Auth method required"})

			// Abort the request
			c.AbortWithStatus(401)

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

		tags := map[string]string{
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
