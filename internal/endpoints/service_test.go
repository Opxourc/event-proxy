package endpoints

import (
	"context"
	"errors"
	"reflect"
	"testing"
)

type repositoryStub struct {
	endpoints  []string
	createdURL string
	deletedURL string
	createErr  error
}

func (repo *repositoryStub) GetEndpoints(context.Context) ([]string, error) {
	return repo.endpoints, nil
}

func (repo *repositoryStub) CreateEndpoint(_ context.Context, rawURL string) error {
	repo.createdURL = rawURL
	return repo.createErr
}

func (repo *repositoryStub) DeleteEndpoint(_ context.Context, rawURL string) error {
	repo.deletedURL = rawURL
	return nil
}

type checkerStub struct {
	url string
	err error
}

func (checker *checkerStub) Check(_ context.Context, rawURL string) error {
	checker.url = rawURL
	return checker.err
}

func TestEndpointServiceDelegatesOperations(t *testing.T) {
	repo := &repositoryStub{endpoints: []string{"https://example.com"}}
	checker := &checkerStub{}
	service := NewService(repo, checker)
	ctx := context.Background()

	endpoints, err := service.GetEndpoints(ctx)
	if err != nil {
		t.Fatalf("GetEndpoints() error = %v", err)
	}
	if !reflect.DeepEqual(endpoints, repo.endpoints) {
		t.Fatalf("GetEndpoints() = %v, want %v", endpoints, repo.endpoints)
	}

	const rawURL = "https://new.example.com"
	if err := service.CreateEndpoint(ctx, rawURL); err != nil {
		t.Fatalf("CreateEndpoint() error = %v", err)
	}
	if checker.url != rawURL || repo.createdURL != rawURL {
		t.Fatalf("CreateEndpoint() checker URL = %q, repository URL = %q; want %q", checker.url, repo.createdURL, rawURL)
	}

	if err := service.DeleteEndpoint(ctx, rawURL); err != nil {
		t.Fatalf("DeleteEndpoint() error = %v", err)
	}
	if repo.deletedURL != rawURL {
		t.Fatalf("DeleteEndpoint() URL = %q, want %q", repo.deletedURL, rawURL)
	}
}

func TestEndpointServiceDoesNotStoreUncheckedEndpoint(t *testing.T) {
	checkErr := errors.New("endpoint unavailable")
	repo := &repositoryStub{}
	service := NewService(repo, &checkerStub{err: checkErr})

	err := service.CreateEndpoint(context.Background(), "https://unavailable.example.com")
	if !errors.Is(err, checkErr) {
		t.Fatalf("CreateEndpoint() error = %v, want %v", err, checkErr)
	}
	if repo.createdURL != "" {
		t.Fatalf("CreateEndpoint() stored URL %q after checker failure", repo.createdURL)
	}
}
