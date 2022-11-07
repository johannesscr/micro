# Micro

## Microservice Package (MSP)

The `msp` package is basic scaffolding to write a testable microservice package.
The basic architecture would be
```
incoming-request-> Service -HTP-request-> Microservice or API
```
In this diagram the intermediary `Service` could be anything, a backend, 
an API Gateway or just another microservice that communicates with other 
microservices. 

The `msp` is there to easily setup the integration package between two services. 
Such that if Service 1 want to integrate/communicate/exchange with Service 2.
Ideally Service 2 would be responsible for writing the `Service 2 msp`. Service 1
would then import `Service 2 msp` and use the package to exchange with Service 2.
```
[Service1 (msp of Service 2)] -HTTP-request-> [Service 2]
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
import (
	"github.com/johannesscr/micro/msp"
)

type Service msp.Service

var microServiceName string = "micro"

// NewService creates a microservice-package (msp) instance. The msp
// is an instance that loads the environmental variables to be able
// to connect to the specific microservice. The msp contains all the
// implementations to correctly exchange with the microservice.
func NewService(token string) *Service {
	return (*Service)(msp.NewService(token, microServiceName))
}

// SetURL sets the scheme and host of the service. Also makes the service
// a mock-able service with `microtest`
func (s *Service) SetURL(scheme, host string) {
	s.URL.Scheme = scheme
	s.URL.Host = host
}
```

## Microtest

The `microtest` package is to help simplify testing of a microservice and more
specifically it mock the microservice.

### The basics
We look at the type definition of a `Mock` to get an idea of what is needed
to test using the `microtest` package.

