#errors

[![GoDoc](https://godoc.org/github.com/atdiar/errors?status.svg)](https://godoc.org/github.com/atdiar/errors)

errors - A drop-in replacement for the errors package
-------------------------------------------------------------

This package defines an `Error` type which, obviously, implements the `error` interface.

It can be used as a replacement to the errors package that can be found in the standard library.

A few facilities have been added:
* An error can be decorated with additional information such as the date, time or file line at which it occured.
* the specification of an error encoding/decoding format (Codec) can be provided for wireframe sending.


For completeness, please refer to the package [documentation].

[documentation]:https://godoc.org/github.com/atdiar/errors
