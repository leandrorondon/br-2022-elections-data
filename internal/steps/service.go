package steps

import (
	"context"
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
		// Ignore and keep going...
		log.Printf("Erro ao verificar se passo %s já foi executado - %v.", step, err)
	}

	if stepCompleted {
		log.Printf("Dados do passo %s já foram salvos anteriormente.", step)
		return nil
	}

	log.Printf("Executando passo %s.", step)

	err = fn(ctx)
	if err != nil {
		log.Printf("Erro ao executar passo %s: %v.", step, err)
		return err
	}

	err = s.repository.Insert(ctx, step)
	if err != nil {
		// Ignore and keep going...
		log.Printf("Erro ao guardar execução do passo %s - %v.", step, err)
	}

	return nil
}
