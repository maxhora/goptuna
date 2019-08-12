package rdb

import (
	"time"

	"github.com/google/uuid"
	"github.com/jinzhu/gorm"

	"github.com/c-bata/goptuna"
)

var _ goptuna.Storage = &Storage{}

// NewStorage returns new RDB storage.
func NewStorage(db *gorm.DB) *Storage {
	return &Storage{
		db: db,
	}
}

// Storage stores data in your relational databases.
type Storage struct {
	db *gorm.DB
}

// CreateNewStudyID creates study and returns studyID.
func (s *Storage) CreateNewStudyID(name string) (int, error) {
	if name == "" {
		u, err := uuid.NewUUID()
		if err != nil {
			return -1, err
		}
		name = goptuna.DefaultStudyNamePrefix + u.String()
	}
	study := &studyModel{
		Name:      name,
		Direction: directionNotSet,
	}
	result := s.db.Create(study)
	if result.Error != nil {
		return -1, result.Error
	}
	return study.ID, nil
}

// SetStudyDirection sets study direction of the objective.
func (s *Storage) SetStudyDirection(studyID int, direction goptuna.StudyDirection) error {
	d := directionMinimize
	if direction == goptuna.StudyDirectionMaximize {
		d = directionMaximize
	}

	result := s.db.Model(&studyModel{}).Where("study_id = ?", studyID).Update("direction", d)
	if result.Error != nil {
		return result.Error
	}

	return nil
}

// SetStudyUserAttr to store the value for the user.
func (s *Storage) SetStudyUserAttr(studyID int, key string, value string) error {
	result := s.db.Create(&studyUserAttributeModel{
		UserAttributeReferStudy: studyID,
		Key:                     key,
		ValueJSON:               value,
	})
	return result.Error
}

// SetStudySystemAttr to store the value for the system.
func (s *Storage) SetStudySystemAttr(studyID int, key string, value string) error {
	result := s.db.Create(&studySystemAttributeModel{
		SystemAttributeReferStudy: studyID,
		Key:                       key,
		ValueJSON:                 value,
	})
	return result.Error
}

// GetStudyIDFromName return the study id from study name.
func (s *Storage) GetStudyIDFromName(name string) (int, error) {
	panic("implement me")
}

// GetStudyIDFromTrialID return the study id from trial id.
func (s *Storage) GetStudyIDFromTrialID(trialID int) (int, error) {
	panic("implement me")
}

// GetStudyNameFromID return the study name from study id.
func (s *Storage) GetStudyNameFromID(studyID int) (string, error) {
	var study studyModel
	s.db.First(&study, "study_id = ?", studyID)
	return study.Name, nil
}

// GetStudyUserAttrs to restore the attributes for the user.
func (s *Storage) GetStudyUserAttrs(studyID int) (map[string]string, error) {
	var attrs []studyUserAttributeModel
	result := s.db.Find(&attrs, "study_id = ?", studyID)
	if result.Error != nil {
		return nil, result.Error
	}

	res := make(map[string]string, len(attrs))
	for i := range attrs {
		res[attrs[i].Key] = attrs[i].ValueJSON
	}
	return res, nil
}

// GetStudySystemAttrs to restore the attributes for the system.
func (s *Storage) GetStudySystemAttrs(studyID int) (map[string]string, error) {
	var attrs []studySystemAttributeModel
	result := s.db.Find(&attrs, "study_id = ?", studyID)
	if result.Error != nil {
		return nil, result.Error
	}

	res := make(map[string]string, len(attrs))
	for i := range attrs {
		res[attrs[i].Key] = attrs[i].ValueJSON
	}
	return res, nil
}

// GetAllStudySummaries returns all study summaries.
func (s *Storage) GetAllStudySummaries(studyID int) ([]goptuna.StudySummary, error) {
	panic("implement me")
}

// CreateNewTrialID creates trial and returns trialID.
func (s *Storage) CreateNewTrialID(studyID int) (int, error) {
	start := time.Now()
	trial := &trialModel{
		TrialReferStudy: studyID,
		State:           trialStateRunning,
		DatetimeStart:   &start,
	}
	result := s.db.Create(trial)
	if result.Error != nil {
		return -1, result.Error
	}
	_, err := s.createNewTrialNumber(studyID, trial.ID)
	if err != nil {
		return -1, err
	}
	return trial.ID, nil
}

func (s *Storage) createNewTrialNumber(studyID int, trialID int) (int, error) {
	var number int
	err := s.db.Model(&trialSystemAttributeModel{}).Where("trial_id = ?", trialID).Count(&number).Error
	if err != nil {
		return -1, err
	}
	err = s.SetTrialSystemAttr(trialID, "_number", string(number))
	return number, err
}

