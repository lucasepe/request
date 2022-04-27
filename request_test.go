package request

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"
)

func TestRequestURL(t *testing.T) {
	cases := map[string]struct {
		apiUrl string
		path   string
		params []PathParam
		want   string
	}{
		"user-accounts": {
			"https://www.bungie.net",
			"/Platform/User/GetMembershipsForCurrentUser/",
			nil,
			"https://www.bungie.net/Platform/User/GetMembershipsForCurrentUser/",
		},
		"characters": {
			apiUrl: "https://www.bungie.net",
			path:   "/Platform/Destiny2/{membershipType}/Profile/{destinyMembershipId}/{?components}",
			params: []PathParam{
				{"membershipType", "2"},
				{"destinyMembershipId", "1234567890"},
				{"components", "200,205"},
			},
			want: "https://www.bungie.net/Platform/Destiny2/2/Profile/1234567890/?components=200,205",
		},
	}

	tr := roundTripperFunc(func(r *http.Request) (*http.Response, error) {
		// Assert on request attributes
		// Return a response or error you want
		return &http.Response{
			Status:     "OK",
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader(r.URL.String())),
		}, nil
	})

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			var got string
			err := Get(tc.apiUrl).Path(tc.path, tc.params...).
				Transport(tr).
				IntoString(&got).
				Do(context.TODO())
			if err != nil {
				t.Fatal(err)
			}
			if got != tc.want {
				t.Fatalf("got %q; want %q", got, tc.want)
			}
		})
	}
}

type roundTripperFunc func(*http.Request) (*http.Response, error)

func (fn roundTripperFunc) RoundTrip(r *http.Request) (*http.Response, error) {
	return fn(r)
}
