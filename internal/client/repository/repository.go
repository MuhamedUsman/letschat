package repository

type LocalRepository struct {
	LocalUserRepository
	LocalConversationRepository
}

func NewLocalRepository(db *DB) *LocalRepository {
	return &LocalRepository{
		LocalUserRepository:         newLocalUserRepository(db),
		LocalConversationRepository: NewLocalConversationRepository(db),
	}
}
