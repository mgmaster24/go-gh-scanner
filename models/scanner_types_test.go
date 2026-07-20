package models

import "testing"

func TestStripRangePrefix(t *testing.T) {
	cases := []struct {
		input string
		want  string
	}{
		{"^2.1.0", "2.1.0"},
		{"~1.0.0", "1.0.0"},
		{">=3.0.0", "3.0.0"},
		{"<=1.5.0", "1.5.0"},
		{">1.0.0", "1.0.0"},
		{"<2.0.0", "2.0.0"},
		{"=1.2.3", "1.2.3"},
		{"workspace:^", ""},
		{"workspace:1.0.0", "1.0.0"},
		{"1.0.0", "1.0.0"},
		{"", ""},
	}

	for _, c := range cases {
		got := stripRangePrefix(c.input)
		if got != c.want {
			t.Errorf("stripRangePrefix(%q) = %q, want %q", c.input, got, c.want)
		}
	}
}

func TestGetDepVersion(t *testing.T) {
	cases := []struct {
		name    string
		frag    string
		dep     string
		want    string
		wantOk  bool
	}{
		{
			name:   "caret range mid-object",
			frag:   `"@m2s2/ng-lib": "^2.1.0", "@other/pkg": "1.0.0"`,
			dep:    "@m2s2/ng-lib",
			want:   "2.1.0",
			wantOk: true,
		},
		{
			name:   "tilde range last key",
			frag:   `"@m2s2/react-lib": "~3.0.0"}`,
			dep:    "@m2s2/react-lib",
			want:   "3.0.0",
			wantOk: true,
		},
		{
			name:   "exact version",
			frag:   `"@m2s2/vue-lib": "1.5.2", "something": "else"`,
			dep:    "@m2s2/vue-lib",
			want:   "1.5.2",
			wantOk: true,
		},
		{
			name:   "workspace protocol",
			frag:   `"@m2s2/ng-lib": "workspace:^", "b": "1"`,
			dep:    "@m2s2/ng-lib",
			want:   "",
			wantOk: false,
		},
		{
			name:   "dep not present",
			frag:   `"some-other-lib": "1.0.0"`,
			dep:    "@m2s2/ng-lib",
			want:   "",
			wantOk: false,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got, ok := FragmentStr(c.frag).GetDepVersion(c.dep)
			if ok != c.wantOk {
				t.Errorf("ok = %v, want %v", ok, c.wantOk)
			}
			if got != c.want {
				t.Errorf("version = %q, want %q", got, c.want)
			}
		})
	}
}
