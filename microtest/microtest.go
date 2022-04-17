package microtest

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
)

// mock is the interface that connects all micro-services
type mock interface {
	SetURL(scheme string, host string)
}

// Response contains the basic fields required to mock a response to be
// expected to be sent back from any micro-service.
type Response struct {
	Status int
	Header map[string][]string
	Body   string
}

// Exchange is a Request / Response pair as defined by the IETF RFC2616
// https://datatracker.ietf.org/doc/html/rfc2616#section-1.4
// between two servers when using HTTP.
type Exchange struct {
	Response Response
	Request  *http.Request
}

// Mock server structure that groups the URL to which the mock server should
// connect, the mock server itself, the series of exchanges as defined by an
// Exchange and a counter to count the number of transmissions that have
// occurred.
type Mock struct {
	URL          url.URL
	Server       *httptest.Server
	Exchanges    []*Exchange
	transmission int
}

// MockServer takes any mock or mock-able micro-service and creates a
// mock http.Server and a Mock structure to aggregate all the mocked methods
// together.
func MockServer(mx mock) *Mock {
	m := &Mock{
		transmission: 0,
	}
	m.Server = m.mockServer(mx)
	return m
}

// NewMockServer takes any mock or mock-able micro-service and creates a
// mock http.Server and a Mock structure to aggregate all the mocked methods
// together.
func NewMockServer(envScheme string, envHost string) *Mock {
	m := &Mock{
		transmission: 0,
	}
	m.Server = m.mockServer(m)
	err := os.Setenv(envScheme, m.URL.Scheme)
	if err != nil {
		log.Panic(err)
	}
	err = os.Setenv(envHost, m.URL.Host)
	if err != nil {
		log.Panic(err)
	}
	return m
}

// SetURL makes the Mock also of type mock interface
func (m *Mock) SetURL(scheme string, host string) {
	m.URL.Scheme = scheme
	m.URL.Host = host
}

// MockServer takes a type mock interface, the type mock interface is the
// interface for any micro-service. Due to go routing any request to the mock
// handler the type mock interface which points (via the URL) to the
// MockServer can return any response provided for any request make to the
// type mock interface
func (m *Mock) mockServer(mx mock) *httptest.Server {
	mockServer := httptest.NewServer(m.mockHandler())

	xs := strings.Split(mockServer.URL, "/")
	scheme := strings.Replace(xs[0], ":", "", 1)
	host := strings.Join(xs[2:], "")
	mx.SetURL(scheme, host)
	return mockServer
}

// Append adds an Exchange to the queue (Q) of exchanges between the
// api-gateway and the microservice. Exchanges in the Q are processed a
// First-In-First-Out (FIFO) manner.
//
// If a nil Exchange is passed then ignore the exchange.
func (m *Mock) Append(e *Exchange) {
	if e == nil {
		return
	}
	m.Exchanges = append(m.Exchanges, e)
}

// transmit mocks the action where the micro-service receives the request
// and keeps a reference to the request pointed to and returning the response
// that should be responded with from the mock micro-service.
func (m *Mock) transmit(r *http.Request) (Response, error) {
	if m.transmission == len(m.Exchanges) {
		return Response{}, NewErr("transmission", []string{"exceeded mock request/response exchange transmissions"})
	}

	e := m.Exchanges[m.transmission]
	e.Request = r
	m.transmission++
	return e.Response, nil
}

// mockHandler takes the request properties defined on the Mock and writes
// it to the response of the mockServer which is a mock representing the
// micro-service being tested
func (m *Mock) mockHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		//log.Println(m.Response.Status)
		//log.Println(m.Response.Header)
		//log.Println(m.Response.Body)
		res, err := m.transmit(r)
		if err != nil {
			log.Panic(err)
		}

		for key, values := range res.Header {
			w.Header().Set(key, values[0])
		}
		w.WriteHeader(res.Status)
		_, err = w.Write([]byte(res.Body))
		if err != nil {
			log.Panic(err)
		}
	}
}

// ReadRecorder reads the recorder to get the response and decodes the body
// to a slice of bytes.
func ReadRecorder(rec *httptest.ResponseRecorder) (*http.Response, []byte) {
	res := rec.Result()
	xb, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Panic(err)
	}
	err = res.Body.Close()
	if err != nil {
		log.Panic(err)
	}
	return res, xb
}

// NewRequest is based on a httptest.NewRequest and makes it easy to also
// add the query parameters.
func NewRequest(method string, target string, query map[string]string, headers map[string][]string, body io.Reader) *http.Request {
	// new request
	r := httptest.NewRequest(method, target, body)
	// set headers
	r.Header = headers
	// set query params
	q := url.Values{}
	for key, val := range query {
		q.Add(key, val)
	}
	r.URL.RawQuery = q.Encode()
	return r
}

// DoRequest is based on a httptest.NewRequest and makes it easy to also
// add the query parameters.
func DoRequest(method string, target string, query map[string]string, headers map[string][]string, body io.Reader) *http.Response {
	c := http.Client{}
	// new request
	r, err := http.NewRequest(method, target, body)
	if err != nil {
		log.Panicf("unexpected err: %v", err)
	}
	// set headers
	r.Header = headers
	// set query params
	q := url.Values{}
	for key, val := range query {
		q.Add(key, val)
	}
	r.URL.RawQuery = q.Encode()

	res, err := c.Do(r)
	if err != nil {
		log.Panicf("unexpected err: %v", err)
	}
	return res
}

// Err is set defined structure for representing errors
type Err struct {
	errors map[string][]string
}

// NewErr creates a micro-test error struct and returns the and returns an Err struct
// being pointed to.
func NewErr(key string, errors []string) *Err {
	return &Err{
		errors: map[string][]string{
			key: errors,
		},
	}
}

// Error returns a string representation of Err. Also makes it of the
// type error interface.
func (e *Err) Error() string {
	return fmt.Sprintf("%s", e.errors)
}
