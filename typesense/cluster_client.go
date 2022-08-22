package typesense

import "github.com/go-resty/resty/v2"

type ClusterClient struct {
	httpClient resty.Client
}
