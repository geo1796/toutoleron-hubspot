package crm

type ObjectOwner struct {
	ID        string
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
}
