package crm

const (
	ContactInternalName = "contacts"
	ContactObjectTypeID = "0-1"

	ContactPropertyEmail      = "email"
	ContactPropertyFirstName  = "firstname"
	ContactPropertyLastName   = "lastname"
	ContactPropertyPhone      = "phone"
	ContactPropertyAddress    = "address"
	ContactPropertyCity       = "city"
	ContactPropertyZip        = "zip"
	ContactPropertyJob        = "job"
	ContactPropertyWedaID     = "user_id_new"
	ContactPropertySpeciality = "specialite"
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

func (c *Contact) Phone() string {
	return c.GetProperty(ContactPropertyPhone)
}

func (c *Contact) Address() string {
	return c.GetProperty(ContactPropertyAddress)
}

func (c *Contact) City() string {
	return c.GetProperty(ContactPropertyCity)
}

func (c *Contact) Zip() string {
	return c.GetProperty(ContactPropertyZip)
}

func (c *Contact) Job() string {
	return c.GetProperty(ContactPropertyJob)
}

func (c *Contact) WedaID() string {
	return c.GetProperty(ContactPropertyWedaID)
}

func (c *Contact) Speciality() string {
	return c.GetProperty(ContactPropertySpeciality)
}

func (c *Contact) TrainingAssociations() []AssociationData {
	return c.GetAssociations(TrainingInternalName)
}
