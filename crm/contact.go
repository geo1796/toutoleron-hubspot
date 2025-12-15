package crm

const (
	ContactInternalName = "contacts"
	ContactObjectTypeID = "0-1"

	ContactPropertyEmail     = "email"
	ContactPropertyLastName  = "lastname"
	ContactPropertyFirstName = "firstname"
)

type Contact struct {
	Object
}

func (c *Contact) Email() string {
	return c.GetProperty(ContactPropertyEmail)
}

func (c *Contact) LastName() string {
	return c.GetProperty(ContactPropertyLastName)
}

func (c *Contact) FirstName() string {
	return c.GetProperty(ContactPropertyFirstName)
}

func (c *Contact) TrainingAssociations() []AssociationData {
	return c.GetAssociations(TrainingInternalName)
}
