package user

import (
	"time"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/tryanzu/core/core/config"
	"gopkg.in/mgo.v2/bson"
)

type userToken struct {
	Address string   `json:"address"`
	UserID  string   `json:"user_id"`
	Scopes  []string `json:"scope"`
	jwt.StandardClaims
}

func CanBeTrusted(user User) bool {
	return user.Warnings < 6
}

func IsBanned(d deps, id bson.ObjectId) bool {
	ledis := d.LedisDB()
	k := []byte("ban:")
	k = append(k, []byte(id)...)
	n, err := ledis.Exists(k)
	if err != nil {
		panic(err)
	}
	return n == 1
}

func genToken(address string, id bson.ObjectId, roles []UserRole, expiration int) string {
	scope := make([]string, len(roles))
	for k, role := range roles {
		scope[k] = role.Name
	}
	if expiration <= 0 {
		expiration = 24
	}
	claims := userToken{
		address,
		id.Hex(),
		scope,
		jwt.StandardClaims{
			ExpiresAt: time.Now().Add(time.Hour * time.Duration(expiration)).Unix(),
			Issuer:    "anzu",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	c := config.C.Copy()

	// Use the secret inside the configuration to encrypt it
	tkn, err := token.SignedString([]byte(c.Security.Secret))
	if err != nil {
		panic(err)
	}

	return tkn
}
