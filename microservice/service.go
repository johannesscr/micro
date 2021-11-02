package microservice

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
)

// Service id the shorthand for the integration to the Security Micro-Service
type service struct {
	URL url.URL
}

func NewService() *service {
	s := &service{
		URL: url.URL{
			Scheme: os.Getenv("MICROSERVICE_SCHEME"),
			Host: os.Getenv("MICROSERVICE_HOST"),
		},
	}
	return s
}

// SetURL sets the URL for the Security Micro-Service to point to
// SetURL is also the interface that makes it a mock service
func (s *service) SetURL(sc string, h string) {
	s.URL.Scheme = sc
	s.URL.Host = h
}

// SetEnv set the current service scheme and host as environmental variable
//
// Mostly used for testing when the Env Vars need to be set dynamically when
// service instances need to be created and stopped by the tests.
func (s *service) SetEnv() error {
	err := os.Setenv("MICROSERVICE_SCHEME", s.URL.Scheme)
	if err != nil {
		return err
	}
	err = os.Setenv("MICROSERVICE_HOST", s.URL.Host)
	if err != nil {
		return err
	}
	return nil
}

// GetHome is a PING function to test connection to the Security Micro-Service
// is healthy
func (s *service) GetHome() *http.Response {
	res, err := http.Get(s.URL.String())
	if err != nil {
		log.Println(err)
	}
	return res
}

func (s *service) GetUser(uUUID string) (User, map[string][]string) {
	q := url.Values{}
	q.Add("uuid", uUUID)

	s.URL.Path = "/user/-"
	s.URL.RawQuery = q.Encode()

	resp := struct {
		HTTPCode int                 `json:"http_code"`
		Message  string              `json:"message"`
		Data     map[string]User     `json:"data"`
		Errors   map[string][]string `json:"errors"`
	}{}

	client := http.Client{}
	req, _ := http.NewRequest("GET", s.URL.String(), nil)
	req.Header.Set("x-token", "eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJzLXRva2VuIjoiMGRhNzk1MjItY2VjZS00YWFkLTllMmEtZjQ1MzkwOGRlNTVmIiwicy1pZCI6Ijc4OWU4NDAxMTIyYTBmYmQ2M2NkM2JjNjhkMTQ5NzlmODc3NjZiMTk1MzdiZThkYmRjNDFmNTE4ZDFjZWViY2QiLCJ1LWlkIjoiMWNhMGFlNjgtMWJmMi00YTE4LWE4MTktYmU1YWE4MGVkOThlIiwiY3JlYXRlZCI6IjA5LzI5LzIwMjEsIDA4OjEwOjUyIn0.ZVNDocuNd760gwJFLY5V5Mg_gBf8I1oydMOvTqJes6M")
	res, err := client.Do(req)
	if err != nil {
		log.Println(err)
	}

	bs, _ := ioutil.ReadAll(res.Body)
	//log.Printf("%s\n", bs)
	err = res.Body.Close()
	if err != nil {
		log.Println(err)
	}

	err = json.Unmarshal(bs, &resp)
	if err != nil {
		log.Println(err)
		return User{}, resp.Errors
	}
	return resp.Data["user"], nil
}
