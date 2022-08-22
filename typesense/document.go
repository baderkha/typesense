package typesense

import (
	"fmt"
	"net/http"

	"github.com/baderkha/library/pkg/conditional"
	http2 "github.com/baderkha/library/pkg/http"
	"github.com/pkg/errors"
	"github.com/spf13/afero"
)

const (
	// DocumentActionUpsert : upsert a bunch of documents . if update (document must have all full required fields)
	DocumentActionUpsert = "upsert"
	// DocumentActionUpsert : update a bunch of documents , error if id does not exist
	DocumentActionUpdate = "update"
	// DocumentActionUpsert : create a new document / documents , error if already exists it will merge
	DocumentActionEmplace = "emplace"

	// DocumentDirtyStratCORreject : Attempt coercion of the field's value to previously inferred type.
	// If coercion fails, reject the write outright with an error message.
	DocumentDirtyStratCORreject = "corece_or_reject"
	// DocumentDirtyStratReject : Reject the document outright.
	DocumentDirtyStratReject = "reject"
	// DocumentDirtyStratDrop : Drop the particular field and index the rest of the document.
	DocumentDirtyStratDrop = "drop"
	// DocumentDirtyStratCODrop :Attempt coercion of the field's value to previously inferred type.
	// If coercion fails, drop the particular field and index the rest of the document.
	DocumentDirtyStratCODrop = "corece_or_drop"
)

var (
	defaultDocumentOperationBatchSize int64 = 100
	defaultDocumentDirtyStrat               = "reject"
)

// OverrideDocBatchSize : override the batch size when doing document operations
func OverrideDocBatchSize(docBSize int64) {
	defaultDocumentOperationBatchSize = docBSize
}

// OverrideDocDirtyStrat : Override the dirty value strategy when uploading documents
func OverrideDocDirtyStrat(dirtyStrat string) {
	defaultDocumentDirtyStrat = dirtyStrat
}

// DocumentClient : Document client (meant for simple gets , post , patch , deletes
type IDocumentClient[T any] interface {
	// GetById : get 1 document by id otherwise error out not found
	GetById(id string) (m *T, err error)
	// IsExistById : check if document exists
	IsExistById(id string) bool
	// ExportAll : Export all the document as a byte slice (string represntation of jsonL)
	ExportAll() ([]byte, error) // exports as []byte (string represenation of )
	// ExportAll : Export all the document with a query filter as a byte slice (string represntation of jsonL)
	ExportAllWithQuery(query string) ([]byte, error)
	// Index : create / add new document . can also do upserts
	Index(m *T) error
	// Update : update existing document completley. will error out if required details are missing or if document not found
	Update(m *T, id string) error
	// DeleteById : Delete by an id
	DeleteById(id string) error
	// DeleteManyWithQuery : delete more than 1 with a query criteria
	DeleteManyWithQuery(query string) error
	// IndexMany : create multiple documents (transforms them to jsonL behind the scenes)
	IndexMany(m []*T, action string) error
	// ImportMany : import many documents with json lines
	ImportMany(jsonLines []byte, action string) error
	// ImportManyFromFile : opens a file from path and sends it to your typesense backend
	//
	// This gives you the operatunity to specifc the file system to be used
	// By default this package uses the os file system , but since this client is built
	// with the afero package (see https://github.com/spf13/afero) your file system options are limitless
	//
	// Example :
	//				// your model
	// 				type User struct {
	//					FirstName string `json:"first_name" tsense_sort:"1" tsense_required:"1"`
	//				}
	//				func main() {
	//					filePath := "/tmp/some_file.jsonl"
	//					memoryFS := afero.NewMemMapFs()
	//
	//					// memory file system instead of the os
	//					// (there's a bunch of implementations you can check out)
	//					typesense.OverrideFS(memoryFS)
	//
	//					docClient := typesense.NewDocumentClient[User]("api key" , "host" , false)
	//					err := docClient.ImportManyFromFile(
	//							filePath,
	//							typesense.DocumentActionUpsert,
	//						)
	//					if err != nil {
	//						log.Fatal(err)
	//					}
	//				}
	//
	ImportManyFromFile(path string, action string) error
	// WithBatchSize : Override The batch Size for a local operation and not globally
	WithBatchSize(batchSize int64) IDocumentClient[T]
	// WithDirtyStrat : Override the dirty document strategy for a local operation and not globally
	WithDirtyStrat(dirtyStrat string) IDocumentClient[T]
	// WithCollectionName : Override the collection name for local operations and not globally
	WithCollectionName(colName string) IDocumentClient[T]
	// WithoutAutoAlias : if you used the migration tool , it probably auto aliased your collection . if you're doing your own migration
	//                       then call this method to not call the alias route to resolve the doc
	WithoutAutoAlias() IDocumentClient[T]
}

// NewDocumentClient : create a new document client which allows you to do basic crud operations on documents
func NewDocumentClient[T any](apiKey string, host string, logging bool) IDocumentClient[T] {
	base := newBaseClient[T](apiKey, host, logging)
	return &DocumentClient[T]{
		baseClient: base,
	}
}

// DocumentClient : Document client (meant for simple gets , post , patch , deletes
type DocumentClient[T any] struct {
	*baseClient[T]
	batchSize  int64
	dirtyStrat string
}

