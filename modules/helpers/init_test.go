package helpers

import (
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

func TestEmailValidation(t *testing.T) {

	var tests = []struct {
		in  string
		out bool
	}{
		{"fernandez14@outlook.com", true},
		{"ftrl@io", false},
		{"ftetetggdhs", false},
	}

	Convey("Validate emails properly", t, func() {

		for _, test := range tests {

			Convey(test.in+" should be ", func() {

				So(IsEmail(test.in), ShouldEqual, test.out)
			})
		}
	})
}
