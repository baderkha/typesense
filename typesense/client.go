package typesense

// IClient : General Client that contains all operations supported by typesense
type IClient[T any] interface {
	// Migration : returns back migration client
	Migration() IMigration[T]
	// Document : returns back Document client
	Document() IDocumentClient[T]
	// Search : returns back Search client
	Search() ISearchClient[T]
	// Cluster : return back cluster client
	Cluster() IClusterClient
}

// Client : General Client that contains all operations supported by typesense
//
//
type Client[T any] struct {
	migration IMigration[T]
	doc       IDocumentClient[T]
	search    ISearchClient[T]
	cluster   IClusterClient
}

// Migration : returns back migration client
func (c Client[T]) Migration() IMigration[T] {
	return c.migration
}

// Document : returns back Document client
func (c Client[T]) Document() IDocumentClient[T] {
	return c.doc
}

// Search : returns back Search client
func (c Client[T]) Search() ISearchClient[T] {
	return c.search
}

// Cluster : return back cluster client
func (c Client[T]) Cluster() IClusterClient {
	return c.cluster
}
