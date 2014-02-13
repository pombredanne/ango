## Ango: Angular <-> Go communication

`ango` is a tool that generates a protocol for communication between [Go](http://golang.org) and [AngularJS](http://angularjs.org) over http/websockets.

**This project is still under development, please do look arround and let me know if you have good idea's**

### Goals

The main goals are:
 - async server > client RPC
 - async client > server RPC
 - typed arguments and return values (see types below)
 - work with Go packages `net/http` and `encoding/json`
 - integrate into AngularJS as includable module + ng service
 - underlying protocol and communication is not directly visible for user. Calls feel native/local.

What I don't want to do:
 - runtime discovery of available procedures
 - A generic protocol designed for multiple types of servers/clients

Therefore I chose to generate go and javascript, so both server and client hold all information to communicate.

NEEDS THINKING: Code generated for Go can be copied into any go package. The code doesn't form a package itself.

For angular a single `.js` file is generated  holding an angular module. The module can be included by any other angular module.

### Terms

A **service** exists of one or more **procedures** server- and/or client-side.
A **procedure** within a service resides client- or server-side, and can be called by the other party.

### .ango definition spec
The `.ango` definition file specifies the protocol name, and several services.

Comments can be placed on any line and are started with `//`. Everything until newline (`\n`) is ignored.

`// this is a comment`

Name, first non-comment/non-empty line:

`name <serviceName>`

Procedure description:

`{'server'|'client'} ['synchronized']? ['oneway']? procedureName '(' argument [',' argument]* ')' [ '('result [',' result]* ')' ]`

 - `server`/`client` indicates which party provides the procedure.
 - `synchronized` (idea) see Idea's section.
 - `oneway` (idea) indicates a fire-and-forget procedure. The caller returns imediatly once the call has been sent over the websocket. There's no result expected back. Any possible error should be handled server-side only. Cannot be combined with the `synchronized` keyword.
 - `args` is a list of argument names and their type.
 - `rets` is a list of return values and their type. `rets` is not available for oneway procedures.

Argument/result:
`name type`

### Types
The following types are available. Numeric types in javascript are checked to be within limits, before sent over websocket.
```
.ango    go       angular   description
string   string   string    A string value is a (possibly empty) sequence of bytes.

uint8    uint8    number    The set of all unsigned  8-bit integers (0 to 255)
uint16   uint16   number    The set of all unsigned 16-bit integers (0 to 65535)
uint32   uint32   number    The set of all unsigned 32-bit integers (0 to 4294967295)
uint64   uint64   number    The set of all unsigned 64-bit integers (0 to 18446744073709551615)

int8     int8     number    The set of all signed  8-bit integers (-128 to 127)
int16    int16    number    The set of all signed 16-bit integers (-32768 to 32767)
int32    int32    number    The set of all signed 32-bit integers (-2147483648 to 2147483647)
int64    int64    number    The set of all signed 64-bit integers (-9223372036854775808 to 9223372036854775807)

float32  float32  number    The set of all IEEE-754 32-bit floating-point numbers
float64  float64  number    The set of all IEEE-754 64-bit floating-point numbers

IDEA: provide bytes as base64 string. `type GongBytes []byte` with MarshalJSON.. etc. functions..
bytes    []byte   string-base64? array with chars?
If base64 encoding/decoding is required: http://play.golang.org/p/2FJllMmHk1
```

### Idea's

 - keyword 'synchronized', which locks the websocket both ways, before starting the remote procedure. The websocket is locked both ways until the result has been sent back to the caller. Dangerous because this can seriously block communication for some time. Also, a preceding async that takes some time will have to be cleared before the synchronized procedure can continue. No idea if this will actually work. Incomming/outgoing messages might have to be buffered.. You probably don't want that..

 - versioning. A version string (sha256) is created by combining the .ango outline and the ango version used to generate the code. If version match fails at start of websocket, it's closed. Rather have a strict error and no communication, then working protocol that could yield unexpected results. Imagine if someone changed euro-cents to euro's, but only client-side...

 - objects as arg/ret (presented as object in ng/js and as interface{} in go)

 or

 - typed object, definition in-file, must and types generated by ango
