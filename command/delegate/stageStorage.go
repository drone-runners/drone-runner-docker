package delegate

import (
	"fmt"
	"sync"
	"time"

	"github.com/drone/runner-go/pipeline/runtime"
)

// Stages is storage for stages
// Defined as interface to later add support for storage in external databases.
// TODO: Currently defined as global variable, change to use dependency injection
var Stages StageStorage

func init() {
	s := &storage{}
	s.storage = make(map[string]*stageStorageEntry)
	Stages = s
}

type StageStorage interface {
	Store(id string, spec runtime.Spec, secrets []SecretInfo) error
	Remove(id string) (bool, error)
	Get(id string) (runtime.Spec, []SecretInfo, error)
}

type storage struct {
	sync.Mutex
	storage map[string]*stageStorageEntry
}

type stageStorageEntry struct {
	sync.Mutex

	AddedAt time.Time
	Spec    runtime.Spec
	Secrets []SecretInfo
}

func (s *storage) Store(id string, spec runtime.Spec, secrets []SecretInfo) error {
	s.Lock()
	defer s.Unlock()

	_, ok := s.storage[id]
	if ok {
		return fmt.Errorf("stage with id=%s already present", id)
	}

	s.storage[id] = &stageStorageEntry{
		AddedAt: time.Now(),
		Spec:    spec,
		Secrets: secrets,
	}

	return nil
}

func (s *storage) Remove(id string) (bool, error) {
	s.Lock()
	defer s.Unlock()

	_, ok := s.storage[id]
	if !ok {
		return false, nil
	}

	delete(s.storage, id)

	return true, nil
}

func (s *storage) Get(id string) (runtime.Spec, []SecretInfo, error) {
	s.Lock()
	defer s.Unlock()

	entry := s.storage[id]

	return entry.Spec, entry.Secrets, nil
}
