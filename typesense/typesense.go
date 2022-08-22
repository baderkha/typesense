package typesense

import (
	"fmt"
	"sync"

	"github.com/go-resty/resty/v2"
	"github.com/spf13/afero"
)

const (
	typesenseErrPrefix = "TypeSenseClient : Bad Response : With Code"
)

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

var (
	fs = afero.NewOsFs()

	httpRetryCount = 1
)

// SetHTTPRetryCount : sets the retry count for the requests
func SetHTTPRetryCount(retryCount int) {
	httpRetryCount = retryCount
}

// OverrideFS : change the file ssytem that will be used in the client
//
// call this before constructing the Client . if you want a different file system
//
// see :
//
// https://github.com/spf13/afero#available-backends
//
// // forexample you can have a s3 bucket to store your jsonl files
//
// https://github.com/fclairamb/afero-s3  // (cool right :))
//
//
//
func OverrideFS(newFs afero.Fs) {
	fs = newFs
}

func typesenseToError(responseBody []byte, statusCode int) error {
	return fmt.Errorf("%s : %d  : %s", typesenseErrPrefix, statusCode, string(responseBody))
}

// CollectionField : field for a typesense collection
type CollectionField struct {
	Facet    bool `json:"facet"`
	Index    bool `json:"index"`
	Optional bool `json:"optional"`
	Sort     bool `json:"sort"`

	Name string `json:"name"`
	Type string `json:"type"`
}

// Collection : typesense collection
type Collection struct {
	Name                string            `json:"name"`
	Fields              []CollectionField `json:"fields"`
	DefaultSortingField string            `json:"default_sorting_field"`
}

// CollectionField : field for a typesense collection
type CollectionFieldUpdate struct {
	Facet    bool   `json:"facet"`
	Index    bool   `json:"index"`
	Optional bool   `json:"optional"`
	Sort     bool   `json:"sort"`
	Drop     bool   `json:"drop"`
	Name     string `json:"name"`
	Type     string `json:"type"`
}

// Collection : typesense collection
type CollectionUpdate struct {
	Fields              []CollectionField `json:"fields"`
	DefaultSortingField string            `json:"default_sorting_field"`
}

// Alias : alias to a collection
type Alias struct {
	Name           string `json:"name "`
	CollectionName string `json:"collection_name"`
}

type SearchResultBase struct {
	FacetCounts   []string      `json:"facet_counts"`
	Found         int           `json:"found"`
	OutOf         int           `json:"out_of"`
	Page          int           `json:"page"`
	RequestParams RequestParams `json:"request_params"`
	SearchTimeMs  int           `json:"search_time_ms"`
}

// SearchResult : search result without grouping
type SearchResult[T any] struct {
	SearchResultBase
	Hits []Hit[T] `json:"hits"`
}

// SearchResultGrouped : search result with grouping
type SearchResultGrouped[T any] struct {
	SearchResultBase
	GroupedHits []GroupedHits[T] `json:"grouped_hits"`
}

// Hits : results , houses your documents
type Hit[T any] struct {
	Document   T          `json:"document"`
	Highlights Highlights `json:"highlights"`
	TextMatch  int        `json:"text_match"`
}

// GroupedHits : results , houses your documents and group by info
type GroupedHits[T any] struct {
	GroupKey []string `json:"group_key"`
	Hits     []Hit[T] `json:"hits"`
}

// RequestParams : metadata describing the request params you used
type RequestParams struct {
	CollectionName string `json:"collection_name"`
	PerPage        int    `json:"per_page"`
	Q              string `json:"q"`
}

// Highlights : more metadata
type Highlights struct {
	Field         string   `json:"field"`
	MatchedTokens []string `json:"matched_tokens"`
	Snippet       string   `json:"snippet"`
}

func newHTTPClient(apiKey, host string, logging bool) *resty.Client {
	return resty.
		New().
		SetHeaders(map[string]string{
			"Content-Type":        "application/json",
			"X-TYPESENSE-API-KEY": apiKey,
		}).
		SetBaseURL(host).
		SetRetryCount(httpRetryCount).
		SetDebug(logging)
}

// NewModelMigration : Migration if you want to use Model Dependent migration ie tie your migration client to a
//A Specific struct declaration
func NewModelMigration[T any](apiKey string, host string, logging bool) IMigration[T] {
	base := newBaseClient[T](apiKey, host, logging)
	return &Migration[T]{
		baseClient: base,
	}
}

// NewDocumentClient : create a new document client which allows you to do basic crud operations on documents
func NewDocumentClient[T any](apiKey string, host string, logging bool) IDocumentClient[T] {
	base := newBaseClient[T](apiKey, host, logging)
	return &DocumentClient[T]{
		baseClient: base,
	}
}

// NewSearchClient : create a new search client which allows you to do advanced search
func NewSearchClient[T any](apiKey string, host string, logging bool) ISearchClient[T] {
	base := newBaseClient[T](apiKey, host, logging)
	return &SearchClient[T]{
		baseClient: base,
	}
}

// NewManualMigration : Migration if you want to do your own thing and use the low level wrapper methods for the rest calls
//
// Disclaimer :
// 				// Do not use Auto Method
// 				migrator := typesense.NewManualMigration("","","")
//				migrator.Auto() // should panic or error or give you a bad status code
//
//				// instead you will have to make the calls yourself
func NewManualMigration(apiKey string, host string, logging bool) IMigration[any] {
	base := newBaseClient[any](apiKey, host, logging)
	return &Migration[any]{
		baseClient: base,
	}
}

// NewClusterClient : client for cluster operations
func NewClusterClient(apiKey string, host string, logging bool) IClusterClient {
	base := newBaseClient[any](apiKey, host, logging)
	return &ClusterClient{
		baseClient: base,
	}
}

// NewClient : default client , has all the other clients wrapped
func NewClient[T any](apiKey string, host string, logging bool) IClient[T] {
	return &Client[T]{
		migration: NewModelMigration[T](apiKey, host, logging),
		doc:       NewDocumentClient[T](apiKey, host, logging),
		search:    NewSearchClient[T](apiKey, host, logging),
		cluster:   NewClusterClient(apiKey, host, logging),
	}
}

// NewClientNoGeneric : default client , has all the other clients wrapped . This will not have any generic bindings
//						usage is limited
func NewClientNoGeneric(apiKey string, host string, logging bool) IClient[any] {
	return &Client[any]{
		migration: NewModelMigration[any](apiKey, host, logging),
		doc:       NewDocumentClient[any](apiKey, host, logging),
		search:    NewSearchClient[any](apiKey, host, logging),
		cluster:   NewClusterClient(apiKey, host, logging),
	}
}

func newBaseClient[T any](apiKey string, host string, logging bool) *baseClient[T] {
	return &baseClient[T]{
		r:          newHTTPClient(apiKey, host, logging),
		aliasCache: make(map[string]string),
		mu:         sync.Mutex{},
	}
}
