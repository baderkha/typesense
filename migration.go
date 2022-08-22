package typesense

import (
	"fmt"
	"reflect"
	"sort"
	"strings"
	"sync"

	"github.com/baderkha/library/pkg/conditional"
	http2 "github.com/baderkha/library/pkg/http"
	"github.com/baderkha/library/pkg/store/entity"
	"github.com/davecgh/go-spew/spew"
	"github.com/tkrajina/go-reflector/reflector"
)

var _ IMigration[any] = &Migration[any]{}

// Migration : Migration Client for typesense
//
// Houses Both Collection / Alias Operations + Adds a handy AutoMigration function / ManualMigration
type IMigration[T any] interface {
	// AliasCollection : create an alias for a collection
	AliasCollection(a *Alias) error
	// Auto : AutoMigrate Depending on the model it will construct a collection via typesense , alias it and maintain the schema
	// Example:
	//			// Your Model you want to map to the client
	//			type MyCoolModel struct {
	//				Field string `db:"field"`
	//			}
	//
	//			func main (){
	//				migration := typesense.NewModelMigration[MyCoolModel]("<api_key>","<http_server_url>",false)
	//				// your alias will be my_cool_model -> my_cool_model_2022-10-10_<SomeHash>
	//				// where the latter value is the underlying collection name
	//				migration.Auto()
	//			}
	//
	Auto() error
	// DeleteAliasCollection : deletes an alias pointer
	DeleteAliasCollection(aliasName string) error
	// DeleteCollection : deletes collection and collection data
	DeleteCollection(colName string, col *CollectionUpdate) error
	// GetAlias : gets an alias label and returns back collection name
	GetAlias(aliasName string) (doesExist bool, alias Alias)
	// GetCollection : gets a collection and checks if it exists
	GetCollection(collection string) (doesExist bool, col Collection)
	// GetCollectionFromAlias : get underlying collection for an alias name if the binding exists
	GetCollectionFromAlias(aliasName string) (doesExist bool, col Collection)
	// Manual : if you don't trust auto migration , you can always migrate it yourself
	// or build your own auto schema converter yourself .
	//
	// Note that the implementation uses aliasing .
	//
	// Example:
	//			// Your Model you want to map to the client
	//			type MyCoolModel struct {
	//				Field string `db:"field"`
	//			}
	//
	//			func main (){
	//				migration := typesense.NewModelMigration[MyCoolModel]("<api_key>","<http_server_url>",false)
	//
	//				// your alias will be my_cool_model -> my_cool_model_2022-10-10_<SomeHash>
	//				// where the latter value is the underlying collection name
	//				migration.Manual(&typesense.Collection{Name : "my_cool_model"},true)
	//				// no alias created your collectio nname is my_cool_model
	//				migration.Manual(&typesense.Collection{Name : "my_cool_model"},false)
	//			}
	//
	Manual(col *Collection, alias bool) error
	// ModelToCollection : converts a model to a typesense collection , useful for manual migration
	ModelToCollection() (*Collection, error)
	// MustAuto : Must auto migrate basically calls AutoMigrate and panics on failure
	MustAuto()
	// MustAuto : Must auto migrate basically calls AutoMigrate and panics on failure
	MustManual(col *Collection, alias bool)
	// NewCollection : create a new collection
	NewCollection(col *Collection) error
	// UpdateCollection : updates collection schema
	UpdateCollection(colName string, col *CollectionUpdate) error
	// VersionCollectionName : adds a version to the collectioName
	VersionCollectionName(colName string) string
}

// Migration : Migration Client for typesense
//
// Houses Both Collection / Alias Operations + Adds a handy AutoMigration function / ManualMigration
type Migration[T any] struct {
	*baseClient[T]
}

// Auto : AutoMigrate Depending on the model it will construct a collection via typesense , alias it and maintain the schema
// Example:
//			// Your Model you want to map to the client
//			type MyCoolModel struct {
//				Field string `db:"field"`
//			}
//
//			func main (){
//				migration := typesense.NewModelMigration[MyCoolModel]("<api_key>","<http_server_url>",false)
//				// your alias will be my_cool_model -> my_cool_model_2022-10-10_<SomeHash>
//				// where the latter value is the underlying collection name
//				migration.Auto()
//			}
//
func (m Migration[T]) Auto() error {
	colSchema, err := m.ModelToCollection()
	if err != nil {
		return err
	}
	return m.Manual(colSchema, false)
}

func sortFieldsFunc(col *Collection, wg *sync.WaitGroup) {
	defer wg.Done()
	sort.Slice(col.Fields, func(i, j int) bool {
		return strings.ToLower(col.Fields[i].Name) < strings.ToLower(col.Fields[j].Name)
	})
}

