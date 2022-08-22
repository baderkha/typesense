package typesense

// IClient : General Client that contains all operations supported by typesense
type IClient[T any] interface {
	// Migration : returns back migration client
	Migration() IMigration[T]
	// Document : returns back Document client
	Document() IDocumentClient[T]
	// Search : returns back Search client
	Search() SearchClient[T]
}

// Client : General Client that contains all operations supported by typesense
//
//
type Client[T any] struct {
}
