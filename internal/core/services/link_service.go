package services

import (
	"construct-backend/internal/core/domain"
	"construct-backend/internal/core/ports"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type LinkService struct {
	linkRepo ports.LinkRepository
}

func NewLinkService(linkRepo ports.LinkRepository) *LinkService {
	return &LinkService{
		linkRepo: linkRepo,
	}
}

func (s *LinkService) CreateLink(companyID, userID, url, description string) (*domain.Link, error) {
	link := &domain.Link{
		ID:          uuid.New().String(),
		URL:         url,
		Description: description,
		UserID:      userID,
		CompanyID:   companyID,
		CreatedAt:   time.Now(),
	}

	if err := s.linkRepo.CreateLink(link); err != nil {
		return nil, err
	}

	return link, nil
}

func (s *LinkService) UpdateLink(companyID, url, description, id string) (*domain.Link, error) {
	link := &domain.Link{
		ID:          id,
		URL:         url,
		Description: description,
		CompanyID:   companyID,
		UpdatedAt:   time.Now(),
	}

	err := s.linkRepo.UpdateLink(link)
	if err != nil {
		return nil, err
	}

	return link, nil
}

func (s *LinkService) ListLinks(companyID string) ([]domain.Link, error) {
	return s.linkRepo.GetAllLinks(companyID)
}

func (s *LinkService) GetLinkAnalytics(companyID, startDate, endDate string) (*domain.LinkAnalyticsResponse, error) {
	var (
		startTime *time.Time
		endTime   *time.Time
		filters   *domain.LinkAnalyticsFilters
	)

	if (startDate == "" && endDate != "") || (startDate != "" && endDate == "") {
		return nil, fmt.Errorf("start_date and end_date must be provided together")
	}

	if startDate != "" && endDate != "" {
		parsedStart, err := time.Parse("2006-01-02", startDate)
		if err != nil {
			return nil, fmt.Errorf("invalid start_date")
		}

		parsedEnd, err := time.Parse("2006-01-02", endDate)
		if err != nil {
			return nil, fmt.Errorf("invalid end_date")
		}

		if parsedStart.After(parsedEnd) {
			return nil, fmt.Errorf("start_date cannot be after end_date")
		}

		endExclusive := parsedEnd.AddDate(0, 0, 1)
		startTime = &parsedStart
		endTime = &endExclusive
		filters = &domain.LinkAnalyticsFilters{
			StartDate: startDate,
			EndDate:   endDate,
		}
	}

	analytics, err := s.linkRepo.GetLinkAnalytics(companyID, startTime, endTime)
	if err != nil {
		return nil, err
	}

	response := &domain.LinkAnalyticsResponse{
		Summary: domain.LinkAnalyticsSummary{},
		Filters: filters,
		Links:   analytics,
	}

	response.Summary.TotalLinks = len(analytics)

	for _, item := range analytics {
		response.Summary.TotalClicks += item.Clicks
		if item.Clicks > response.Summary.TopLinkClicks {
			response.Summary.TopLinkClicks = item.Clicks
			response.Summary.TopLinkID = item.ID
			response.Summary.TopLinkDescription = item.Description
		}
	}

	if response.Summary.TotalLinks > 0 {
		response.Summary.AverageClicksPerLink = float64(response.Summary.TotalClicks) / float64(response.Summary.TotalLinks)
	}

	if response.Summary.TotalClicks > 0 {
		for index := range response.Links {
			response.Links[index].SharePercent = (float64(response.Links[index].Clicks) / float64(response.Summary.TotalClicks)) * 100
		}
	}

	return response, nil
}

func (s *LinkService) DeleteLink(id, companyID string) error {
	return s.linkRepo.DeleteLink(id, companyID)
}

func (s *LinkService) TrackLinkClick(id string) error {
	return s.linkRepo.RegisterClick(id)
}
