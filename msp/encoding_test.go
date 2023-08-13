package msp

import (
	"github.com/dottics/dutil"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"
)

func TestDecode(t *testing.T) {
	type payload struct {
		Name string `json:"name"`
	}
	type E struct {
		body string
		data payload
		e    dutil.Err
	}
	tt := []struct {
		name string
		res  *http.Response
		v    interface{}
		E    E
	}{
		{
			name: "no interface",
			res: &http.Response{
				Body: ioutil.NopCloser(strings.NewReader(`{"name":"james"}`)),
			},
			v: nil,
			E: E{
				data: payload{},
				body: `{"name":"james"}`,
			},
		},
		{
			name: "unmarshal error",
			res: &http.Response{
				Body: ioutil.NopCloser(strings.NewReader(`{"name":1}`)),
			},
			v: &payload{},
			E: E{
				body: "",
				data: payload{},
				e: dutil.Err{
					Status: 500,
					Errors: map[string][]string{
						"unmarshal": {"json: cannot unmarshal number into Go struct field payload.name of type string"},
					},
				},
			},
		},
		{
			name: "successful unmarshal",
			res: &http.Response{
				Body: ioutil.NopCloser(strings.NewReader(`{"name":"james"}`)),
			},
			v: &payload{},
			E: E{
				body: `{"name":"james"}`,
				data: payload{Name: "james"},
				e:    dutil.Err{},
			},
		},
	}

	//s := NewService(Config{Name: "micro"})

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			xb, e := Decode(tc.res, tc.v)
			if tc.E.e.Status != 0 {
				if tc.E.e.Error() != e.Error() {
					t.Errorf("expected '%v' got '%v'", tc.E.e.Error(), e.Error())
				}
			}
			if string(xb) != tc.E.body {
				t.Errorf("expected '%v' got '%v'", tc.E.body, string(xb))
			}
			if tc.E.data != (payload{}) {
				if tc.E.data.Name != "james" {
					t.Errorf("expected '%v' got '%v'", tc.E.data.Name, "james")
				}
			}
		})
	}
}
