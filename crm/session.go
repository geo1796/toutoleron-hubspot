package crm

const (
	SessionInternalName = "sessions"
	SessionObjectTypeID = "2-141027484"

	SessionToTrainingAssociationCategory = "USER_DEFINED"
	SessionToTrainingAssociationTypeID   = 291

	SessionPropertyName          = "hour"
	SessionPropertyComment       = "comment"
	SessionPropertyTrainer       = "trainer"
	SessionPropertyStartTime     = "start_time"
	SessionPropertyEndTime       = "end_time"
	SessionPropertyValidated     = "validated"
	SessionPropertyTrainingStage = "training_stage"
)

type Session struct {
	Object
}

func (s *Session) Name() string {
	return s.GetProperty(SessionPropertyName)
}

func (s *Session) StartTime() string {
	return s.GetProperty(SessionPropertyStartTime)
}

func (s *Session) EndTime() string {
	return s.GetProperty(SessionPropertyEndTime)
}
