package typesense

import (
	"encoding/json"
	"fmt"
)

var (
	hasSearchCache        bool = true
	searchCacheTTLSeconds int  = 60
)

// SetSearchCache : sets search cache when using the search client
func SetSearchCache(isSearchCache bool) {
	hasSearchCache = isSearchCache
}

// SetSearchCache : sets search cache when using the search client
func SetSearchCacheTTL(searchCacheTTL int) {
	searchCacheTTLSeconds = searchCacheTTL
}

// SearchParameters : Search parameters for the serarcvh client
type SearchParameters struct {
	SearchTerm string `json:"q"`
	QueryBy    string `json:"query_by"`
	FilterBy   string `json:"filter_by"`
	SortBy     string `json:"sort_by"`
	Page       string `json:"page"`
	PerPage    string `json:"per_page"`
}

// SearchGroupedParameters : Search Parametes with grouping added
type SearchGroupedParameters struct {
	SearchParameters
	GroupBy    string `json:"group_by"`
	GroupLimit string `json:"group_limit"`
}

func NewSearchParams() SearchParameters {
	return (SearchParameters{SearchTerm: "*"})
}

func NewSearchGroupedParams() SearchGroupedParameters {
	baseSearch := NewSearchParams()
	baseSearch.AddSearchTerm("*")
	return (SearchGroupedParameters{SearchParameters: baseSearch})
}
func (s *SearchParameters) AddSearchTerm(q string) *SearchParameters {
	s.SearchTerm = q
	return s
}
func (s *SearchParameters) AddQueryBy(fieldQueryBy string) *SearchParameters {
	s.QueryBy = fieldQueryBy
	return s
}
func (s *SearchParameters) AddPage(page int) *SearchParameters {
	s.Page = fmt.Sprintf("%d", page)
	return s
}
func (s *SearchParameters) AddPerPage(perPage int) *SearchParameters {
	s.PerPage = fmt.Sprintf("%d", perPage)
	return s
}
func (s *SearchParameters) AddFilterBy(fieldFilterBy string) *SearchParameters {
	s.FilterBy = fieldFilterBy
	return s
}

func (s *SearchParameters) AddSortBy(SortBy string) *SearchParameters {
	s.SortBy = SortBy
	return s
}

func (s *SearchGroupedParameters) AddGroupBy(GroupBy string) *SearchGroupedParameters {
	s.GroupBy = GroupBy
	return s
}

func (s *SearchGroupedParameters) AddGroupLimit(GroupLimit int) *SearchGroupedParameters {
	s.GroupLimit = fmt.Sprintf("%d", GroupLimit)
	return s
}

// ISearchClient : search client interface
type ISearchClient[T any] interface {
	// Search : search without grouping and allows for pagination
	Search(s *SearchParameters) (SearchResult[T], error)
	// SearchGrouped : search with grouping by field and allows for pagniantion
	SearchGrouped(s *SearchGroupedParameters) (SearchResultGrouped[T], error)
	// WithCollectionName : Override the collection name for local operations and not globally
	WithCollectionName(colName string) ISearchClient[T]
	// WithoutDocAutoAlias : if you used the migration tool , it probably auto aliased your collection . if you're doing your own migration
	//                       then call this method to not call the alias route to resolve the doc
	WithoutAutoAlias() ISearchClient[T]
}

// SearchClient : a client that allows you to do advanced
//queries for a specified model
type SearchClient[T any] struct {
	*baseClient[T]
}

func (s *SearchClient[T]) searchRestAny(queryParams any, castValue interface{}) error {
	b, _ := json.Marshal(queryParams)
	var params map[string]string
	_ = json.Unmarshal(b, &params)
	s.Req().
		SetQueryParams(params).
		SetResult(castValue).
		Get(fmt.Sprintf("/collections/%s/documents/search", s.resolveColName()))
	return nil

}

// Search : search without grouping and allows for pagination
func (s *SearchClient[T]) Search(search *SearchParameters) (SearchResult[T], error) {
	var res SearchResult[T]
	err := s.searchRestAny(search, &res)
	return res, err
}

// SearchGrouped : search with grouping by field and allows for pagniantion
func (s *SearchClient[T]) SearchGrouped(search *SearchGroupedParameters) (SearchResultGrouped[T], error) {
	var res SearchResultGrouped[T]
	err := s.searchRestAny(search, &res)
	return res, err
}

// WithCollectionName : Override the collection name for local operations and not globally
func (s *SearchClient[T]) WithCollectionName(colName string) ISearchClient[T] {
	var newSearch SearchClient[T]
	newSearch = *s
	newSearch.colName = colName
	return &newSearch
}

// WithoutDocAutoAlias : if you used the migration tool , it probably auto aliased your collection . if you're doing your own migration
//                       then call this method to not call the alias route to resolve the doc
func (s *SearchClient[T]) WithoutAutoAlias() ISearchClient[T] {
	var newSearch SearchClient[T]
	newSearch = *s
	newSearch.isNotAliased = true
	return &newSearch
}
