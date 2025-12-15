package crm

const (
	TrainingInternalName = "trainings"
	TrainingObjectTypeID = "2-141027445"

	TrainingToSessionAssociationCategory = "USER_DEFINED"
	TrainingToSessionAssociationTypeID   = 290

	TrainingToContactAssociationCategory = "USER_DEFINED"
	TrainingToContactAssociationTypeID   = 288

	TrainingPropertyName          = "training_name"
	TrainingPropertyTrainer       = "trainer"
	TrainingPropertyTimeSpent     = "time_spent"
	TrainingPropertyPipelineStage = "hs_pipeline_stage"
)

type Training struct {
	Object
}

func (t *Training) Name() string {
	return t.GetProperty(TrainingPropertyName)
}

func (t *Training) Trainer() string {
	return t.GetProperty(TrainingPropertyTrainer)
}

func (t *Training) PipelineStage() string {
	raw := t.GetProperty(TrainingPropertyPipelineStage)

	switch raw {
	case "2767825099":
		return "Contrat Weda OK"
	case "2031944925":
		return "Contrat Toutoléron ok"
	case "2031944926":
		return "Compte actif"
	case "2031944932":
		return "Formation programmée"
	case "2031944933":
		return "En cours / programmation à faire"
	case "2031944934":
		return "En cours / programmée"
	case "2031944935":
		return "À facturer"
	case "3129890016":
		return "Facturée"
	case "":
		return "Non renseigné"
	default:
		return "Inconnu"
	}
}

func (t *Training) SessionAssociations() []AssociationData {
	return t.GetAssociations(SessionInternalName)
}
