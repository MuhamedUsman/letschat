package facade

import "context"

type Facade struct {
	*UserFacade
	*TokenFacade
	*MessageFacade
}

func New(uf *UserFacade, tf *TokenFacade, mf *MessageFacade) *Facade {
	return &Facade{
		UserFacade:    uf,
		TokenFacade:   tf,
		MessageFacade: mf,
	}
}

type TXManager interface {
	RunInTX(ctx context.Context, fn func(ctx context.Context) error) error
}
