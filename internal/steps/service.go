package steps

import (
	"context"
	"errors"
	"log"

	"github.com/jmoiron/sqlx"
)

type Service struct {
	repository *Repository
}

func NewService(db *sqlx.DB) *Service {
	return &Service{
		repository: NewRepository(db),
	}
}

func (s *Service) Execute(ctx context.Context, step string, fn func(context.Context) error) error {
	stepCompleted, err := s.repository.Exists(ctx, step)
	if err != nil {
		if errors.Is(err, context.Canceled) {
			return err
		}
		log.Printf("[%s] Erro ao verificar se passo já foi executado - %v.", step, err)
	}

	if stepCompleted {
		log.Printf("[%s] Passo já executado anteriormente.", step)
		return nil
	}

	log.Printf("[%s] Executando passo.", step)

	err = fn(ctx)
	if err != nil {
		log.Printf("[%s] Erro ao executar passo: %v.", step, err)
		return err
	}

	err = s.repository.Insert(ctx, step)
	if err != nil {
		// Ignore and keep going...
		log.Printf("[%s] Erro ao guardar execução do passo: %v.", step, err)
	}

	log.Printf("[%s] Sucesso.", step)

	return nil
}
