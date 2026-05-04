package task

import (
	"fmt"
	"time"
)

type DateGenerator interface {
	Generate(cfg RecurrenceConfig) ([]time.Time, error)
}

type GeneratorRegistry struct {
	strategies map[RecurrenceType]DateGenerator
}

func NewGeneratorRegistry() *GeneratorRegistry {
	return &GeneratorRegistry{
		strategies: make(map[RecurrenceType]DateGenerator),
	}
}

func (r *GeneratorRegistry) Register(t RecurrenceType, gen DateGenerator) {
	r.strategies[t] = gen
}

func (r *GeneratorRegistry) Get(t RecurrenceType) (DateGenerator, error) {
	gen, ok := r.strategies[t]
	if !ok {
		return nil, fmt.Errorf("%w: %q", ErrUnknownType, t)
	}
	return gen, nil
}