package oauth

import (
	"fmt"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/fernandez14/spartangeek-blacker/modules/security"
	"github.com/fernandez14/spartangeek-blacker/modules/user"
	"github.com/gin-gonic/contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/markbates/goth"
	"github.com/olebedev/config"
	"gopkg.in/mgo.v2/bson"
)

type API struct {
	Users    *user.Module     `inject:""`
	Config   *config.Config   `inject:""`
	Security *security.Module `inject:""`
}

func (a API) GetAuthRedirect(c *gin.Context) {
	siteUrl, err := a.Config.String("application.siteUrl")
	if err != nil {
		panic(err)
	}

	provider, err := goth.GetProvider(c.Param("provider"))
	if err != nil {
		c.JSON(500, gin.H{"status": "error", "message": err.Error()})
		return
	}

	sess, err := provider.BeginAuth("state")
	if err != nil {
		c.JSON(500, gin.H{"status": "error", "message": err.Error()})
		return
	}

	url, err := sess.GetAuthURL()
	if err != nil {
		c.JSON(500, gin.H{"status": "error", "message": err.Error()})
		return
	}

	if !strings.HasPrefix(c.Query("redir"), siteUrl) && !strings.HasPrefix(c.Query("redir"), "https://buldar.com") {
		c.JSON(401, gin.H{"status": "unauthorized."})
		return
	}

	bucket := sessions.Default(c)
	bucket.Set("oauth", sess.Marshal())
	bucket.Set("redir", c.Query("redir"))
	bucket.Save()

	c.Redirect(303, url)
}

func (a API) CompleteAuth(c *gin.Context) {
	provider, err := goth.GetProvider("facebook")
	if err != nil {
		c.JSON(500, gin.H{"status": "provider-error", "message": err.Error()})
		return
	}

	bucket := sessions.Default(c)
	oauth := bucket.Get("oauth")
	if oauth == nil {
		c.JSON(500, gin.H{"status": "error", "message": "Could not get oauth session ref"})
		return
	}
	redir := bucket.Get("redir")
	bucket.Delete("oauth")
	bucket.Delete("redir")
	bucket.Save()

	sess, err := provider.UnmarshalSession(oauth.(string))
	if err != nil {
		c.JSON(500, gin.H{"status": "session-error", "oauth": oauth, "message": err.Error()})
		return
	}

	_, err = sess.Authorize(provider, c.Request.URL.Query())
	if err != nil {
		c.JSON(500, gin.H{"status": "authorize-error", "message": err.Error()})
		return
	}

	usr, err := provider.FetchUser(sess)
	if err != nil {
		c.JSON(500, gin.H{"status": "error", "message": err.Error()})
		return
	}

	var constraints []bson.M
	var id bson.ObjectId

	if len(usr.UserID) > 0 {
		field := usr.Provider + ".id"
		constraints = append(constraints, bson.M{field: usr.UserID, "deleted_at": bson.M{"$exists": false}})
	}

	if len(usr.Email) > 0 {
		constraints = append(constraints, bson.M{"email": usr.Email, "deleted_at": bson.M{"$exists": false}})
	}

	u, err := a.Users.Get(bson.M{"$or": constraints})
	if err != nil {
		u, err = a.Users.OauthSignup(usr.Provider, usr)

		if err != nil {
			c.JSON(400, gin.H{"status": "error", "message": err.Error()})
			return
		}
	}

	// The id for the token would be the same as the facebook user
	id = u.Data().Id
	trusted_user := a.Security.TrustUserIP(c.ClientIP(), u)
	if !trusted_user {
		c.JSON(403, gin.H{"status": "error", "message": "Not trusted."})
		return
	}

	if len(usr.Email) > 0 {
		_ = u.Update(map[string]interface{}{usr.Provider: usr.RawData, "email": usr.Email})
	}

	forward := redir.(string)

	if len(forward) < 6 {
		forward, err = a.Config.String("application.siteUrl")
		if err != nil {
			panic(err)
		}
	}

	// Generate JWT with the information about the user
	token := a.generateUserToken(id, u.Data().Roles, 72)
	url := fmt.Sprintf("%s/?token=%s", forward, token)

	c.Redirect(303, url)
}

type UserToken struct {
	UserID string   `json:"user_id"`
	Scopes []string `json:"scope"`
	jwt.StandardClaims
}

func (a API) generateUserToken(id bson.ObjectId, roles []user.UserRole, expiration int) string {
	scope := []string{}
	for _, role := range roles {
		scope = append(scope, role.Name)
	}

	claims := UserToken{
		id.Hex(),
		scope,
		jwt.StandardClaims{
			ExpiresAt: time.Now().Add(time.Hour * time.Duration(expiration)).Unix(),
			Issuer:    "spartangeek",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Use the secret inside the configuration to encrypt it
	secret, err := a.Config.String("application.secret")
	if err != nil {
		panic(err)
	}

	tokenString, err := token.SignedString([]byte(secret))
	if err != nil {
		panic(err)
	}

	return tokenString
}
