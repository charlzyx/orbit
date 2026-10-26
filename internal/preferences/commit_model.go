package preferences

import (
	"errors"
	"os"
)

// CommitModel is Orbit-owned, separate from Pi/session defaults. Nil follows the
// active session; an explicit provider/id is retained when absent from catalog.
func (s *Store) CommitModel() (*Model, error) {
	var value struct {
		Model *Model `json:"model"`
	}
	err := s.Read("commit-model.json", &value)
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if value.Model == nil || value.Model.Provider == "" || value.Model.ID == "" {
		return nil, nil
	}
	return value.Model, nil
}
func (s *Store) SaveCommitModel(model *Model) error {
	return s.Write("commit-model.json", struct {
		Model *Model `json:"model"`
	}{model})
}
