# 11. Use Testify suites for integration testing

Date: 2024-05-31

## Status

Accepted

## Context

We currently use `TestMain` to set up and teardown the TestContainers test instance before and after each test in the 
`service` package. However, it is useful to be able to also unit test functions within that package independent of the 
database, as it would allow for a faster feedback loop and quicker builds. 

## Decision

We will use the Testify library to solve this issue. This is a library we already have as a dependency for its improved
assertions, but it also includes functionality to create custom test suites. This allows us to have an `IntegrationSuite`
that contains the database as well as the before and after functions for creating and tearing it down. Refactoring to this
pattern is straight forward, and each test that requires access to a running database just needs to become a member function
of the suite, as opposed to taking `testing.T` as you would with a standard test function. e.g.

```
func TestService_Example(t *testing.T)
```
becomes
```
func (suite *IntegrationSuite) TestService_Example()
```

## Consequences

This is an existing external dependency that is not included in the build.