package typesense

import (
	"fmt"
	"net/http"

	"github.com/baderkha/library/pkg/conditional"
	http2 "github.com/baderkha/library/pkg/http"
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
	// WithDocBatchSize : Override The batch Size for a local operation and not globally
	WithDocBatchSize(batchSize int64) IDocumentClient[T]
	// WithDocDirtyStrat : Override the dirty document strategy for a local operation and not globally
	WithDocDirtyStrat(dirtyStrat string) IDocumentClient[T]
	// WithCollectionName : Override the collection name for local operations and not globally
	WithDocCollectionName(colName string) IDocumentClient[T]
	// WithoutDocAutoAlias : if you used the migration tool , it probably auto aliased your collection . if you're doing your own migration
	//                       then call this method to not call the alias route to resolve the doc
	WithoutDocAutoAlias() IDocumentClient[T]
}

// DocumentClient : Document client (meant for simple gets , post , patch , deletes
type DocumentClient[T any] struct {
	*baseClient[T]
	colName      string
	batchSize    int64
	dirtyStrat   string
	isNotAliased bool
}

func (d *DocumentClient[T]) getColName() string {
	colName := conditional.Ternary(d.colName != "", d.colName, d.getCollectionName())
	if !d.isNotAliased {
		_, al := d.GetAliasCached(colName)
		colName = al.CollectionName
	}
	return colName
}

func (d *DocumentClient[T]) getBatchSize() string {
	return fmt.Sprintf("%d", conditional.Ternary(d.batchSize != 0, d.batchSize, defaultDocumentOperationBatchSize))
}

func (d *DocumentClient[T]) getDirtyStrat() string {
	return conditional.Ternary(d.dirtyStrat != "", d.dirtyStrat, defaultDocumentDirtyStrat)
}

func (d *DocumentClient[T]) docRoute(subRoute string) string {
	subRoute = conditional.Ternary(subRoute == "", "", "/"+subRoute)
	return fmt.Sprintf("/collections/%s/documents%s", d.getColName(), subRoute)
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
func (d *DocumentClient[T]) WithDocBatchSize(batchSize int64) IDocumentClient[T] {
	var newDoc DocumentClient[T]
	newDoc = *d
	newDoc.batchSize = batchSize
	return &newDoc
}
func (d *DocumentClient[T]) WithDocDirtyStrat(dirtyStrat string) IDocumentClient[T] {
	var newDoc DocumentClient[T]
	newDoc = *d
	newDoc.dirtyStrat = dirtyStrat
	return &newDoc
}

func (d *DocumentClient[T]) WithDocCollectionName(colName string) IDocumentClient[T] {
	var newDoc DocumentClient[T]
	newDoc = *d
	newDoc.colName = colName
	return &newDoc
}

func (d *DocumentClient[T]) WithoutDocAutoAlias() IDocumentClient[T] {
	var newDoc DocumentClient[T]
	newDoc = *d
	newDoc.isNotAliased = true
	return &newDoc
}
