package typesense

import (
	"errors"
	"fmt"

	"github.com/baderkha/library/pkg/conditional"
	"github.com/baderkha/library/pkg/http"
	http2 "github.com/baderkha/library/pkg/http"
)

type ClusterOperationResponse struct {
	Success bool `json:"success"`
}

// IClusterClient : cluster client
type IClusterClient interface {
	// Health : pings the cluster for health information
	Health() bool
	// Stats : shows latency stats
	Stats() (map[string]interface{}, error)
	// Metrics : general metrics about the cluster
	Metrics() (map[string]interface{}, error)
	// ReelectLeader : triggers follower note to init rafting process
	ReelectLeader() (bool, error)
	// ToggleSlowRequest : enable logging of requests that take too long
	ToggleSlowRequest(reqTimeMS int64) (bool, error)
	// CreateSnapShot : createa a snapshot for the client
	CreateSnapShot(path string) error
}

type ClusterClient struct {
	*baseClient[any]
}

// Health : pings the cluster for health information
func (c *ClusterClient) Health() bool {
	res, err := c.Req().Get("/health")
	return err == nil && http.StatusIsSuccess(res.StatusCode())
}

// Stats : shows latency stats
func (c *ClusterClient) Stats() (map[string]interface{}, error) {
	var s map[string]interface{} = make(map[string]interface{})
	res, err := c.
		Req().
		SetResult(&s).
		Get("/stats.json")
	if err != nil {
		return nil, err
	} else if !http2.StatusIsSuccess(res.StatusCode()) {
		return nil, typesenseToError(res.Body(), res.StatusCode())
	}
	return s, nil
}

// Metrics : general metrics about the cluster
func (c *ClusterClient) Metrics() (map[string]interface{}, error) {
	var s map[string]interface{} = make(map[string]interface{})
	res, err := c.
		Req().
		SetResult(&s).
		Get("/stats.json")
	if err != nil {
		return nil, err
	} else if !http2.StatusIsSuccess(res.StatusCode()) {
		return nil, typesenseToError(res.Body(), res.StatusCode())
	}
	return s, nil
}

// ReelectLeader : triggers follower note to init rafting process
func (c *ClusterClient) ReelectLeader() (bool, error) {
	var cor ClusterOperationResponse
	res, err := c.Req().SetBody(&cor).Post("/operations/vote")
	if err != nil {
		return false, err
	} else if !http2.StatusIsSuccess(res.StatusCode()) {
		return false, typesenseToError(res.Body(), res.StatusCode())
	}
	return cor.Success, nil
}

// ToggleSlowRequest : enable logging of requests that take too long
func (c *ClusterClient) ToggleSlowRequest(time int64) (bool, error) {
	var cor ClusterOperationResponse
	var s map[string]interface{} = make(map[string]interface{})
	s["log-slow-requests-time-ms"] = time
	res, err := c.
		Req().
		SetResult(&cor).
		SetBody(&s).
		Post("/config")
	if err != nil {
		return false, err
	} else if !http2.StatusIsSuccess(res.StatusCode()) {
		return false, typesenseToError(res.Body(), res.StatusCode())
	}
	return cor.Success, nil
}

// CreateSnapShot : createa a snapshot for the client
func (c *ClusterClient) CreateSnapShot(path string) error {
	var cor ClusterOperationResponse
	res, err := c.
		Req().
		SetResult(&cor).
		SetQueryParam("snapshot_path", fmt.Sprintf("%s", path)).
		Post("/operations/snapshot")
	if err != nil {
		return err
	} else if !http2.StatusIsSuccess(res.StatusCode()) {
		return typesenseToError(res.Body(), res.StatusCode())
	}
	return conditional.Ternary(cor.Success, nil, errors.New("snap shot could not be created"))
}
