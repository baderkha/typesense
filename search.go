package typesense

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
	Page       int    `json:"page"`
	PerPage    int    `json:"per_page"`
}

type SearchGroupedParameters struct {
	SearchParameters
	GroupBy    string `json:"group_by"`
	GroupLimit int    `json:"group_limit"`
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
	s.Page = page
	return s
}
func (s *SearchParameters) AddPerPage(perPage int) *SearchParameters {
	s.PerPage = perPage
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
	s.GroupLimit = GroupLimit
	return s
}

// ISearchClient : search client interface
type ISearchClient[T any] interface {
	// Search : search without grouping and allows for pagination
	Search(s *SearchParameters) (SearchResult[T], error)
	// SearchGrouped : search with grouping by field and allows for pagniantion
	SearchGrouped(s *SearchGroupedParameters) (SearchResultGrouped[T], error)
}

// SearchClient : a client that allows you to do advanced
//queries for a specified model
type SearchClient[T any] struct {
	*baseClient[T]
}

// Search : search without grouping and allows for pagination
func (s *SearchClient[T]) Search(search *SearchParameters) (SearchResult[T], error) {
	return SearchResult[T]{}, nil
}

// SearchGrouped : search with grouping by field and allows for pagniantion
func (s *SearchClient[T]) SearchGrouped(search *SearchGroupedParameters) (SearchResultGrouped[T], error) {

}
