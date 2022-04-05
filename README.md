# Micro

## msp

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

## microtest

The `microtest` package is to help simplify testing of a microservice and more
specifically it mock the microservice.

