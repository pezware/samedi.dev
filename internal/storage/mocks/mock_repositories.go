// Copyright (c) 2025 Samedi Contributors
// SPDX-License-Identifier: MIT

package mocks

import (
	"context"

	"github.com/pezware/samedi.dev/internal/storage"
)

// PlanRepositoryMock provides a configurable double for storage.PlanRepository.
type PlanRepositoryMock struct {
	UpsertFunc func(context.Context, *storage.PlanRecord) error
	GetFunc    func(context.Context, string) (*storage.PlanRecord, error)
	ListFunc   func(context.Context, *storage.PlanFilter) ([]*storage.PlanRecord, error)
	DeleteFunc func(context.Context, string) error
}

func (m *PlanRepositoryMock) Upsert(ctx context.Context, plan *storage.PlanRecord) error {
	if m.UpsertFunc != nil {
		return m.UpsertFunc(ctx, plan)
	}
	return nil
}

func (m *PlanRepositoryMock) Get(ctx context.Context, id string) (*storage.PlanRecord, error) {
	if m.GetFunc != nil {
		return m.GetFunc(ctx, id)
	}
	return nil, nil
}

func (m *PlanRepositoryMock) List(ctx context.Context, filter *storage.PlanFilter) ([]*storage.PlanRecord, error) {
	if m.ListFunc != nil {
		return m.ListFunc(ctx, filter)
	}
	return nil, nil
}

func (m *PlanRepositoryMock) Delete(ctx context.Context, id string) error {
	if m.DeleteFunc != nil {
		return m.DeleteFunc(ctx, id)
	}
	return nil
}

// SessionRepositoryMock provides a configurable double for storage.SessionRepository.
type SessionRepositoryMock struct {
	CreateFunc func(context.Context, *storage.SessionRecord) error
	GetFunc    func(context.Context, string) (*storage.SessionRecord, error)
	ListFunc   func(context.Context, *storage.SessionFilter) ([]*storage.SessionRecord, error)
	UpdateFunc func(context.Context, *storage.SessionRecord) error
	DeleteFunc func(context.Context, string) error
}

func (m *SessionRepositoryMock) Create(ctx context.Context, session *storage.SessionRecord) error {
	if m.CreateFunc != nil {
		return m.CreateFunc(ctx, session)
	}
	return nil
}

func (m *SessionRepositoryMock) Get(ctx context.Context, id string) (*storage.SessionRecord, error) {
	if m.GetFunc != nil {
		return m.GetFunc(ctx, id)
	}
	return nil, nil
}

func (m *SessionRepositoryMock) List(ctx context.Context, filter *storage.SessionFilter) ([]*storage.SessionRecord, error) {
	if m.ListFunc != nil {
		return m.ListFunc(ctx, filter)
	}
	return nil, nil
}

func (m *SessionRepositoryMock) Update(ctx context.Context, session *storage.SessionRecord) error {
	if m.UpdateFunc != nil {
		return m.UpdateFunc(ctx, session)
	}
	return nil
}

func (m *SessionRepositoryMock) Delete(ctx context.Context, id string) error {
	if m.DeleteFunc != nil {
		return m.DeleteFunc(ctx, id)
	}
	return nil
}

// CardRepositoryMock provides a configurable double for storage.CardRepository.
type CardRepositoryMock struct {
	UpsertFunc func(context.Context, *storage.CardRecord) error
	GetFunc    func(context.Context, string) (*storage.CardRecord, error)
	ListFunc   func(context.Context, *storage.CardFilter) ([]*storage.CardRecord, error)
	DeleteFunc func(context.Context, string) error
}

func (m *CardRepositoryMock) Upsert(ctx context.Context, card *storage.CardRecord) error {
	if m.UpsertFunc != nil {
		return m.UpsertFunc(ctx, card)
	}
	return nil
}

func (m *CardRepositoryMock) Get(ctx context.Context, id string) (*storage.CardRecord, error) {
	if m.GetFunc != nil {
		return m.GetFunc(ctx, id)
	}
	return nil, nil
}

func (m *CardRepositoryMock) List(ctx context.Context, filter *storage.CardFilter) ([]*storage.CardRecord, error) {
	if m.ListFunc != nil {
		return m.ListFunc(ctx, filter)
	}
	return nil, nil
}

func (m *CardRepositoryMock) Delete(ctx context.Context, id string) error {
	if m.DeleteFunc != nil {
		return m.DeleteFunc(ctx, id)
	}
	return nil
}

var (
	_ storage.PlanRepository    = (*PlanRepositoryMock)(nil)
	_ storage.SessionRepository = (*SessionRepositoryMock)(nil)
	_ storage.CardRepository    = (*CardRepositoryMock)(nil)
)
