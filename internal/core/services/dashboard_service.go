package services

import (
	"construct-backend/internal/core/domain"
	"construct-backend/internal/core/ports"
)

type DashboardService struct {
	dashboardRepo ports.DashboardRepository
}

func NewDashboardService(dashboardRepo ports.DashboardRepository) *DashboardService {
	return &DashboardService{
		dashboardRepo: dashboardRepo,
	}
}

func (s *DashboardService) GetMetrics(companyID string) (*domain.DashboardMetrics, error) {
	projectsInProgress, err := s.dashboardRepo.CountProjectsInProgress(companyID)
	if err != nil {
		return nil, err
	}

	completedProjects, err := s.dashboardRepo.CountCompletedProjects(companyID)
	if err != nil {
		return nil, err
	}

	activeTasks, err := s.dashboardRepo.CountActiveTasks(companyID)
	if err != nil {
		return nil, err
	}

	linkClicks, err := s.dashboardRepo.CountLinkClicksByCompany(companyID)
	if err != nil {
		return nil, err
	}

	clientsCount, err := s.dashboardRepo.CountClientsByCompany(companyID)
	if err != nil {
		return nil, err
	}

	return &domain.DashboardMetrics{
		ProjectsInProgress: projectsInProgress,
		CompletedProjects:  completedProjects,
		ActiveTasks:        activeTasks,
		LinkClicks:         linkClicks,
		ClientsCount:       clientsCount,
	}, nil
}
