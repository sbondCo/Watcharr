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

type RootFolder struct {
	Path            string `json:"path"`
	Accessible      bool   `json:"accessible"`
	FreeSpace       int64  `json:"freeSpace"`
	UnmappedFolders []any  `json:"unmappedFolders"`
	ID              int    `json:"id"`
}

type LanguageProfile struct {
	Name           string `json:"name"`
	UpgradeAllowed bool   `json:"upgradeAllowed"`
	Cutoff         struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	} `json:"cutoff"`
	Languages []struct {
		Language struct {
			ID   int    `json:"id"`
			Name string `json:"name"`
		} `json:"language"`
		Allowed bool `json:"allowed"`
	} `json:"languages"`
	ID int `json:"id"`
}
