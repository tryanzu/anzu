package handle 

import (
	"testing"
	. "github.com/smartystreets/goconvey/convey"
)

func TestSlugAscii(t *testing.T) {
	
	var tests = []struct{ in, out string }{
		{"PC '¿Que opinan spartanos? 2.0'", "pc-que-opinan-spartanos-2-0"},
		{"Clan de destiny ayudamos a todos :)", "clan-de-destiny-ayudamos-a-todos"},
		{"AMD FX-6300 ó AMD Athlon x4 750k?", "amd-fx-6300-o-amd-athlon-x4-750k"},
	}
	
	Convey("Make sure the slug generator works properly with weird characters", t, func() {
		
		for _, test := range tests {
			
			Convey(test.in + " should be " + test.out, func() {
				
				So(str_slug(test.in), ShouldEqual, test.out)
			})
		}
	})
}