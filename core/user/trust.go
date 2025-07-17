package user

import (
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/tryanzu/core/core/config"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type userToken struct {
	Address string   `json:"address"`
	UserID  string   `json:"user_id"`
	Scopes  []string `json:"scope"`
	jwt.RegisteredClaims
}

func CanBeTrusted(user User) bool {
	return user.Warnings < 6
}

func IsBanned(d deps, id primitive.ObjectID) bool {
	ledis := d.LedisDB()
	k := []byte("ban:")
	k = append(k, []byte(id.Hex())...)
	n, err := ledis.Exists(k)
	if err != nil {
		panic(err)
	}
	return n == 1
}

func genToken(address string, id primitive.ObjectID, roles []UserRole, expiration int) string {
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
		jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour * time.Duration(expiration))),
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
