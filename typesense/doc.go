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
// - Document Client  => Allows indexing (inserting / del / update) Documents + simple gets by id export ..etc
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

const (
	// TagSort : attach this to your struct field tsense_sort
	//
	// Example :
	//           // your model
	//			 type Model struct {
	//				Field string `tsense_sort:"1"` // this will tell typesense you want this field sorted
	//			 }
	//
	TagSort = "tsense_sort"
	// TagIndex : attach this to your struct field tsense_index
	//
	// Example :
	//           // your model
	//			 type Model struct {
	//				Field string `tsense_index:"1"` // this will tell typesense you want this field indexed
	//			 }
	//
	TagIndex = "tsense_index"
	// TagRequired : attach this to your struct field tsense_required
	//
	// Example :
	//           // your model
	//			 type Model struct {
	//				Field string `tsense_required:"1"` // this will tell typesense you want this field to be required during creates
	//			 }
	//
	TagRequired = "tsense_required"
	// TagRequired : attach this to your struct field tsense_facet
	//
	// Example :
	//           // your model
	//			 type Model struct {
	//				Field string `tsense_facet:"1"` // this will tell typesense you want this field as a facet
	//			 }
	//
	TagFacet = "tsense_facet"
	// TagTypeOverride : attach this to your struct field tsense_type
	//
	// Example :
	//           // your model
	//			type Model struct {
	//				Field int8 `tsense_type:"int32"` // this will tell typesense you want
	//												 // this field to override the type instead of the auto type (int64)
	//			}
	//
	TagTypeOverride = "tsense_type"
	// TagTypeOverride : attach this to your struct field tsense_default_sort
	//
	// Example :
	//           // your model
	//			type Model struct {
	//				Field string `tsense_default_sort:"1"` // this will tell typesense you want
	//												   // this field to be the default sort field
	//			}
	//
	TagDefaultSort = "tsense_default_sort"
)
