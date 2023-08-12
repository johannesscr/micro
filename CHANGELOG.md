## Changelog
All notable changes to this project will be documented in this file.

The format is based on [Keep a changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [Released]

## [0.2.0] - 2023-08-12
### Added
- A new type `Config` to hold the configuration of the microservice.

### Updated
- The `NewService` method to accept a `Config` type.
- The request functions to accept query parameters.

## [0.1.1] - 2022-04-17
### Changed
- updated the `Mock.Append` method to be able to use `nil` as an `Exchange`,
however, that exchange will be ignored in the sequence of exchange calls when
mocking a microservice.

## [0.1.0] - 2022-04-05
### Added
- The `microtest` package as a tool to easily test a `microservice` but more
  specifically testing of an API Gateway where GoLang is used as the API Gateway
  language. To be able to mock a `micorservice` instance and to mock each and
  every response from that `microservice`.
- The `msp` package which stands for `microservice package` and as the name
  may suggest it is used to scaffold the basic microservice integration package.
- Added basic README.