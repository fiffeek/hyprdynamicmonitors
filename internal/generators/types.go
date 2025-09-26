package generators

type MonitorSpec struct {
	Name        string `json:"name"`
	ID          *int   `json:"id"`
	Description string `json:"description"`
	Disabled    bool   `json:"disabled"`
}
