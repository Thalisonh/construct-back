package domain

type DashboardMetrics struct {
	ProjectsInProgress int64 `json:"projects_in_progress"`
	CompletedProjects  int64 `json:"completed_projects"`
	ActiveTasks        int64 `json:"active_tasks"`
	LinkClicks         int64 `json:"link_clicks"`
	ClientsCount       int64 `json:"clients_count"`
}