// SetTrialValue sets the value of trial.
func (s *Storage) SetTrialValue(trialID int, value float64) error {
	panic("implement me")
}

// SetTrialIntermediateValue sets the intermediate value of trial.
func (s *Storage) SetTrialIntermediateValue(trialID int, step int, value float64) error {
	panic("implement me")
}

// SetTrialParam sets the sampled parameters of trial.
func (s *Storage) SetTrialParam(
	trialID int,
	paramName string,
	paramValueInternal float64,
	distribution goptuna.Distribution,
) error {
	j, err := goptuna.DistributionToJSON(distribution)
	if err != nil {
		return err
	}
	err = s.db.Create(&trialParamModel{
		TrialParamReferTrial: trialID,
		Name:                 paramName,
		Value:                paramValueInternal,
		DistributionJSON:     string(j),
	}).Error
	if err != nil {
		return err
	}
	return nil
}

// SetTrialState sets the state of trial.
func (s *Storage) SetTrialState(trialID int, state goptuna.TrialState) error {
	panic("implement me")
}

// SetTrialUserAttr to store the value for the user.
func (s *Storage) SetTrialUserAttr(trialID int, key string, value string) error {
	result := s.db.Create(&trialUserAttributeModel{
		UserAttributeReferTrial: trialID,
		Key:                     key,
		ValueJSON:               value,
	})
	return result.Error
}

// SetTrialSystemAttr to store the value for the system.
func (s *Storage) SetTrialSystemAttr(trialID int, key string, value string) error {
	result := s.db.Create(&trialSystemAttributeModel{
		SystemAttributeReferTrial: trialID,
		Key:                       key,
		ValueJSON:                 value,
	})
	return result.Error
}

// GetTrialNumberFromID returns the trial's number.
func (s *Storage) GetTrialNumberFromID(trialID int) (int, error) {
	panic("implement me")
}

// GetTrialParam returns the internal parameter of the trial
func (s *Storage) GetTrialParam(trialID int, paramName string) (float64, error) {
	panic("implement me")
}

// GetTrialParams returns the external parameters in the trial
func (s *Storage) GetTrialParams(trialID int) (map[string]interface{}, error) {
	panic("implement me")
}

// GetTrialUserAttrs to restore the attributes for the user.
func (s *Storage) GetTrialUserAttrs(trialID int) (map[string]string, error) {
	var attrs []trialUserAttributeModel
	result := s.db.Find(&attrs, "trial_id = ?", trialID)
	if result.Error != nil {
		return nil, result.Error
	}

	res := make(map[string]string, len(attrs))
	for i := range attrs {
		res[attrs[i].Key] = attrs[i].ValueJSON
	}
	return res, nil
}

// GetTrialSystemAttrs to restore the attributes for the system.
func (s *Storage) GetTrialSystemAttrs(trialID int) (map[string]string, error) {
	var attrs []trialSystemAttributeModel
	result := s.db.Find(&attrs, "trial_id = ?", trialID)
	if result.Error != nil {
		return nil, result.Error
	}

	res := make(map[string]string, len(attrs))
	for i := range attrs {
		res[attrs[i].Key] = attrs[i].ValueJSON
	}
	return res, nil
}

// GetBestTrial returns the best trial.
func (s *Storage) GetBestTrial(studyID int) (goptuna.FrozenTrial, error) {
	panic("implement me")
}

// GetAllTrials returns the all trials.
func (s *Storage) GetAllTrials(studyID int) ([]goptuna.FrozenTrial, error) {
	var trials []trialModel

	err := s.db.
		Where("study_id = ?", studyID).
		Preload("UserAttributes").
		Preload("SystemAttributes").
		Preload("TrialParams").
		Preload("TrialValues").
		Find(&trials).Error
	if err != nil {
		return nil, err
	}

	res := make([]goptuna.FrozenTrial, len(trials))
	for i := range trials {
		ft, err := toFrozenTrial(trials[i])
		if err != nil {
			return nil, err
		}
		res[i] = ft
	}
	return res, nil
}

// GetStudyDirection returns study direction of the objective.
func (s *Storage) GetStudyDirection(studyID int) (goptuna.StudyDirection, error) {
	var study studyModel
	err := s.db.First(&study, "study_id = ?", studyID).Error
	if err != nil {
		return goptuna.StudyDirectionMinimize, err
	}
	return toGoptunaStudyDirection(study.Direction), nil
}

// GetTrial returns Trial.
func (s *Storage) GetTrial(trialID int) (goptuna.FrozenTrial, error) {
	var trial trialModel
	var systemAttrs []trialSystemAttributeModel
	var userAttrs []trialUserAttributeModel
	s.db.Model(&trial).Related(&userAttrs).Related(&systemAttrs)
	panic("implement me")
}
