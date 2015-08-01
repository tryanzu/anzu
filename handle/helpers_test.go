package handle 

import "testing"

func TestSlugAscii(t *testing.T) {
	var tests = []struct{ in, out string }{
		{"PC '¿Que opinan spartanos? 2.0'", "pc-que-opinan-spartanos-2-0"},
		{"Clan de destiny ayudamos a todos :)", "clan-de-destiny-ayudamos-a-todos"},
		{"AMD FX-6300 ó AMD Athlon x4 750k?", "amd-fx-6300-o-amd-athlon-x4-750k"},
	}

	for _, test := range tests {
		if out := str_slug(test.in); out != test.out {
			t.Errorf("%q: %q != %q", test.in, out, test.out)
		}
	}
}