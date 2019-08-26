package helpers

import (
	"crypto/sha256"
	"encoding/hex"
	"math/rand"
	"reflect"
	"regexp"
	"strings"
	"time"
	"unicode"

	"golang.org/x/crypto/bcrypt"
	"golang.org/x/text/unicode/norm"
)

var letters = []rune("0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
var lat = []*unicode.RangeTable{unicode.Letter, unicode.Number}
var nop = []*unicode.RangeTable{unicode.Mark, unicode.Sk, unicode.Lm}
var email_exp = regexp.MustCompile(`^[a-z0-9._%+\-]+@[a-z0-9.\-]+\.[a-z]{2,4}$`)

func InArray(val interface{}, array interface{}) (exists bool, index int) {
	exists = false
	index = -1

	switch reflect.TypeOf(array).Kind() {
	case reflect.Slice:
		s := reflect.ValueOf(array)

		for i := 0; i < s.Len(); i++ {
			if reflect.DeepEqual(val, s.Index(i).Interface()) == true {
				index = i
				exists = true
				return
			}
		}
	}

	return
}

func Truncate(s string, length int) string {
	var numRunes = 0
	for index, _ := range s {
		numRunes++
		if numRunes > length {
			return s[:index]
		}
	}
	return s
}

func StrSlug(s string) string {

	// Trim before counting
	s = strings.Trim(s, " ")

	buf := make([]rune, 0, len(s))
	dash := false
	for _, r := range norm.NFKD.String(s) {
		switch {
		// unicode 'letters' like mandarin characters pass through
		case unicode.IsOneOf(lat, r):
			buf = append(buf, unicode.ToLower(r))
			dash = true
		case unicode.IsOneOf(nop, r):
			// skip
		case dash:
			buf = append(buf, '-')
			dash = false
		}
	}
	if i := len(buf) - 1; i >= 0 && buf[i] == '-' {
		buf = buf[:i]
	}
	return string(buf)
}

func StrSlugRandom(s string) string {

	var letters = []rune("0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

	now := time.Now()
	rand.Seed(now.UnixNano())

	b := make([]rune, 6)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}

	suffix := string(b)

	return StrSlug(s) + suffix
}

func StrRandom(length int) string {
	r := rand.New(rand.NewSource(time.Now().UnixNano() + int64(length)))
	b := make([]rune, length)
	for i := range b {
		b[i] = letters[r.Intn(len(letters))]
	}
	return string(b)
}

func StrCapRandom(length int) string {

	var letters = []rune("0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ")

	now := time.Now()
	rand.Seed(now.UnixNano())

	b := make([]rune, length)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}

	generated := string(b)

	return generated
}

func StrNumRandom(length int) string {

	var letters = []rune("0123456789")

	now := time.Now()
	rand.Seed(now.UnixNano())

	b := make([]rune, length)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}

	generated := string(b)

	return generated
}

func Sha256(s string) string {
	encrypted := []byte(s)
	sha256 := sha256.New()
	sha256.Write(encrypted)
	return hex.EncodeToString(sha256.Sum(nil))
}

func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}

func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

func IsEmail(s string) bool {
	return email_exp.MatchString(s)
}
