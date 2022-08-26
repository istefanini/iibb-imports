package models

type Payment struct {
	Key                string  `json:"key"`
	External_reference string  `json:"external_reference"`
	Rate               float64 `json:"rate"`
	Status             string  `json:"status"`
}
