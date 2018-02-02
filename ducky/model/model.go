package model

type Duck struct {
	Name string
}

type DuckRepository interface {
	GetAll() ([]Duck, error)

	GetByName(name string) (*Duck, error)

	Add(duck Duck) error
}

type InMemoryDuckRepository struct {
	Ducks map[string]Duck
}

type Token struct {
	Owner                string
	Value                string
	CreatedAt            string
	Valid                bool
	BasicAuthHeaderValue string
}

func (r InMemoryDuckRepository) GetAll() ([]Duck, error) {
	ducks := make([]Duck, len(r.Ducks))
	i := 0
	for _, d := range r.Ducks {
		ducks[i] = d
		i++
	}
	return ducks, nil
}

func (r InMemoryDuckRepository) GetByName(name string) (*Duck, error) {
	duck, ok := r.Ducks[name]
	if ok {
		return &duck, nil
	}
	return nil, nil
}

func (r InMemoryDuckRepository) Add(d Duck) error {
	r.Ducks[d.Name] = d
	return nil
}
