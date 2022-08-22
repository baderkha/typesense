package types

import (
	"encoding/json"
	"fmt"
	"time"
)

const (
	jsFormat = "2006-01-02T15:04:05.999Z07:00"
)

type Timestamp time.Time

// UnmarshalJSON decodes an int64 timestamp into a time.Time object
func (p *Timestamp) UnmarshalJSON(bytes []byte) error {

	// 1. Decode the bytes into an int64
	var raw int64
	err := json.Unmarshal(bytes, &raw)

	if err != nil {

		fmt.Printf("error decoding timestamp: %s\n", err)
		return err
	}
	t := time.Unix(raw, 0)
	// 2 - Parse the unix timestamp
	*p = Timestamp(t)
	return nil

}

func (p *Timestamp) MarshalJSON() ([]byte, error) {
	if p == nil {
		return nil, nil
	}
	// 1. Decode the bytes into an int64
	x := time.Time(*p)
	return json.Marshal(x.Unix())
}
