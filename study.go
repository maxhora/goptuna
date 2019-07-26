package goptuna

import (
	"context"
	"fmt"
	"sync"

	"go.uber.org/zap"
)

// FuncObjective is a type of objective function
type FuncObjective func(trial Trial) (float64, error)

// StudyDirection represents the direction of the optimization
type StudyDirection string

const (
	// StudyDirectionMaximize maximizes objective function value
	StudyDirectionMaximize StudyDirection = "maximize"
	// StudyDirectionMinimize minimizes objective function value
	StudyDirectionMinimize StudyDirection = "minimize"
)

// Study corresponds to an optimization task, i.e., a set of trials.
type Study struct {
	id                 string
	storage            Storage
	sampler            Sampler
	direction          StudyDirection
	logger             *zap.Logger
	ignoreObjectiveErr bool
	trialNotifyChan    chan FrozenTrial
	mu                 sync.RWMutex
	ctx                context.Context
}

// GetTrials returns all trials in this study.
func (s *Study) GetTrials() ([]FrozenTrial, error) {
	return s.storage.GetAllTrials(s.id)
}

// Direction returns the direction of objective function value
func (s *Study) Direction() StudyDirection {
	return s.direction
}

// Report reports an objective function value
func (s *Study) Report(trialID string, value float64) error {
	return s.storage.SetTrialValue(trialID, value)
}

// WithContext sets a context and it might cancel the execution of Optimize.
func (s *Study) WithContext(ctx context.Context) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.ctx = ctx
}

func (s *Study) runTrial(objective FuncObjective) (string, error) {
	trialID, err := s.storage.CreateNewTrialID(s.id)
	if err != nil {
		return "", err
	}

	trial := Trial{
		study: s,
		id:    trialID,
	}
	evaluation, objerr := objective(trial)
	if objerr != nil {
		saveerr := s.storage.SetTrialState(trialID, TrialStateFail)
		if saveerr != nil {
			return trialID, saveerr
		}

		if !s.ignoreObjectiveErr {
			return trialID, fmt.Errorf("objective: %s", objerr)
		}
	}

	if err = s.storage.SetTrialValue(trialID, evaluation); err != nil {
		return trialID, err
	}
	if err = s.Report(trialID, evaluation); err != nil {
		return trialID, err
	}
	if err = s.storage.SetTrialState(trialID, TrialStateComplete); err != nil {
		return trialID, err
	}
	return trialID, nil
}

func (s *Study) notifyFinishedTrial(trialID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.trialNotifyChan == nil && s.logger == nil {
		return nil
	}

	trial, err := s.storage.GetTrial(trialID)
	if err != nil {
		return err
	}

	if s.trialNotifyChan != nil {
		s.trialNotifyChan <- trial
	}
	if s.logger != nil {
		s.logger.Info("Finished trial",
			zap.String("trialID", trialID),
			zap.String("state", trial.State.String()),
			zap.Float64("value", trial.Value),
			zap.String("paramsInIR", fmt.Sprintf("%v", trial.ParamsInIR)))
	}
	return nil
}

// Optimize optimizes an objective function.
func (s *Study) Optimize(objective FuncObjective, evaluateMax int) error {
	evaluateCnt := 0
	for {
		if evaluateCnt > evaluateMax {
			break
		}
		evaluateCnt++

		if s.ctx != nil {
			select {
			case <-s.ctx.Done():
				return s.ctx.Err()
			default:
				// do nothing
			}
		}

		trialID, err := s.runTrial(objective)
		if err != nil {
			return err
		}

		err = s.notifyFinishedTrial(trialID)
		if err != nil {
			return err
		}
	}
	return nil
}

// GetBestValue return the best objective value
func (s *Study) GetBestValue() (float64, error) {
	trial, err := s.storage.GetBestTrial(s.id)
	if err != nil {
		return 0.0, err
	}
	return trial.Value, nil
}

// GetBestParams return parameters of the best trial
func (s *Study) GetBestParams() (map[string]float64, error) {
	// TODO: avoid using internal representation value
	trial, err := s.storage.GetBestTrial(s.id)
	if err != nil {
		return nil, err
	}
	return trial.ParamsInIR, nil
}

// CreateStudy creates a new Study object.
func CreateStudy(
	name string,
	opts ...StudyOption,
) (*Study, error) {
	storage := NewInMemoryStorage()
	studyID, err := storage.CreateNewStudyID(name)
	if err != nil {
		return nil, err
	}
	sampler := NewRandomSearchSampler()
	study := &Study{
		id:                 studyID,
		storage:            storage,
		sampler:            sampler,
		direction:          StudyDirectionMinimize,
		logger:             nil,
		ignoreObjectiveErr: false,
	}

	for _, opt := range opts {
		if err := opt(study); err != nil {
			return nil, err
		}
	}
	return study, nil
}

// StudyOption to pass the custom option
type StudyOption func(study *Study) error

// StudyOptionSetDirection change the direction of optimize
func StudyOptionSetDirection(direction StudyDirection) StudyOption {
	return func(s *Study) error {
		s.direction = direction
		return nil
	}
}

// StudyOptionSetLogger sets zap.Logger.
func StudyOptionSetLogger(logger *zap.Logger) StudyOption {
	return func(s *Study) error {
		s.logger = logger
		return nil
	}
}

// StudyOptionStorage sets the storage object.
func StudyOptionStorage(storage Storage) StudyOption {
	return func(s *Study) error {
		s.storage = storage
		return nil
	}
}

// StudyOptionSampler sets the sampler object.
func StudyOptionSampler(sampler Sampler) StudyOption {
	return func(s *Study) error {
		s.sampler = sampler
		return nil
	}
}

// StudyOptionIgnoreObjectiveErr sets the option to ignore error returned from objective function
// If true, Optimize method continues to run new trial.
func StudyOptionIgnoreObjectiveErr(ignore bool) StudyOption {
	return func(s *Study) error {
		s.ignoreObjectiveErr = ignore
		return nil
	}
}

// StudyOptionSetTrialNotifyChannel to subscribe the finished trials.
func StudyOptionSetTrialNotifyChannel(notify chan FrozenTrial) StudyOption {
	return func(s *Study) error {
		s.trialNotifyChan = notify
		return nil
	}
}
