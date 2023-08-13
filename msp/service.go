package msp

import (
	"fmt"
	"github.com/dottics/dutil"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
)

// Service is the microservice-package structure that contains all the
// information required to connect to the microservice.
type Service struct {
	Name   string
	Header http.Header
	URL    url.URL
	Values url.Values
}

// Config is the configuration for the microservice-package.
type Config struct {
	Name      string
	UserToken string
	APIKey    string
	Header    http.Header
	URL       url.URL
	Values    url.Values
}

// NewService creates a microservice-package instance. The
// microservice-package is an instance that loads the environmental
// variables to be able to connect to the specific microservice. The
// microservice-package contains all the implementations to correctly
// exchange with the microservice.
func NewService(config Config) *Service {
	s := &Service{
		Name: config.Name,
		URL: url.URL{
			Scheme: os.Getenv(fmt.Sprintf("%s_SERVICE_SCHEME", strings.ToUpper(config.Name))),
			Host:   os.Getenv(fmt.Sprintf("%s_SERVICE_HOST", strings.ToUpper(config.Name))),
		},
		Header: make(http.Header),
		Values: make(url.Values),
	}
	// set config headers if given
	if config.Header != nil {
		s.Header = config.Header
	}
	// set config values if given
	if config.Values != nil {
		s.Values = config.Values
	}
	// default microservice headers
	s.Header.Set("content-type", "application/json")
	s.Header.Set("x-user-token", config.UserToken)
	s.Header.Set("x-api-key", config.APIKey)

	return s
}

// SetURL sets the URL for the Security Micro-Service to point to
// SetURL is also the interface that makes it a mock service
func (s *Service) SetURL(sc string, h string) {
	s.URL.Scheme = sc
	s.URL.Host = h
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

// DoRequest consistently maps and executes requests to the requirements
// for the service and returns the response.
func (s *Service) DoRequest(method string, URL url.URL, query url.Values, headers http.Header, payload io.Reader) (*http.Response, dutil.Error) {
	client := http.Client{}
	qs := s.Values
	// set / override additional query params iff necessary
	for key, values := range query {
		for _, value := range values {
			qs.Add(key, value)
		}
	}
	// set the query params
	URL.RawQuery = qs.Encode()

	// create the request
	req, _ := http.NewRequest(method, URL.String(), payload)

	// set the default service headers
	req.Header = s.Header
	// set / override additional headers iff necessary
	for key, values := range headers {
		for _, value := range values {
			req.Header.Add(key, value)
		}
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

// HealthCheck is the health-check function which makes a request to the
// microservice to check that the service is still up and running.
// Simply return a true if a request is successful.
func (s *Service) HealthCheck() (bool, dutil.Error) {
	s.URL.Path = "/"

	resp := struct {
		Message string              `json:"message"`
		Data    interface{}         `json:"data"`
		Errors  map[string][]string `json:"errors"`
	}{}

	res, e := s.DoRequest("GET", s.URL, nil, nil, nil)
	if e != nil {
		return false, e
	}
	_, e = Decode(res, &resp)
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
