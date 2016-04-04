package helpers

import (
	"crypto/sha256"
	"encoding/hex"
	"golang.org/x/text/unicode/norm"
	"math/rand"
	"reflect"
	"regexp"
	"strings"
	"unicode"
	"time"
)

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

	var letters = []rune("0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

	b := make([]rune, length)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}

	generated := string(b)

	return generated
}

func StrCapRandom(length int) string {

	var letters = []rune("0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ")

	b := make([]rune, length)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}

	generated := string(b)

	return generated
}


func Sha256(s string) string {

	password_encrypted := []byte(s)
	sha256 := sha256.New()
	sha256.Write(password_encrypted)
	md := sha256.Sum(nil)

	return hex.EncodeToString(md)
}

func IsEmail(s string) bool {

	return email_exp.MatchString(s)
}
