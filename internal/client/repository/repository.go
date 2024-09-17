package repository

type LocalRepository struct {
	LocalUserRepository
	LocalConversationRepository
	LocalMessageRepository
}

func NewLocalRepository(db *DB) *LocalRepository {
	return &LocalRepository{
		LocalUserRepository:         newLocalUserRepository(db),
		LocalConversationRepository: NewLocalConversationRepository(db),
		LocalMessageRepository:      NewLocalMessageRepository(db),
	}
}
