package typesense

// Client : General Client that contains all operations supported by typesense
//
//
type IClient[T any] interface {
	Migration() IMigration[T]
	Document() IDocumentClient[T]
}

// Client : General Client that contains all operations supported by typesense
//
//
type Client[T any] struct {
}
