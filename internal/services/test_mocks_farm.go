package services

import (
	"context"

	"github.com/Kisanlink/farmers-module/internal/entities/farm"
	"github.com/Kisanlink/farmers-module/internal/entities/requests"
	"github.com/stretchr/testify/mock"
)

// MockFarmRepository is a mock implementation of FarmRepository
type MockFarmRepository struct {
	mock.Mock
}

func (m *MockFarmRepository) Create(ctx context.Context, farm *farm.Farm) error {
	args := m.Called(ctx, farm)
	return args.Error(0)
}

func (m *MockFarmRepository) GetByID(ctx context.Context, id string) (*farm.Farm, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*farm.Farm), args.Error(1)
}

func (m *MockFarmRepository) Update(ctx context.Context, farm *farm.Farm) error {
	args := m.Called(ctx, farm)
	return args.Error(0)
}

func (m *MockFarmRepository) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockFarmRepository) Find(ctx context.Context, filter interface{}) ([]*farm.Farm, error) {
	args := m.Called(ctx, filter)
	return args.Get(0).([]*farm.Farm), args.Error(1)
}

func (m *MockFarmRepository) Count(ctx context.Context, filter interface{}, model interface{}) (int64, error) {
	args := m.Called(ctx, filter, model)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockFarmRepository) ListByBoundingBox(ctx context.Context, bbox requests.BoundingBox, filters map[string]interface{}) ([]*farm.Farm, error) {
	args := m.Called(ctx, bbox, filters)
	return args.Get(0).([]*farm.Farm), args.Error(1)
}

func (m *MockFarmRepository) ValidateGeometry(ctx context.Context, wkt string) error {
	args := m.Called(ctx, wkt)
	return args.Error(0)
}

func (m *MockFarmRepository) CheckOverlap(ctx context.Context, wkt string, excludeFarmID string, orgID string) (bool, []string, float64, error) {
	args := m.Called(ctx, wkt, excludeFarmID, orgID)
	return args.Bool(0), args.Get(1).([]string), args.Get(2).(float64), args.Error(3)
}

func (m *MockFarmRepository) FindOne(ctx context.Context, filter interface{}) (*farm.Farm, error) {
	args := m.Called(ctx, filter)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*farm.Farm), args.Error(1)
}
