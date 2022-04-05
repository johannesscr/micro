package msp

import (
	"encoding/json"
	"fmt"
	"github.com/dottics/dutil"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
)

type Service struct {
	Name   string
	Header http.Header
	URL    url.URL
}

// NewService creates a microservice-package instance. The
// microservice-package is an instance that loads the environmental
// variables to be able to connect to the specific microservice. The
// microservice-package contains all the implementations to correctly
// exchange with the microservice.
func NewService(token, serviceName string) *Service {
	s := &Service{
		Name: serviceName,
		URL: url.URL{
			Scheme: os.Getenv(fmt.Sprintf("%s_SERVICE_SCHEME", strings.ToUpper(serviceName))),
			Host:   os.Getenv(fmt.Sprintf("%s_SERVICE_HOST", strings.ToUpper(serviceName))),
		},
		Header: make(http.Header),
	}
	// default microservice headers
	s.Header.Set("Content-Type", "application/json")
	s.Header.Set("X-User-Token", token)
	return s
}

// SetURL sets the URL for the microservice-package to point to the service.
//
// SetURL is also the function which makes the service a mock service
// interface.
func (s *Service) SetURL(scheme, host string) {
	s.URL.Scheme = scheme
	s.URL.Host = host
}

// SetEnv is used for testing, when the dynamic microservice is created
// then SetEnv is used to dynamically set the env vars for the temporary
// service.
func (s *Service) SetEnv() error {
	err := os.Setenv(fmt.Sprintf("%s_SERVICE_SCHEME", strings.ToUpper(s.Name)), s.URL.Scheme)
	if err != nil {
		return err
	}
	err = os.Setenv(fmt.Sprintf("%s_SERVICE_HOST", strings.ToUpper(s.Name)), s.URL.Host)
	if err != nil {
		return nil
	}
	return nil
}

// NewRequest consistently maps and executes requests to the requirements
// for the service and returns the response.
func (s *Service) NewRequest(method, url string, headers http.Header, payload io.Reader) (*http.Response, dutil.Error) {
	client := http.Client{}
	req, _ := http.NewRequest(method, url, payload)
	// set the default service headers
	req.Header = s.Header
	// set / override additional headers iff necessary
	for key, values := range headers {
		req.Header.Set(key, values[0])
	}
	// send the request
	res, err := client.Do(req)
	log.Printf("- %s-service -> [%s %s] <- %d", s.Name, req.Method, req.URL.String(), res.StatusCode)
	// if there was an error making the request not an error response
	if err != nil {
		e := dutil.NewErr(500, "request", []string{err.Error()})
		return nil, e
	}
	return res, nil
}

// Decode is a function that decodes a body into a slice of bytes and also
// will unmarshal the data into an interface pointer value if the value
// pointed to by the interface is provided.
func (s *Service) Decode(res *http.Response, v interface{}) ([]byte, dutil.Error) {
	xb, err := ioutil.ReadAll(res.Body)
	if err != nil {
		e := dutil.NewErr(500, "read", []string{err.Error()})
		return nil, e
	}
	err = res.Body.Close()
	if err != nil {
		e := dutil.NewErr(500, "readClose", []string{err.Error()})
		return nil, e
	}
	if v != nil {
		err = json.Unmarshal(xb, v)
		if err != nil {
			e := dutil.NewErr(500, "unmarshal", []string{err.Error()})
			return nil, e
		}
	}
	return xb, nil
}

// GetHome is the health-check function which makes a request to the
// microservice to check that the service is still up and running.
// Simply return a true if a request is successful.
func (s *Service) GetHome() (bool, dutil.Error) {
	s.URL.Path = "/"

	resp := struct {
		Message string              `json:"message"`
		Data    interface{}         `json:"data"`
		Errors  map[string][]string `json:"errors"`
	}{}

	res, e := s.NewRequest("GET", s.URL.String(), nil, nil)
	if e != nil {
		return false, e
	}
	_, e = s.Decode(res, &resp)
	if e != nil {
		return false, e
	}

	// manage the response separately
	if res.StatusCode != 200 {
		e := &dutil.Err{
			Status: res.StatusCode,
			Errors: resp.Errors,
		}
		return false, e
	}
	return true, nil
}
