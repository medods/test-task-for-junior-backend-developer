package task

import (
	"errors"
	"time"
)

type Specification interface {
	IsSatisfiedBy(cfg RecurrenceConfig) error
}

type AndSpec []Specification

func (s AndSpec) IsSatisfiedBy(cfg RecurrenceConfig) error {
	var errs []error
	for _, spec := range s {
		if err := spec.IsSatisfiedBy(cfg); err != nil {
			errs = append(errs, err)
		}
	}
	if len(errs) > 0 {
		return errors.Join(errs...)
	}
	return nil
}

type OrSpec []Specification

func (s OrSpec) IsSatisfiedBy(cfg RecurrenceConfig) error {
	for _, spec := range s {
		if err := spec.IsSatisfiedBy(cfg); err == nil {
			return nil
		}
	}
	return errors.New("at least one condition must be satisfied")
}

type NotSpec struct{ Spec Specification }

func (s NotSpec) IsSatisfiedBy(cfg RecurrenceConfig) error {
	if err := s.Spec.IsSatisfiedBy(cfg); err == nil {
		return errors.New("condition must not be satisfied")
	}
	return nil
}