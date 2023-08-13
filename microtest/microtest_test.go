package microtest

import (
	"bytes"
	"fmt"
	"github.com/johannesscr/micro/microservice"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

func TestMockServer(t *testing.T) {
	s := microservice.NewService()
	ms := MockServer(s)
	defer ms.Server.Close()

	e := &Exchange{
		Response: Response{
			Status: 200,
			Body: `{
				"data": {},
				"errors": {},
				"message": "Welcome to the POS api"
			}`,
		},
	}
	ms.Append(e)

	b, _ := s.GetHome()

	if !b {
		t.Errorf("failed to create mock server")
	}
}

func TestMock_Append(t *testing.T) {
	m := &Mock{}
	if len(m.Exchanges) != 0 {
		t.Errorf("expected 0 got %d", len(m.Exchanges))
	}

	e1 := &Exchange{
		Response: Response{Status: 201},
	}
	e2 := &Exchange{
		Response: Response{Status: 202},
	}

	m.Append(e1)
	if len(m.Exchanges) != 1 {
		t.Errorf("expected 1 got %d", len(m.Exchanges))
	}
	if m.Exchanges[0].Response.Status != e1.Response.Status {
		t.Errorf("expected %d got %d", e1.Response.Status, m.Exchanges[0].Response.Status)
	}

	m.Append(e2)
	if len(m.Exchanges) != 2 {
		t.Errorf("expected 2 got %d", len(m.Exchanges))
	}
	if m.Exchanges[1].Response.Status != e2.Response.Status {
		t.Errorf("expected %d got %d", e2.Response.Status, m.Exchanges[1].Response.Status)
	}
}

func TestMock_mockHandler(t *testing.T) {
	ms := MockServer(&Mock{})
	URI := ms.URL.String()

	e := &Exchange{
		Response: Response{
			Status: 203,
			Header: map[string][]string{
				"Content-Type": {"text/text"},
			},
			Body: "hi",
		},
	}
	ms.Append(e)

	req := httptest.NewRequest("GET", URI+"/one", nil)
	rec := httptest.NewRecorder()

	handler := ms.mockHandler()
	handler(rec, req)

	res, bs := ReadRecorder(rec)
	if res.StatusCode != e.Response.Status {
		t.Errorf("expected %d got %d", e.Response.Status, res.StatusCode)
	}
	if string(bs) != e.Response.Body {
		t.Errorf("expected '%v' got '%v'", e.Response.Body, string(bs))
	}
	for key, values := range e.Response.Header {
		h := res.Header.Get(key)
		v := values[0]
		if h != v {
			t.Errorf("%v: expected '%v' got '%v'", key, v, h)
		}
	}
}

func TestMock_mockHandler_request(t *testing.T) {
	ms := MockServer(&Mock{})
	URI := ms.URL.String()

	e := &Exchange{
		Response: Response{
			Status: 203,
		},
	}
	ms.Append(e)

	ior := strings.NewReader("hi")
	req := httptest.NewRequest("GET", URI+"/one", ior)
	req.Header.Set("Content-Type", "text/text")
	req.Header.Set("X-Random", "123-456")
	rec := httptest.NewRecorder()

	handler := ms.mockHandler()
	handler(rec, req)

	if e.Request.URL.String() != req.URL.String() {
		t.Errorf("expected '%v' got '%v'",
			req.URL.String(), e.Request.URL.String())
	}

	if e.Request.Header.Get("Content-Type") != "text/text" {
		t.Errorf("expected '%v' got '%v'",
			"text/text", e.Request.Header.Get("Content-Type"))
	}
	if e.Request.Header.Get("X-Random") != "123-456" {
		t.Errorf("expected '%v' got '%v'",
			"123-456", e.Request.Header.Get("X-Random"))
	}

	xb, _ := ioutil.ReadAll(e.Request.Body)
	defer func() {
		_ = e.Request.Body.Close()
	}()
	if string(xb) != "hi" {
		t.Errorf("expected '%v', got '%v'", "hi", string(xb))
	}
}

func TestMock_transmit_error(t *testing.T) {
	m := &Mock{}
	_, err := m.transmit(nil)

	eString := fmt.Sprintf("%s", err)
	EErr := "map[transmission:[exceeded mock request/response exchange transmissions]]"

	if eString != EErr {
		t.Errorf("expected '%v' got '%v'", EErr, eString)
	}
}

func TestMock_transmit_response(t *testing.T) {
	m := &Mock{}
	if m.transmission != 0 {
		t.Errorf("expected 0 got %d", m.transmission)
	}

	e := &Exchange{
		Response: Response{Status: 203, Header: nil, Body: "hi"},
	}
	m.Append(e)

	res, err := m.transmit(&http.Request{})
	if m.transmission != 1 {
		t.Errorf("expected 1 got %d", m.transmission)
	}
	if err != nil {
		t.Errorf("expected nil got %v", err)
	}
	if res.Status != e.Response.Status {
		t.Errorf("expected %d got %d", e.Response.Status, res.Status)
	}
}

func TestMock_SetURL(t *testing.T) {
	m := &Mock{
		URL: url.URL{
			Scheme: "tcp",
			Host:   "mydefaultdomain",
		},
	}
	m.SetURL("https", "mytestdomain")

	if m.URL.Scheme != "https" {
		t.Errorf("expected '%v' got '%v'", "https", m.URL.Scheme)
	}
	if m.URL.Host != "mytestdomain" {
		t.Errorf("expected '%v' got '%v'", "mytestdomain", m.URL.Host)
	}
}

func TestReadRecorder(t *testing.T) {
	type testCase struct {
		name   string
		status int
		header map[string][]string
		body   []byte
	}
	tt := []testCase{
		{
			name:   "only status",
			status: 200,
		},
		{
			name:   "with headers",
			status: 200,
			header: map[string][]string{
				"Content-Type": {"application/json; charset=utf-8"},
			},
		},
		{
			name:   "with body",
			status: 200,
			body:   []byte("hi there"),
		},
		{
			name:   "json body",
			status: 200,
			header: map[string][]string{
				"Content-Type": {"application/json; charset=utf-8"},
			},
			body: []byte(`{"message": "test message"}`),
		},
	}

	handler := func(w http.ResponseWriter, tc testCase) {
		for key, values := range tc.header {
			w.Header().Set(key, values[0])
		}
		w.WriteHeader(tc.status)
		_, err := w.Write(tc.body)
		if err != nil {
			log.Panic(err)
		}
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			// set recorder
			rec := httptest.NewRecorder()

			// call the handler
			// From https://pkg.go.dev/net/http/httptest#ResponseRecorder
			handler(rec, tc)

			// read the recorder
			res, xb := ReadRecorder(rec)
			if res.StatusCode != tc.status {
				t.Errorf("expected %d got %d", tc.status, res.StatusCode)
			}

			for key, values := range tc.header {
				h := res.Header.Get(key)
				v := values[0]
				if h != v {
					t.Errorf("%v: expected '%v' got '%v'", key, v, h)
				}
			}

			if string(xb) != string(tc.body) {
				t.Errorf("expected '%v' got '%v'", string(xb), string(tc.body))
			}
		})
	}

}

