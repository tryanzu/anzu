package handle

import (
	"github.com/cactus/go-statsd-client/statsd"
	"github.com/dgrijalva/jwt-go"
	"github.com/fernandez14/spartangeek-blacker/modules/acl"
	"github.com/fernandez14/spartangeek-blacker/mongo"
	"github.com/getsentry/raven-go"
	"github.com/gin-gonic/contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/olebedev/config"
	"github.com/satori/go.uuid"
	"gopkg.in/mgo.v2/bson"
	"gopkg.in/op/go-logging.v1"

	"errors"
	"fmt"
	"log"
	"net"
	"os"
	"strings"
	"syscall"
	"time"
)

type MiddlewareAPI struct {
	ErrorService  *raven.Client   `inject:""`
	ConfigService *config.Config  `inject:""`
	DataService   *mongo.Service  `inject:""`
	StatsService  *statsd.Client  `inject:""`
	Acl           *acl.Module     `inject:""`
	Logger        *logging.Logger `inject:""`
}

func (di *MiddlewareAPI) CORS() gin.HandlerFunc {
	return func(c *gin.Context) {

		origin := c.Request.Header.Get("Origin")

		c.Writer.Header().Set("Content-Type", "application/json")
		c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET,POST,OPTIONS,PUT,DELETE,PATCH")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, X-Requested-With, Auth-Token, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(200)
			return
		}
		c.Next()
	}
}

func (di *MiddlewareAPI) ValidateBsonID(name string) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param(name)
		if bson.IsObjectIdHex(id) == false {
			c.JSON(400, gin.H{"message": "Invalid request, present ID is not valid.", "status": "error"})
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
		name = strings.Replace(name, "github.com/fernandez14/spartangeek-blacker/handle.*", "", -1)
		name = name[0 : len(name)-4]

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

		// Session ID
		var session_id string

		bucket := sessions.Default(c)
		session := bucket.Get("session_id")

		if session == nil {
			uuid := uuid.NewV4()
			session_id = uuid.String()
			bucket.Set("session_id", session_id)
			bucket.Save()
		} else {
			session_id = session.(string)
		}

		// Use same session id anywhere
		c.Set("session_id", session_id)

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

				scope := []string{}
				claims := signed.Claims.(jwt.MapClaims)

				if scopes, exists := claims["scope"]; exists {
					for _, role := range scopes.([]interface{}) {
						scope = append(scope, role.(string))
					}
				}

				// Set the token for further usage
				c.Set("token", jtw_token)
				c.Set("user_id", claims["user_id"].(string))
				c.Set("scope", scope)
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

func (di *MiddlewareAPI) NeedAclAuthorization(permission string) gin.HandlerFunc {
	return func(c *gin.Context) {

		di.Logger.Debug("Need ACL Auth for " + permission)

		if scope, exists := c.Get("scope"); exists {
			roles := scope.([]string)

			di.Logger.Debugf("%v", roles)

			if di.Acl.CheckPermissions(roles, permission) {
				c.Next()
				return
			}
		}

		c.AbortWithStatus(401)
	}
}

func (di *MiddlewareAPI) ErrorTracking(debug bool) gin.HandlerFunc {
	return func(c *gin.Context) {

		if debug == false {
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
				case *net.OpError:
					if rval.Temporary() || rval.Err == syscall.EPIPE || strings.Contains(rval.Error(), "write: broken pipe") {
						return
					}

					packet = raven.NewPacket(rval.Error(), raven.NewException(rval, raven.NewStacktrace(2, 3, nil)))

					// Show the error
					log.Printf("[error][net.OpError] %v\n", rval)
				case error:
					packet = raven.NewPacket(rval.Error(), raven.NewException(rval, raven.NewStacktrace(2, 3, nil)))

					// Show the error
					log.Printf("[error] %v\n", rval)
				default:
					rvalStr := fmt.Sprint(rval)
					packet = raven.NewPacket(rvalStr, raven.NewException(errors.New(rvalStr), raven.NewStacktrace(2, 3, nil)))

					// Show the error
					log.Printf("[error] %v\n", rval)
				}

				// Grab the error and send it to sentry
				di.ErrorService.Capture(packet, tags)

				// Also abort the request with 500
				c.AbortWithStatus(500)
			}()
		} else {

			defer func() {

				switch rval := recover().(type) {
				case nil:
					return
				case error:

					// Show the error
					log.Panic(rval)

				default:

					// Show the error
					log.Panic(rval)
				}

				// Also abort the request with 500
				c.AbortWithStatus(500)
			}()
		}

		c.Next()
	}
}
