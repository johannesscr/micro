# Micro

Goal:
- create a microservice package (`msp`) that can be used to communicate with a 
  microservice using HTTP exchanges.
  - be able to mock the microservice for testing purposes.
- create a microservice test package (`microtest`) that can be used to mock a
  microservice for testing purposes.

## Microservice Package (MSP)

The `msp` package is basic scaffolding to write a testable microservice package.
The basic architecture would be
```
incoming-request -> Service-HTTP-request-> Microservice or API
```
In this diagram the intermediary `Service` could be anything, a backend, 
an API Gateway or just another microservice that communicates with other 
microservices. 

The `msp` is there to easily set up the integration package between two services. 
Such that if Service 1 want to integrate/communicate/exchange with Service 2.
Ideally Service 2 would be responsible for writing the `Service 2 msp`. Service 1
would then import `Service 2 msp` and use the package to exchange with Service 2.
```
[[Service1]-[msp of Service 2](args...)]  --HTTP-request-->  [Service 2]
```

However, if you are Service 1, you would want to be able to test the various
responses from Service 2. Then use the `microtest` package.

#### How to create a MSP

For the microservice package (msp) to be able to communicate with the 
microservice the msp needs to know the Scheme (e.g. http) and Host (domain.com)
where to send the exchange request, these are set as environmental variables.

Suppose we want to connect to a microservice called *MICRO*

```bash
# we set the environmental variables using the convection {microservice-name}_SERVICE_{SCHEME|HOST} as 
export MICRO_SERVICE_SCHEME=http 
export MICRO_SERVICE_HOST=service.com 
```

Unfortunately there is some boilerplate code needed. (But I will work on making
this better as I learn more about Go).

```go 
package microservice

import (
	"github.com/johannesscr/micro/msp"
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
  //s := (*Service)(msp.NewService(config))
  s := &Service{
    Service: *msp.NewService(config),
  }
  return s
}
```

Then to add integration methods with the microservice we can add them to the
`Service` type.

```go
package microservice

import (
    "github.com/johannesscr/micro/msp"
	"log"
)

// GetUser returns a user from the microservice
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

  // DoRequest is a method of the msp.Service that will do the request to the
  // microservice and return the response.
  res, e := s.DoRequest("GET", s.URL, q, nil, nil)
  if e != nil {
    log.Println(e)
  }
  // Decode is a function added to the msp package to decode the response from
  // the microservice to JSON.
  bs, _ := msp.Decode(res, &resp)
  if res.StatusCode != 200 {
    log.Printf("response: %s", string(bs))
    log.Println(resp.Errors)
    return User{}, resp.Errors
  }
  return resp.Data["user"], nil
}
```

## Microtest

The `microtest` package is to help simplify testing of a microservice and more
specifically it mock the microservice.

### The basics
We look at the type definition of a `Mock` to get an idea of what is needed
to test using the `microtest` package.

### Example

Here is an example of how to use the `microtest` package to mock a microservice.

```go
import (
    "github.com/johannesscr/micro/microservice"
    "github.com/johannesscr/micro/microtest"
)

func TestMockServer(t *testing.T) {
    s := microservice.NewService()
    ms := microtest.MockServer(s)
    defer ms.Server.Close()

    e := &microtest.Exchange{
        Response: microtest.Response{
            Status: 200,
            Body: `{
                "data": {},
                "errors": {},
                "message": "Welcome to the POS api"
            }`,
        },
    }
	// we can append as many exchanges as we want, which will be returned
	// as a queue (FIFO) when the microservice is called.
    ms.Append(e)

    res := s.GetHome()

    if res.StatusCode != 200 {
        t.Errorf("failed to create mock server")
    }
}
```

