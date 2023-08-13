package msp

import (
	"github.com/dottics/dutil"
	"github.com/johannesscr/micro/microtest"
	"net/http"
	"net/url"
	"os"
	"strings"
	"testing"
)

func TestNewService(t *testing.T) {
	type E struct {
		scheme string
		host   string
		token  string
	}
	tt := []struct {
		name            string
		budgetSchemeEnv string
		budgetHostEnv   string
		token           string
		E               E
	}{
		{
			name: "default",
			E: E{
				scheme: "",
				host:   "",
				token:  "",
			},
		},
		{
			name:            "env vars",
			budgetSchemeEnv: "https",
			budgetHostEnv:   "micro.ms.dottics.com",
			E: E{
				scheme: "https",
				host:   "micro.ms.dottics.com",
				token:  "",
			},
		},
		{
			name:  "token",
			token: "my-test-token",
			E: E{
				scheme: "",
				host:   "",
				token:  "my-test-token",
			},
		},
		{
			name:            "token and env vars",
			budgetSchemeEnv: "https",
			budgetHostEnv:   "micro.ms.dottics.com",
			token:           "my-test-token",
			E: E{
				scheme: "https",
				host:   "micro.ms.dottics.com",
				token:  "my-test-token",
			},
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			err := os.Setenv("MICRO_SERVICE_SCHEME", tc.budgetSchemeEnv)
			err = os.Setenv("MICRO_SERVICE_HOST", tc.budgetHostEnv)
			if err != nil {
				t.Errorf("unexpected error before: %v", err)
			}

			s := NewService(Config{Name: "micro", UserToken: tc.token})
			xut := s.Header.Get("X-User-Token")
			if tc.E.token != xut {
				t.Errorf("expected '%v' got '%v'", tc.E.token, xut)
			}
			if tc.E.scheme != s.URL.Scheme {
				t.Errorf("expected '%v' got '%v'", tc.E.scheme, s.URL.Scheme)
			}
			if tc.E.host != s.URL.Host {
				t.Errorf("expected '%v' got '%v'", tc.E.host, s.URL.Host)
			}

			// reset to blank
			err = os.Setenv("MICRO_SERVICE_SCHEME", "")
			err = os.Setenv("MICRO_SERVICE_HOST", "")
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestService_SetURL(t *testing.T) {
	s := NewService(Config{Name: "micro"})
	s.SetURL("http", "micro.ms.test.dottics.com")
	if s.URL.Scheme != "http" {
		t.Errorf("expected '%v' got '%v'", "http", s.URL.Scheme)
	}
	if s.URL.Host != "micro.ms.test.dottics.com" {
		t.Errorf("expected '%v' got '%v'", "micro.ms.test.dottics.com", s.URL.Host)
	}
}

func TestService_SetEnv(t *testing.T) {
	s := Service{
		Name: "micro",
		URL: url.URL{
			Scheme: "https",
			Host:   "test.host.com",
		},
	}
	err := s.SetEnv()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	x1 := os.Getenv("MICRO_SERVICE_SCHEME")
	x2 := os.Getenv("MICRO_SERVICE_HOST")
	if x1 != "https" {
		t.Errorf("expected '%v' got '%v'", "https", x1)
	}
	if x2 != "test.host.com" {
		t.Errorf("expected '%v' got '%v'", "test.host.com", x2)
	}
}

func TestService_NewRequest(t *testing.T) {
	s := NewService(Config{Name: "micro", UserToken: "test-fake-token"})
	ms := microtest.MockServer(s)
	defer ms.Server.Close()

	// add an exchange to capture the request sent
	ex := &microtest.Exchange{
		Response: microtest.Response{
			Status: 200,
			Body:   `{"message":"successful request"}`,
		},
	}
	ms.Append(ex)

	// encode the url
	s.URL.Path = "/my/path"
	q := url.Values{}
	q.Add("u", "my value")
	// set the headers
	h := http.Header{
		"X-Random": {"my-random-header"},
	}
	// add the body
	p := strings.NewReader(`{"name":"james"}`)

	// now to make the request
	_, e := s.DoRequest("PUT", s.URL, q, h, p)

	if e != nil {
		t.Errorf("unexpected error: %v", e)
	}
	// test that the request was made correctly
	// test the request URI
	if ex.Request.RequestURI != "/my/path?u=my+value" {
		t.Errorf("expected '%v' got '%v'", "/my/path?u=my+value", ex.Request.RequestURI)
	}
	// test the headers
	h1 := ex.Request.Header.Get("X-Random")
	h2 := ex.Request.Header.Get("Content-Type")
	h3 := ex.Request.Header.Get("X-User-Token")
	if h1 != "my-random-header" {
		t.Errorf("expected '%v' got '%v'", "my-random-header", h1)
	}
	if h2 != "application/json" {
		t.Errorf("expected '%v' got '%v'", "application/json", h2)
	}
	if h3 != "test-fake-token" {
		t.Errorf("expected '%v' got '%v'", "test-fake-token", h3)
	}
	// test the body
}

func TestService_GetHome(t *testing.T) {
	type E struct {
		alive bool
		e     dutil.Err
	}
	tt := []struct {
		name     string
		exchange *microtest.Exchange
		E        E
	}{
		{
			name: "decode error",
			exchange: &microtest.Exchange{
				Response: microtest.Response{
					Status: 200,
					Body:   `{"message":"Welcome to the micro-service","data":{},"errors":{"internal_server_error":"server down for some reason"]}}`,
				},
			},
			E: E{
				alive: false,
				e: dutil.Err{
					Status: 500,
					Errors: map[string][]string{
						"unmarshal": {"invalid character ']' after object key:value pair"},
					},
				},
			},
		},
		{
			name: "500 internal server error",
			exchange: &microtest.Exchange{
				Response: microtest.Response{
					Status: 500,
					Body:   `{"message":"Welcome to the micro-service","data":{},"errors":{"internal_server_error":["server down for some reason"]}}`,
				},
			},
			E: E{
				alive: false,
				e: dutil.Err{
					Status: 500,
					Errors: map[string][]string{
						"internal_server_error": {"server down for some reason"},
					},
				},
			},
		},
		{
			name: "200 server alive",
			exchange: &microtest.Exchange{
				Response: microtest.Response{
					Status: 200,
					Body:   `{"message":"Welcome to the micro-service","data":{"alive":true},"errors":{}}`,
				},
			},
			E: E{
				alive: true,
				e:     dutil.Err{},
			},
		},
	}

	s := NewService(Config{Name: "micro"})
	ms := microtest.MockServer(s)
	defer ms.Server.Close()

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			// append the exchange for the test
			ms.Append(tc.exchange)

			alive, e := s.HealthCheck()
			if tc.E.alive != alive {
				t.Errorf("expected '%v' got '%v'", tc.E.alive, alive)
			}
			if tc.E.e.Status != 0 {
				if tc.E.e.Error() != e.Error() {
					t.Errorf("expected '%v' got '%v'", tc.E.e.Error(), e.Error())
				}
			}
		})
	}
}
