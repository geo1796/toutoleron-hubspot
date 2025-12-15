package crm

type Object interface {
	GetInternalName() string
	GetObjectTypeID() string
	GetID() string
	GetProperties() map[string]*string
	GetProperty(key string) string
	GetAssociations(internalName string) []AssociationData
}

type BaseObject struct {
	InternalName string
	ObjectTypeID string
	ID           string                        `json:"id"`
	Properties   map[string]*string            `json:"properties"`
	Associations map[string]ObjectAssociations `json:"associations"`
}

type ObjectAssociations struct {
	Results []AssociationData `json:"results"`
}

type AssociationData struct {
	ID   string `json:"id"`
	Type string `json:"type"`
}

func (o *BaseObject) GetInternalName() string {
	return o.InternalName
}

func (o *BaseObject) GetObjectTypeID() string {
	return o.ObjectTypeID
}

func (o *BaseObject) GetID() string {
	return o.ID
}

func (o *BaseObject) GetProperties() map[string]*string {
	return o.Properties
}

func (o *BaseObject) GetAssociations(internalName string) []AssociationData {
	return o.Associations["p"+cfg.accountID+"_"+internalName].Results
}

func (o *BaseObject) GetProperty(key string) string {
	value := o.Properties[key]

	if value == nil {
		return ""
	}

	return *value
}
