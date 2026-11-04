package money

import "testing"

func TestParseAmount(t *testing.T) {
	cases := []struct {
		in      float64
		want    int64
		wantErr bool
	}{
		{100, 10000, false},
		{20, 2000, false},
		{20.01, 0, true},
		{20.5, 0, true},
		{999_999, 99_999_900, false},
		{10, 0, true},
		{0.01, 0, true},
		{19.99, 0, true},
		{1_000_000, 100_000_000, false},
		{0, 0, true},
		{-1, 0, true},
		{20.005, 0, true},
		{1_000_000.01, 0, true},
	}
	for _, c := range cases {
		got, err := ParseAmount(c.in)
		if (err != nil) != c.wantErr || got != c.want {
			t.Errorf("ParseAmount(%v) = %d, %v; want %d, err=%v", c.in, got, err, c.want, c.wantErr)
		}
	}
}
