package facade

import "context"

type Facade struct {
	*UserFacade
	*TokenFacade
	*MessageFacade
	*ConversationFacade
}

func New(uf *UserFacade, tf *TokenFacade, mf *MessageFacade, cf *ConversationFacade) *Facade {
	return &Facade{
		UserFacade:         uf,
		TokenFacade:        tf,
		MessageFacade:      mf,
		ConversationFacade: cf,
	}
}

type TXManager interface {
	RunInTX(ctx context.Context, fn func(ctx context.Context) error) error
}