func (d *DocumentClient[T]) getBatchSize() string {
	return fmt.Sprintf("%d", conditional.Ternary(d.batchSize != 0, d.batchSize, defaultDocumentOperationBatchSize))
}

func (d *DocumentClient[T]) getDirtyStrat() string {
	return conditional.Ternary(d.dirtyStrat != "", d.dirtyStrat, defaultDocumentDirtyStrat)
}

func (d *DocumentClient[T]) docRoute(subRoute string) string {
	subRoute = conditional.Ternary(subRoute == "", "", "/"+subRoute)
	return fmt.Sprintf("/collections/%s/documents%s", d.resolveColName(), subRoute)
}

// IsExistById : check if document exists
func (d *DocumentClient[T]) IsExistById(id string) bool {
	res, err := d.
		Req().
		Get(d.docRoute(id))
	return err == nil && res.StatusCode() != http.StatusNotFound
}

func (d *DocumentClient[T]) GetById(id string) (*T, error) {
	var m T
	res, err := d.
		Req().
		SetResult(&m).
		Get(d.docRoute(id))

	if err != nil {
		return nil, err
	} else if !http2.StatusIsSuccess(res.StatusCode()) {
		return nil, typesenseToError(res.Body(), res.StatusCode())
	}

	return &m, nil
}
func (d *DocumentClient[T]) ExportAll() ([]byte, error) {
	return d.ExportAllWithQuery("")
}
func (d *DocumentClient[T]) ExportAllWithQuery(query string) ([]byte, error) {

	res, err := d.
		Req().
		SetQueryParam("filter_by", query).
		Get(d.docRoute("export"))

	if err != nil {
		return nil, err
	} else if !http2.StatusIsSuccess(res.StatusCode()) {
		return nil, typesenseToError(res.Body(), res.StatusCode())
	}

	return res.Body(), nil

}
func (d *DocumentClient[T]) Index(m *T) error {
	res, err := d.
		Req().
		SetBody(m).
		SetQueryParam("dirty_values", d.getDirtyStrat()).
		Post(d.docRoute(""))

	if err != nil {
		return err
	} else if !http2.StatusIsSuccess(res.StatusCode()) {
		return typesenseToError(res.Body(), res.StatusCode())
	}

	return nil
}
func (d *DocumentClient[T]) Update(m *T, id string) error {
	res, err := d.
		Req().
		SetBody(m).
		SetQueryParam("dirty_values", d.getDirtyStrat()).
		Put(d.docRoute(id))

	if err != nil {
		return err
	} else if !http2.StatusIsSuccess(res.StatusCode()) {
		return typesenseToError(res.Body(), res.StatusCode())
	}

	return nil
}

func (d *DocumentClient[T]) DeleteById(id string) error {
	res, err := d.
		Req().
		Delete(d.docRoute(id))

	if err != nil {
		return err
	} else if !http2.StatusIsSuccess(res.StatusCode()) {
		return typesenseToError(res.Body(), res.StatusCode())
	}

	return nil

}
func (d *DocumentClient[T]) DeleteManyWithQuery(query string) error {
	res, err := d.
		Req().
		SetQueryParam("filter_by", query).
		SetQueryParam("batch_size", d.getBatchSize()).
		Delete(d.docRoute(""))

	if err != nil {
		return err
	} else if !http2.StatusIsSuccess(res.StatusCode()) {
		return typesenseToError(res.Body(), res.StatusCode())
	}

	return nil
}
func (d *DocumentClient[T]) IndexMany(m []*T, action string) error {
	return d.ImportMany(d.ModelToJSONLines(m), action)
}

func (d *DocumentClient[T]) ImportManyFromFile(path string, action string) error {
	content, err := afero.ReadFile(fs, path)
	if err != nil {
		errors.Wrap(err, typesenseErrPrefix)
	}
	return d.ImportMany(content, action)
}
func (d *DocumentClient[T]) ImportMany(jsonLines []byte, action string) error {
	res, err := d.
		Req().
		SetHeader("Content-Type", "text/plain").
		SetBody(jsonLines).
		SetQueryParam("action", action).
		SetQueryParam("dirty_values", d.getDirtyStrat()).
		SetQueryParam("batch_size", d.getBatchSize()).
		Post(d.docRoute("import"))

	if err != nil {
		return err
	} else if !http2.StatusIsSuccess(res.StatusCode()) {
		return typesenseToError(res.Body(), res.StatusCode())
	}

	return nil
}
func (d *DocumentClient[T]) WithBatchSize(batchSize int64) IDocumentClient[T] {
	var newDoc DocumentClient[T]
	newDoc = *d
	newDoc.batchSize = batchSize
	return &newDoc
}
func (d *DocumentClient[T]) WithDirtyStrat(dirtyStrat string) IDocumentClient[T] {
	var newDoc DocumentClient[T]
	newDoc = *d
	newDoc.dirtyStrat = dirtyStrat
	return &newDoc
}

func (d *DocumentClient[T]) WithCollectionName(colName string) IDocumentClient[T] {
	var newDoc DocumentClient[T]
	newDoc = *d
	newDoc.colName = colName
	return &newDoc
}

func (d *DocumentClient[T]) WithoutAutoAlias() IDocumentClient[T] {
	var newDoc DocumentClient[T]
	newDoc = *d
	newDoc.isNotAliased = true
	return &newDoc
}
