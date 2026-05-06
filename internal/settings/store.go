package settings

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type AdSense struct {
	Enabled      bool   `json:"enabled"`
	ClientID     string `json:"clientId"`
	Slot1ID      string `json:"slot1Id"`
	Slot1Enabled bool   `json:"slot1Enabled"`
	Slot2ID      string `json:"slot2Id"`
	Slot2Enabled bool   `json:"slot2Enabled"`
}

type Settings struct {
	SiteName        string  `json:"siteName"`
	SiteDescription string  `json:"siteDescription"`
	SiteURL         string  `json:"siteUrl"`
	DefaultOgImage  string  `json:"defaultOgImage"`
	TwitterHandle   string  `json:"twitterHandle"`
	AdSense         AdSense `json:"adsense"`
}

type Store struct {
	path string
}

func New(path string) *Store { return &Store{path: path} }

func (s *Store) defaultSettings() Settings {
	return Settings{
		SiteName:        "Diego Noticias",
		SiteDescription: "Noticias breves y al grano.",
		SiteURL:         "https://diegonoticias.com",
		DefaultOgImage:  "/og-default.jpg",
	}
}

func (s *Store) Get() (Settings, error) {
	b, err := os.ReadFile(s.path)
	if err != nil {
		if os.IsNotExist(err) {
			def := s.defaultSettings()
			_ = s.Save(def)
			return def, nil
		}
		return Settings{}, err
	}
	var out Settings
	if err := json.Unmarshal(b, &out); err != nil {
		return Settings{}, err
	}
	return out, nil
}

func (s *Store) Save(v Settings) error {
	if err := os.MkdirAll(filepath.Dir(s.path), 0o755); err != nil {
		return err
	}
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	tmp := s.path + ".tmp"
	if err := os.WriteFile(tmp, b, 0o600); err != nil {
		return err
	}
	return os.Rename(tmp, s.path)
}

