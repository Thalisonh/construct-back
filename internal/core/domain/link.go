package domain

import (
	"time"
)

type Link struct {
	ID          string    `bson:"_id" json:"id" datastore:"-" gorm:"primaryKey"`
	URL         string    `bson:"url" json:"url" datastore:"url"`
	Description string    `bson:"description" json:"description" datastore:"description"`
	UserID      string    `bson:"user_id" json:"user_id" datastore:"user_id" gorm:"index"`
	CompanyID   string    `bson:"company_id" json:"company_id" datastore:"company_id" gorm:"index"`
	CreatedAt   time.Time `bson:"created_at" json:"created_at" datastore:"created_at"`
	UpdatedAt   time.Time `bson:"updated_at" json:"updated_at" datastore:"updated_at"`
	Count       int       `bson:"count" json:"count" datastore:"count"`
}

type LinkClick struct {
	ID        string    `bson:"_id" json:"id" datastore:"-" gorm:"primaryKey"`
	LinkID    string    `bson:"link_id" json:"link_id" datastore:"link_id" gorm:"index"`
	CreatedAt time.Time `bson:"created_at" json:"created_at" datastore:"created_at"`
}

type LinkAnalyticsItem struct {
	ID           string  `json:"id"`
	Description  string  `json:"description"`
	URL          string  `json:"url"`
	Clicks       int64   `json:"clicks" gorm:"column:clicks"`
	SharePercent float64 `json:"share_percent"`
}

type LinkAnalyticsSummary struct {
	TotalLinks           int     `json:"total_links"`
	TotalClicks          int64   `json:"total_clicks"`
	TopLinkID            string  `json:"top_link_id"`
	TopLinkDescription   string  `json:"top_link_description"`
	TopLinkClicks        int64   `json:"top_link_clicks"`
	AverageClicksPerLink float64 `json:"average_clicks_per_link"`
}

type LinkAnalyticsFilters struct {
	StartDate string `json:"start_date"`
	EndDate   string `json:"end_date"`
}

type LinkAnalyticsResponse struct {
	Summary LinkAnalyticsSummary  `json:"summary"`
	Filters *LinkAnalyticsFilters `json:"filters,omitempty"`
	Links   []LinkAnalyticsItem   `json:"links"`
}
