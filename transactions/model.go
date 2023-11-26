package transactions

type TransactionReason struct {
	Id          *int   `json:"id,omitempty"`
	Name        string `json:"name,omitempty"`
	Description string `json:"Description,omitempty"`
	IsPositive  bool   `json:"isPositive,omitempty"`
}
