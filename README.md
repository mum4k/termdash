# termdash

This project implements a terminal based dashboard. The feature set is inspired
by the [gizak/termui](http://github.com/gizak/termui) project. Why the rewrite
you ask?

1. The above mentioned project is abandoned and isn't maintained anymore.
1. The project doesn't follow the design goals outlined below.
1. The project is released under a license I cannot use.

# Design goals

This effort is focused on good software design and maintainability. By a good
software I mean:

1. Write readable, well documented code.
1. Provide an infrastructure that allows development of individual dashboard
   components in separation.
1. The infrastructure should enforce consistency between the dashboard components.
1. Focus on maintainability, the infrastructure and dashboard components must
   have good test coverage, the repository must have CI/CD enabled.

On top of that - let's have fun, learn something and become better developers
together.
