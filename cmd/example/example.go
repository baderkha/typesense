package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/baderkha/typesense"
	"github.com/baderkha/typesense/types"
)

// UserData : a Typical model you would write for your api
type UserData struct {
	// your id , can be generated by you uuid , or typesense can handle it
	ID        string `json:"id"`
	FirstName string `json:"first_name" tsense_required:"1"`
	// by default all fields are optional unless you specify otherwise
	LastName string `json:"last_name" tsense_required:"true"`
	// faceting
	Email string `json:"email" tsense_required:"true" tsense_facet:"1"`
	// default sorting
	GPA float32 `json:"gpa" tsense_default_sort:"1"`
	// default type any int is int64 , you can always override this
	Visit     int32 `json:"visit" tsense_type:"int32"`
	IsDeleted bool  `json:"is_deleted"  tsense_required:"true"`
	// by default time.Time is not supported since time isn't supported in typesense
	//this overrides this issue by turning your time data into unix epoch
	CreatedAt types.Timestamp `json:"created_at"  tsense_required:"true"`
	UpdatedAt types.Timestamp `json:"updated_at"  tsense_required:"true"`
}

func main() {
	apiKey := os.Getenv("TYPESENSE_API_KEY")
	// http://localhost:8080 (include the port if not a standard port ie 80/443)
	host := os.Getenv("TYPESENSE_URL")
	// this will be slow for production , make sure to turn it off
	// the debugger for resty will print everything out
	logging := os.Getenv("TYPESENSE_LOGGING") == "TRUE"

	// this is the main client that does everything and houses all the sub clients
	client := typesense.NewClient[UserData](apiKey, host, logging)

	// Cluster
	// this method will call the api to check for health status
	isHealthy := client.Cluster().Health()
	if !isHealthy {
		log.Fatal(errors.New("not healthy !!! omg"))
	}

	// Migration (create a collection you can put data into)
	// this method will also create 2 resources
	// first is a collection with a version ie user_data_2022-10-10_<some-uuid>
	// second it will create an alias called user_data -> user_data_2022-10-10_<some-uuid>
	//
	// See the typesense documentation to understand the benefits of aliasing
	err := client.Migration().Auto()
	if err != nil {
		log.Fatal(err)
	}

	// Document
	// This method will creat a record under your collection
	// your document will know exactly what collection to write it to
	err = client.Document().Index(&UserData{
		ID:        "some-uuid-for-this",
		FirstName: "Barry",
		LastName:  "Allen",
		Email:     "barry_allen@someFakeEmail.com",
		GPA:       4.00,
		Visit:     50,
		IsDeleted: false,
		CreatedAt: types.Timestamp(time.Now()),
		UpdatedAt: types.Timestamp(time.Now()),
	})

	if err != nil {
		log.Fatal(err)
	}

	// Document
	// this method will get the record we just created and return back
	// a *User pointer without us having to init a var and pass as reference
	// neet right ?
	doc, err := client.Document().GetById("some-uuid-for-this")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(doc.FirstName, doc.LastName, doc.Email)

	// english : give me everyone with the name of barry by first and last name
	// Give me that data as the first page + 20 per page max
	//
	// this is fluent so you can keep adding args and it will build the struct for you
	searchQuery := typesense.
		NewSearchParams().
		AddSearchTerm("barry").
		AddQueryBy("first_name,last_name").
		AddPage(1).
		AddPerPage(20)

	// Search
	// This method will do a filter search / fuzzy search
	// res is of type typesense.SearchResult[UserData]
	res, err := client.Search().Search(searchQuery)
	if err != nil {
		log.Fatal(err)
	}

	userHits := res.Hits

	for _, userHit := range userHits {
		userData := userHit.Document
		fmt.Println(userData.FirstName, userData.LastName)
	}

	fmt.Println(res.Found) // how many found

}
