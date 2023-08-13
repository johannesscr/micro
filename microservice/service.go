package microservice

import (
	"github.com/johannesscr/micro/msp"
	"log"
	"net/url"
)

// Service id the shorthand for the integration to the microservice
type Service struct {
	// defining the msp.Service as an embedded type allows us to access all the
	// methods of the msp.Service without having to redefine them here.
	msp.Service
}

// NewService creates a new instance of the microservice
func NewService() *Service {
	// here we need to define the config for the msp.Service, if additional
	// base config is required we can add / override it here.
	config := msp.Config{
		Name: "microservice",
	}

	// here we just convert the msp.Service to a MicroService as msp.Service
	// is a package we cannot extend it, so we create a new type that is
	// identical to msp.Service and add the methods we need to it.
	s := &Service{
		Service: *msp.NewService(config),
	}
	return s
}

// GetUser returns a user from the Micro-Service
func (s *Service) GetUser(uUUID string) (User, map[string][]string) {
	// set the path
	s.URL.Path = "/user/-"

	// set the query parameters
	q := url.Values{}
	q.Add("uuid", uUUID)

	resp := struct {
		HTTPCode int                 `json:"http_code"`
		Message  string              `json:"message"`
		Data     map[string]User     `json:"data"`
		Errors   map[string][]string `json:"errors"`
	}{}

	res, e := s.DoRequest("GET", s.URL, q, nil, nil)
	if e != nil {
		log.Println(e)
	}
	bs, _ := msp.Decode(res, &resp)
	if res.StatusCode != 200 {
		log.Printf("response: %s", string(bs))
		log.Println(resp.Errors)
		return User{}, resp.Errors
	}
	return resp.Data["user"], nil
}
