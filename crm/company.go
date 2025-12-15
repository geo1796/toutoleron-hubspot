package crm

const (
	CompanyInternalName = "companies"
	CompanyObjectTypeID = "0-2"

	CompanyPropertyName = "name"
)

type Company struct {
	Object
}

func (c *Company) Name() string {
	return c.GetProperty(CompanyPropertyName)
}

func (c *Company) TrainingAssociations() []AssociationData {
	return c.GetAssociations(TrainingInternalName)
}
