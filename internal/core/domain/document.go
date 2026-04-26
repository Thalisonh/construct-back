package domain

import "time"

type ClientDocument struct {
	ID          string    `json:"id" gorm:"primaryKey"`
	ClientID    string    `json:"client_id" gorm:"index"`
	CompanyID   string    `json:"company_id" gorm:"index"`
	FileName    string    `json:"file_name"`
	ContentType string    `json:"content_type"`
	FileSize    int64     `json:"file_size"`
	StorageKey  string    `json:"storage_key"`
	UploadedBy  string    `json:"uploaded_by" gorm:"index"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	ViewURL     string    `json:"view_url,omitempty" gorm:"-"`
	DownloadURL string    `json:"download_url,omitempty" gorm:"-"`
}

type DiaryDocument struct {
	ID           string    `json:"id" gorm:"primaryKey"`
	ProjectID    string    `json:"project_id" gorm:"index"`
	DiaryEntryID string    `json:"diary_entry_id" gorm:"index"`
	CompanyID    string    `json:"company_id" gorm:"index"`
	FileName     string    `json:"file_name"`
	ContentType  string    `json:"content_type"`
	FileSize     int64     `json:"file_size"`
	StorageKey   string    `json:"storage_key"`
	UploadedBy   string    `json:"uploaded_by" gorm:"index"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	ViewURL      string    `json:"view_url,omitempty" gorm:"-"`
	DownloadURL  string    `json:"download_url,omitempty" gorm:"-"`
}
