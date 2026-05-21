package mocks

import (
	"context"

	"github.com/google/uuid"
	"github.com/jcsoftdev/pulzifi-back/modules/usage/domain/entities"
	"github.com/jcsoftdev/pulzifi-back/modules/usage/domain/repositories"
)

// MockOrganizationPlanRepository is a hand-rolled mock of repositories.OrganizationPlanRepository.
type MockOrganizationPlanRepository struct {
	ListOrgsResult         []*repositories.OrgWithPlan
	ListOrgsErr            error
	DeactivateActiveErr    error
	InsertErr              error
	GetSchemaNameResult    string
	GetSchemaNameErr       error
	GetActivePlanResult    *repositories.OrgPlanInfo
	GetActivePlanErr       error
	GetActivePlanTenResult *repositories.OrgPlanInfo
	GetActivePlanTenErr    error
}

func (m *MockOrganizationPlanRepository) ListOrgsWithPlans(_ context.Context) ([]*repositories.OrgWithPlan, error) {
	return m.ListOrgsResult, m.ListOrgsErr
}

func (m *MockOrganizationPlanRepository) DeactivateActive(_ context.Context, _ uuid.UUID) error {
	return m.DeactivateActiveErr
}

func (m *MockOrganizationPlanRepository) Insert(_ context.Context, _ *entities.OrganizationPlan) error {
	return m.InsertErr
}

func (m *MockOrganizationPlanRepository) GetSchemaName(_ context.Context, _ uuid.UUID) (string, error) {
	return m.GetSchemaNameResult, m.GetSchemaNameErr
}

func (m *MockOrganizationPlanRepository) GetActivePlanForOrg(_ context.Context, _ uuid.UUID) (*repositories.OrgPlanInfo, error) {
	return m.GetActivePlanResult, m.GetActivePlanErr
}

func (m *MockOrganizationPlanRepository) GetActivePlanForTenant(_ context.Context, _ string) (*repositories.OrgPlanInfo, error) {
	return m.GetActivePlanTenResult, m.GetActivePlanTenErr
}

func (m *MockOrganizationPlanRepository) WithTx(_ interface{}) repositories.OrganizationPlanRepository {
	return m
}
