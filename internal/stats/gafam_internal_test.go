package stats

import "testing"

func TestMatchGafamCompany(t *testing.T) {
	testCases := []struct {
		domain string
		want   GafamCompany
	}{{
		domain: "www.facebook.com",
		want:   GafamMeta,
	}, {
		domain: "b-graph.facebook.com",
		want:   GafamMeta,
	}, {
		domain: "edge-mqtt.facebook.com",
		want:   GafamMeta,
	}, {
		domain: "static.xx.fbcdn.net",
		want:   GafamMeta,
	}, {
		domain: "fb.some-lite-cdn-provider.net",
		want:   GafamMeta,
	}, {
		domain: "lite.instagram.com",
		want:   GafamMeta,
	}, {
		domain: "www.google.com",
		want:   GafamGoogle,
	}, {
		domain: "example.org",
		want:   -1,
	}}

	for _, tc := range testCases {
		t.Run(tc.domain, func(t *testing.T) {
			got := matchGafamCompany(tc.domain)
			if got != tc.want {
				t.Errorf("matchGafamCompany(%q) = %d, want %d", tc.domain, got, tc.want)
			}
		})
	}
}
