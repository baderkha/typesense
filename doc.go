// Package typense : This package contains a slew of clients that implement the typesense restful interface
//
// Each feature Section under the documentation is abstracted to its own client
//
// IE There is a :
//
// - Migration Client => manages collections / aliases
//
// - Search Client    => Allows advanced search
//
// - Document Client  => Allows indexing (inserting / del / update) Documents
//
// - Cluster Client   => Manages cluster / gets health and other metrics
//
// - Main Client      => A facade for all the clients the fat client that has everything if you're lazy like me
//
// Additionally there are an interfaces for each client as well as a `mock` implementations of the interfaces if you need
// it in a test setting  (built using testify mock package) . However , You are responsible for breaking changes in your testing setup.
//
// Logging is also supported (it will log the outgoing http requests and http responses from typesense)
//
// Final Note :
//
// Create/Update/Delete Operations do not return anything except for errors if there are any . This was a concious design decision .
// Given that Typesense Returns correct status codes ie no need to read the json body data.
// If you disagree feel free to add your implementation and expand the interface
//
//
// See :
//
// https://github.com/baderkha
//
// https://github.com/baderkha/library/blob/main/pkg/store/client/typesense
//
// Author : Ahmad Baderkhan
package typesense
