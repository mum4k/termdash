# Design goals

This effort is focused on good software design and maintainability. By good
design I mean:

1. Write readable, well documented code.
1. Only beautiful, simple APIs, no exposed concurrency, channels, internals, etc.
1. Follow [Effective Go](http://golang.org/doc/effective_go.html).
1. Provide an infrastructure that allows development of individual dashboard
   components in separation.
1. The infrastructure must enforce consistency in how the dashboard components
   are implemented.
1. Focus on maintainability, the infrastructure and dashboard components must
   have good test coverage, the repository must have CI/CD enabled.