func TestNewRequest(t *testing.T) {
	tt := []struct {
		name     string
		method   string
		target   string
		query    url.Values
		headers  http.Header
		body     io.Reader
		EMethod  string
		ETarget  string
		EQuery   string
		EHeaders string
		EBody    string
	}{
		{
			name:     "default",
			method:   "",
			target:   "/",
			EMethod:  "GET",
			ETarget:  "/",
			EQuery:   "",
			EHeaders: "map[]",
			EBody:    "",
		},
		{
			name:   "set query parameters",
			method: "POST",
			target: "/resource/-",
			query: url.Values{
				"a": {"a random value"},
				"q": {"12df34al-389j-23d9jd"},
			},
			EMethod:  "POST",
			ETarget:  "/resource/-",
			EQuery:   "a=a+random+value&q=12df34al-389j-23d9jd",
			EHeaders: "map[]",
			EBody:    "",
		},
		{
			name:   "set headers",
			method: "POST",
			target: "/resource/-",
			headers: map[string][]string{
				"Content-Type": {"application/json", "charset=utf-8"},
			},
			EMethod:  "POST",
			ETarget:  "/resource/-",
			EQuery:   "",
			EHeaders: "map[Content-Type:[application/json charset=utf-8]]",
			EBody:    "",
		},
		{
			name:     "set body from string",
			method:   "POST",
			target:   "/resource/-",
			body:     strings.NewReader("this is my reader"),
			EMethod:  "POST",
			ETarget:  "/resource/-",
			EQuery:   "",
			EHeaders: "map[]",
			EBody:    "this is my reader",
		},
		{
			name:     "set body from bytes",
			method:   "POST",
			target:   "/resource/-",
			body:     bytes.NewReader([]byte("this is my reader")),
			EMethod:  "POST",
			ETarget:  "/resource/-",
			EQuery:   "",
			EHeaders: "map[]",
			EBody:    "this is my reader",
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			r := NewRequest(tc.method, tc.target, tc.query, tc.headers, tc.body)

			if r.Method != tc.EMethod {
				t.Errorf("expected '%v' got '%v'", tc.EMethod, r.Method)
			}
			if r.URL.Path != tc.target {
				t.Errorf("expected '%v' got '%v'", tc.target, r.URL.Path)
			}
			if r.URL.RawQuery != tc.EQuery {
				t.Errorf("expected '%v' got '%v'", tc.EQuery, r.URL.RawQuery)
			}
			hString := fmt.Sprintf("%s", r.Header)
			if hString != tc.EHeaders {
				t.Errorf("expected '%v' got '%v'", tc.EHeaders, hString)
			}
			xb, err := ioutil.ReadAll(r.Body)
			err = r.Body.Close()
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			bs := fmt.Sprintf("%s", xb)
			if bs != tc.EBody {
				t.Errorf("expected '%v' got '%v'", tc.EBody, bs)
			}
		})
	}
}

func TestNewErr(t *testing.T) {
	key := "key-value"
	errors := []string{"error description one", "error description two"}

	e := NewErr(key, errors)
	if e.errors["key-value"][0] != errors[0] {
		t.Errorf("expected '%v' got  '%v'", errors[0], e.errors["key-value"][0])
	}
	if e.errors["key-value"][1] != errors[1] {
		t.Errorf("expected '%v' got  '%v'", errors[1], e.errors["key-value"][1])
	}
}

func TestErr_Error(t *testing.T) {
	e := &Err{
		errors: map[string][]string{
			"key-value": {"error description one", "error description two"},
		},
	}

	eString := e.Error()
	if eString != "map[key-value:[error description one error description two]]" {
		t.Errorf("expected '%v' got '%v'", "map[key-value:[error description one error description two]]", eString)
	}
}
