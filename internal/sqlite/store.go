package sqlite

import (
	"context"
	"time"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

type Store struct {
	Path           string
	db             *gorm.DB
	Users          *UsersRepository
	Projects       *ProjectsRepository
	ProjectMembers *ProjectMembersRepository
	Scripts        *ScriptsRepository
}

func NewStore() *Store {
	return &Store{Path: "sparkedge.db"}
}

func (s *Store) Open(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if s.db != nil {
		return nil
	}

	db, err := gorm.Open(sqlite.Open(s.Path), &gorm.Config{
		NowFunc: func() time.Time {
			return time.Now().UTC()
		},
	})
	if err != nil {
		return err
	}
	if err := migrate(db.WithContext(ctx)); err != nil {
		return err
	}

	s.db = db
	s.Users = NewUsersRepository(db)
	s.Projects = NewProjectsRepository(db)
	s.ProjectMembers = NewProjectMembersRepository(db)
	s.Scripts = NewScriptsRepository(db)
	return nil
}

func (s *Store) Close() error {
	if s.db == nil {
		return nil
	}
	db, err := s.db.DB()
	if err != nil {
		return err
	}
	return db.Close()
}

func migrate(db *gorm.DB) error {
	return db.AutoMigrate(&userModel{}, &projectModel{}, &projectMemberModel{}, &downloadedScriptModel{})
}
