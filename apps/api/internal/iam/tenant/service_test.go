package tenant

import (
	"context"
	"errors"
	"testing"
)

type fakeRepository struct {
	createInput  CreateInput
	createResult Tenant
	createError  error
}

func (f *fakeRepository) Create(
	ctx context.Context,
	input CreateInput,
) (Tenant, error) {
	f.createInput = input
	return f.createResult, f.createError
}

func (f *fakeRepository) Update(
	ctx context.Context,
	id string,
	input UpdateInput,
) (Tenant, error) {
	return Tenant{}, ErrNotFound
}

func (f *fakeRepository) FindByID(
	ctx context.Context,
	id string,
) (Tenant, error) {
	return Tenant{}, ErrNotFound
}

func (f *fakeRepository) FindBySlug(
	ctx context.Context,
	slug string,
) (Tenant, error) {
	return Tenant{}, ErrNotFound
}

func (f *fakeRepository) ListByUserID(
	ctx context.Context,
	userID string,
) ([]Tenant, error) {
	return nil, nil
}

func TestServiceCreateNormalizesSlug(
	t *testing.T,
) {
	t.Parallel()

	repository := &fakeRepository{
		createResult: Tenant{
			ID:     "tenant-id",
			Name:   "Acme Corporation",
			Slug:   "acme-corporation",
			Status: StatusActive,
		},
	}

	service, err := NewService(repository)
	if err != nil {
		t.Fatalf(
			"NewService() error = %v",
			err,
		)
	}

	_, err = service.Create(
		context.Background(),
		CreateInput{
			Name: " Acme Corporation ",
			Slug: " ACME-CORPORATION ",
		},
	)
	if err != nil {
		t.Fatalf(
			"Create() error = %v",
			err,
		)
	}

	if repository.createInput.Name !=
		"Acme Corporation" {
		t.Fatalf(
			"name = %q",
			repository.createInput.Name,
		)
	}

	if repository.createInput.Slug !=
		"acme-corporation" {
		t.Fatalf(
			"slug = %q",
			repository.createInput.Slug,
		)
	}
}

func TestServiceCreateRejectsInvalidSlug(
	t *testing.T,
) {
	t.Parallel()

	service, err := NewService(
		&fakeRepository{},
	)
	if err != nil {
		t.Fatalf(
			"NewService() error = %v",
			err,
		)
	}

	_, err = service.Create(
		context.Background(),
		CreateInput{
			Name: "Acme",
			Slug: "acme corporation",
		},
	)

	if !errors.Is(
		err,
		ErrInvalidInput,
	) {
		t.Fatalf(
			"error = %v; expected ErrInvalidInput",
			err,
		)
	}
}
