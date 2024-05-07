# 10. Adding validator library 

Date: 2024-05-01

## Status

Accepted

## Context

The application needed to be able to check information that is sent to the api and send back any errors. 

## Decision

We went with the [validator](https://github.com/go-playground/validator) library as it was well maintained and part of a larger ecosystem of Go Playground.
The library also offered away to do custom validation in a clear way.

## Consequences

There are some issues outstanding in the project but currently none of these are effecting what we are using it for.
