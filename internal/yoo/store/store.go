package store

import (
	"gorm.io/gorm"
	"sync"
)

var (
	once sync.Once
	S    *database
)

type IStore interface {
	Users() UserStore
	Templates() TemplateStore
	Projects() ProjectStore
	Plans() PlanStore
	Tasks() TaskStore
}

type database struct {
	db *gorm.DB
}

var _ IStore = (*database)(nil)

// NewStore returns a new store.
func NewStore(db *gorm.DB) *database {
	once.Do(func() {
		S = &database{db: db}
	})
	return S
}

func (ds *database) Users() UserStore {
	return newUsers(ds.db)
}

func (ds *database) Templates() TemplateStore {
	return newTemplates(ds.db)
}

func (ds *database) Projects() ProjectStore {
	return newProjects(ds.db)
}

func (ds *database) Plans() PlanStore {
	return newPlans(ds.db)
}

func (ds *database) Tasks() TaskStore {
	return newTasks(ds.db)
}
