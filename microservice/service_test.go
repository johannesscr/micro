package microservice

import (
	"github.com/google/uuid"
	"github.com/johannesscr/micro/microtest"
	"net/http"
	"testing"
)

func TestNewService(t *testing.T) {
	s := NewService()
	if s == nil {
		t.Errorf("expected %#v got %v", Service{}, nil)
	}
}

func TestService_Name(t *testing.T) {
	s := NewService()
	if s.Name != "microservice" {
		t.Errorf("expected service name 'microservice'got %s", s.Name)
	}
}

func TestService_SetURL(t *testing.T) {
	s := NewService()
	if s.URL.Host != "" {
		t.Errorf("expected '%v' got '%v'", "", s.URL.Host)
	}
	sc := "http"
	h := "test.domain.com"
	p := "/user"

	s.SetURL(sc, h)
	s.URL.Path = p

	if s.URL.Host != h {
		t.Errorf("expected '%v' got '%v'", h, s.URL.Host)
	}
	if s.URL.Scheme != sc {
		t.Errorf("expected '%v' got '%v'", sc, s.URL.Scheme)
	}
	if s.URL.String() != "http://test.domain.com/user" {
		t.Errorf("expected '%v' got '%v'", "http://test.domain.com/user", s.URL.String())
	}
}

func TestService_SetEnv(t *testing.T) {
	s := NewService()
	if s.URL.Host != "" {
		t.Errorf("expected '' got '%v'", s.URL.Host)
	}
	if s.URL.Scheme != "" {
		t.Errorf("expected '' got '%v'", s.URL.Scheme)
	}

	// create the dynamic micro-service instance
	ms := microtest.MockServer(s)
	defer ms.Server.Close()

	// set the service environmental variables
	err := s.SetEnv()
	if err != nil {
		t.Errorf("unable to set environmental variables dynamically")
	}

	if s.URL.String() != ms.Server.URL {
		t.Errorf("expected '%v' got '%v'", ms.Server.URL, s.URL.String())
	}
}

func TestService_GetHome(t *testing.T) {
	s := NewService()
	ms := microtest.MockServer(s)
	defer ms.Server.Close()

	e := &microtest.Exchange{
		Response: microtest.Response{
			Status: 200,
			Header: map[string][]string{
				"Content-Type": {"application/json", "charset=utf-8"},
			},
		},
	}
	ms.Append(e)

	res := s.GetHome()
	if res.StatusCode != http.StatusOK {
		t.Errorf("expected %d got %d", http.StatusOK, res.StatusCode)
	}
}

func TestService_GetUser(t *testing.T) {
	// create a microservice instance
	s := NewService()
	// startup the microservice
	ms := microtest.MockServer(s)
	defer ms.Server.Close() // defer shut down the microservice

	e := &microtest.Exchange{
		Response: microtest.Response{
			Status: 200,
			Header: map[string][]string{"x-token": {"124"}},
			Body: `{
				"message": "user found successfully", 
				"data": {
					"user": {
						"uuid": "6a67f46e-d9de-4d63-8283-bf5a5aa1e582", 
						"first_name": "james", 
						"last_name": "bond", 
						"email": "007@mi6.co.uk"
					}
				}, 
				"errors": {}
			}`,
		},
	}
	ms.Append(e)

	u, errors := s.GetUser(uuid.New().String())
	if errors != nil {
		t.Errorf("errors on response: %v", errors)
	}
	if u.FirstName != "james" {
		t.Errorf("expected '%v' got '%v'", "james", u.FirstName)
	}
	if u.LastName != "bond" {
		t.Errorf("expected '%v' got '%v'", "bond", u.LastName)
	}
}
