package arr

type QualityProfile struct {
	Name           string `json:"name"`
	UpgradeAllowed bool   `json:"upgradeAllowed"`
	Cutoff         int    `json:"cutoff"`
	Items          []struct {
		Quality struct {
			ID         int    `json:"id"`
			Name       string `json:"name"`
			Source     string `json:"source"`
			Resolution int    `json:"resolution"`
		} `json:"quality,omitempty"`
		Items   []any  `json:"items"`
		Allowed bool   `json:"allowed"`
		Name    string `json:"name,omitempty"`
		ID      int    `json:"id,omitempty"`
	} `json:"items"`
	ID int `json:"id"`
}