// Manual : if you don't trust auto migration , you can always migrate it yourself
// or build your own auto schema converter yourself .
//
// Note that the implementation uses aliasing .
//
// Example:
//			// Your Model you want to map to the client
//			type MyCoolModel struct {
//				Field string `db:"field"`
//			}
//
//			func main (){
//				migration := typesense.NewModelMigration[MyCoolModel]("<api_key>","<http_server_url>",false)
//
//				// your alias will be my_cool_model -> my_cool_model_2022-10-10_<SomeHash>
//				// where the latter value is the underlying collection name
//				migration.Manual(&typesense.Collection{Name : "my_cool_model"},true)
//				// no alias created your collectio nname is my_cool_model
//				migration.Manual(&typesense.Collection{Name : "my_cool_model"},false)
//			}
//
func (m Migration[T]) Manual(col *Collection, alias bool) error {
	var wg sync.WaitGroup
	var typeSenseCollection Collection
	var colExists bool
	aliasName := col.Name
	colCompareCopy := *col
	if alias {
		colExists, typeSenseCollection = m.GetCollectionFromAlias(aliasName)
	} else {
		colExists, typeSenseCollection = m.GetCollection(col.Name)
	}
	spew.Dump(colExists)
	// if exist , we're doing a put
	if colExists {
		colCompareCopy.Name = typeSenseCollection.Name
		// they're both doing the same thing
		// why not have it concurrent
		wg.Add(1)
		go sortFieldsFunc(&typeSenseCollection, &wg)
		wg.Add(1)
		go sortFieldsFunc(&colCompareCopy, &wg)
		wg.Wait()

		// compare changes via diff
		// gurad check if the one in the backend matches the input
		if reflect.DeepEqual(col, typeSenseCollection) {
			return nil
		}

		panic("Typesense Error : Update is : unimplemented yet sorry :(")

	}

	if alias {
		col.Name = m.VersionCollectionName(aliasName)
	}

	// otherwise we're doing a post request
	err := m.NewCollection(col)
	if err != nil {
		return err
	}
	if alias {
		err := m.AliasCollection(&Alias{
			Name:           aliasName,
			CollectionName: col.Name,
		})
		if err != nil {
			return err
		}
	}

	return nil

}

// AliasCollection : create an alias for a collection
func (m Migration[T]) AliasCollection(a *Alias) error {
	res, err := m.Req().
		SetBody(a).
		Put(fmt.Sprintf("/aliases/%s", a.Name))

	if err != nil {
		return err
	} else if !http2.StatusIsSuccess(res.StatusCode()) {
		return typesenseToError(res.Body(), res.StatusCode())
	}
	return nil
}

// DeleteAliasCollection : deletes an alias pointer
func (m Migration[T]) DeleteAliasCollection(aliasName string) error {
	res, err := m.Req().
		Delete(fmt.Sprintf("/aliases/%s", aliasName))

	if err != nil {
		return err
	} else if !http2.StatusIsSuccess(res.StatusCode()) {
		return typesenseToError(res.Body(), res.StatusCode())
	}
	return nil
}

// NewCollection : create a new collection
func (m Migration[T]) NewCollection(col *Collection) error {

	res, err := m.Req().
		SetBody(col).
		Post("/collections")

	if err != nil {
		return err
	} else if !http2.StatusIsSuccess(res.StatusCode()) {
		return typesenseToError(res.Body(), res.StatusCode())
	}
	return nil
}

// UpdateCollection : updates collection schema
func (m Migration[T]) UpdateCollection(colName string, col *CollectionUpdate) error {
	res, err := m.Req().
		SetBody(col).
		Patch(fmt.Sprintf("/collections/%s", colName))

	if err != nil {
		return err
	} else if !http2.StatusIsSuccess(res.StatusCode()) {
		return typesenseToError(res.Body(), res.StatusCode())
	}
	return nil
}

// DeleteCollection : deletes collection and collection data
func (m Migration[T]) DeleteCollection(colName string, col *CollectionUpdate) error {
	res, err := m.Req().
		SetBody(col).
		Delete(fmt.Sprintf("/collections/%s", colName))

	if err != nil {
		return err
	} else if !http2.StatusIsSuccess(res.StatusCode()) {
		return typesenseToError(res.Body(), res.StatusCode())
	}
	return nil
}

// ModelToCollection : converts a model to a typesense collection , useful for manual migration
func (m Migration[T]) ModelToCollection() (*Collection, error) {
	var col Collection
	var defaultSort string
	col.Name = m.getCollectionName()
	s := entity.Account{}

	// Fields will list every structure exportable fields.
	// Here, it's content would be equal to:
	// []string{"FirstField", "SecondField", "ThirdField"}
	obj := reflector.New(s)
	fieldsGoLang := obj.FieldsFlattened()
	for _, field := range fieldsGoLang {
		colVal, _ := field.Tag("db")
		tType, err := m.golangToTypesenseType(field)
		if err != nil {
			return nil, err
		}
		sortVal, _ := field.Tag(TagSort)
		indexVal, _ := field.Tag(TagIndex)
		requiredVal, _ := field.Tag(TagRequired)
		facetVal, _ := field.Tag(TagFacet)
		overrideTypeVal, _ := field.Tag(TagTypeOverride)
		defaultSortVal, _ := field.Tag(TagDefaultSort)

		if defaultSortVal != "" {
			if defaultSort != "" {
				return nil, fmt.Errorf("Typesense : You cannot have more than 1 default sort field")
			}
			defaultSort = colVal
			requiredVal = "1"
			indexVal = "1"
		}

		col.Fields = append(col.Fields, CollectionField{
			Facet:    facetVal != "",
			Index:    indexVal != "",
			Optional: requiredVal == "",
			Sort:     sortVal != "",
			Name:     colVal,
			Type:     conditional.Ternary(overrideTypeVal != "", overrideTypeVal, tType),
		})
	}
	col.DefaultSortingField = defaultSort
	return &col, nil
}

// MustAuto : Must auto migrate basically calls AutoMigrate and panics on failure
func (m Migration[T]) MustAuto() {
	err := m.Auto()
	if err != nil {
		panic(err)
	}
}

// MustManual : Must manually migrate calls Manual and panics on failure
func (m Migration[T]) MustManual(col *Collection, alias bool) {
	err := m.Manual(col, alias)
	if err != nil {
		panic(err)
	}
}
