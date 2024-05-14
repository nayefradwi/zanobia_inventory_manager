package retailer

import "strconv"

type Retailer struct {
	Id       *int              `json:"id,omitempty"`
	Name     string            `json:"name"`
	Lat      float64           `json:"lat"`
	Lng      float64           `json:"lng"`
	Contacts []RetailerContact `json:"contacts,omitempty"`
}

type RetailerContact struct {
	Id       *int   `json:"id,omitempty"`
	Name     string `json:"name"`
	Position string `json:"position"`
	Email    string `json:"email,omitempty"`
	Phone    string `json:"phone"`
	Website  string `json:"website,omitempty"`
}

func (r Retailer) GetCursorValue() []string {
	return []string{strconv.Itoa(*r.Id)}
}
