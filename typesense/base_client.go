package typesense

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/baderkha/library/pkg/conditional"
	"github.com/baderkha/library/pkg/reflection"
	"github.com/baderkha/library/pkg/stringutil"
	"github.com/go-resty/resty/v2"
	"github.com/lithammer/shortuuid/v4"
	"github.com/tkrajina/go-reflector/reflector"
	"github.com/wlredeye/jsonlines"
)

type baseClient[T any] struct {
	r            *resty.Client
	aliasCache   map[string]string
	mu           sync.Mutex
	isNotAliased bool
	colName      string
}

// getCollectionName : gets an underscore name from the struct field
func (m *baseClient[T]) getCollectionName() string {
	var mdl T
	name, _ := reflection.GetTypeName(mdl)

	return stringutil.Underscore(name)

}

func (m *baseClient[T]) Req() *resty.Request {
	return m.r.R()
}

// VersionCollectionName : adds a version to the collectioName
func (m *baseClient[T]) VersionCollectionName(colName string) string {
	golangDateTime := time.Now().Format("2006-01-02")
	hash := shortuuid.New()
	return fmt.Sprintf("%s_%s_%s", colName, golangDateTime, hash)
}

func (d *baseClient[T]) resolveColName() string {
	colName := conditional.Ternary(d.colName != "", d.colName, d.getCollectionName())
	if !d.isNotAliased {
		_, al := d.GetAliasCached(colName)
		colName = al.CollectionName
	}
	return colName
}

// GetCollectionFromAlias : get underlying collection for an alias name if the binding exists
func (m *baseClient[T]) GetCollectionFromAlias(aliasName string) (doesExist bool, col Collection) {
	doesExist, aliasDetails := m.GetAlias(aliasName)
	if !doesExist {
		return false, col
	}
	return m.GetCollection(aliasDetails.CollectionName)
}

func (m *baseClient[T]) JSONLToByteSlice(jsonL io.Reader) ([]byte, error) {
	var b bytes.Buffer
	_, err := b.ReadFrom(jsonL)
	if err != nil {
		return nil, err
	}
	return b.Bytes(), nil
}

func (m *baseClient[T]) ModelToJSONLines(mod []*T) []byte {
	var buf bytes.Buffer
	_ = jsonlines.Encode(&buf, &mod)
	return buf.Bytes()
}

// GetAlias : gets an alias label and returns back collection name
func (m *baseClient[T]) GetAlias(aliasName string) (doesExist bool, alias Alias) {
	res, err := m.Req().
		SetResult(&alias).
		Get(fmt.Sprintf("/alias/%s", aliasName))
	return err == nil && res.StatusCode() == http.StatusOK, alias
}

// GetAlias : gets an alias label and returns back collection name
func (m *baseClient[T]) GetAliasCached(aliasName string) (doesExist bool, alias Alias) {
	colName := m.aliasCache[aliasName]
	if colName != "" {
		alias.CollectionName = colName
		alias.Name = aliasName
		return true, alias
	}
	exists, alias := m.GetAlias(aliasName)
	if exists {
		m.mu.Lock()
		m.aliasCache[aliasName] = alias.CollectionName
		m.mu.Unlock()
	}
	return exists, alias
}

// GetCollection : gets a collection and checks if it exists
func (m *baseClient[T]) GetCollection(collection string) (doesExist bool, col Collection) {
	res, err := m.Req().
		SetResult(&col).
		Get(fmt.Sprintf("/collections/%s", collection))
	return err == nil && res.StatusCode() == http.StatusOK, col
}

func (m *baseClient[T]) golangToTypesenseType(field reflector.ObjField) (typ string, err error) {
	fieldName, _ := field.Tag("db")
	goType := field.Type().Name()
	if goType == "" {
		goType = field.Type().Elem().Name()
	}
	switch goType {
	case "Time":
		return "int64", nil
	case "bool":
		return "bool", nil
	case "string":
		return "string", nil
	case "[]string":
		return "string[]", nil
	case "int64":
		return "int64", nil
	case "float32":
		return "float", nil
	case "float64":
		return "float", nil
	case "int32":
		fallthrough
	case "int16":
		fallthrough
	case "int8":
		fallthrough
	case "int":
		return "int64 ", nil
	default:
		return "", fmt.Errorf("Typesense : Unsupported field type %s for %s field", goType, fieldName)
	}
}

func newBaseClient[T any](apiKey string, host string, logging bool) *baseClient[T] {
	return &baseClient[T]{
		r:          newHTTPClient(apiKey, host, logging),
		aliasCache: make(map[string]string),
		mu:         sync.Mutex{},
	}
}
