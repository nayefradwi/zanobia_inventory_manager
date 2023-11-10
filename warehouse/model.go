package warehouse

type Warehouse struct {
	Id   *int     `json:"id,omitempty"`
	Name string   `json:"name"`
	Lat  *float64 `json:"lat"`
	Lng  *float64 `json:"lng"`
}

type WarehouseUserInput struct {
	WarehouseId int `json:"warehouseId"`
	UserId      int `json:"userId"`
}
