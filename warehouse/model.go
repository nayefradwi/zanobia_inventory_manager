package warehouse

type Warehouse struct {
	Id   *int    `json:"id,omitempty"`
	Name string  `json:"name"`
	Lat  float64 `json:"lat"`
	Lng  float64 `json:"lng"`
}
