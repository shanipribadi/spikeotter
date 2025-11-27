package spikeotter

import (
	"context"
	"fmt"
	"time"

	"github.com/jaswdr/faker/v2"
)

type Source struct {
	faker *faker.Faker
}

func NewSource() *Source {
	faker := faker.New()
	return &Source{
		faker: &faker,
	}
}

type Model struct {
	ID        string
	CreatedAt time.Time
	FirstName string
	LastName  string
	Address   string
	Age       int
}

func (s *Source) get(id string) (*Model, error) {
	return &Model{
		ID:        id,
		CreatedAt: time.Now(),
		FirstName: s.faker.Person().FirstName(),
		LastName:  s.faker.Person().LastName(),
		Address:   s.faker.Address().Address(),
		Age:       s.faker.Int(),
	}, nil
}

func (s *Source) Get(ctx context.Context, id string) (*Model, error) {
	time.Sleep(time.Microsecond * time.Duration(s.faker.IntBetween(50, 450)))
	return s.get(id)
}

func (s *Source) BulkGet(ctx context.Context, ids []string) (map[string]*Model, error) {
	time.Sleep(time.Microsecond * time.Duration(s.faker.IntBetween(500, 5000)))
	models := make(map[string]*Model, len(ids))
	for _, id := range ids {
		model, _ := s.get(id)
		models[id] = model
	}

	return models, nil
}

func (s *Source) GenIDs() []string {
	n := s.faker.IntBetween(20, 50)
	ids := make([]string, n)
	for i := range n {
		id := s.faker.IntBetween(1, 3000000)
		ids[i] = fmt.Sprintf("%09d", id)
	}
	return ids
}
