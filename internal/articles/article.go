package articles

import "time"

type Frontmatter struct {
	Title       string    `yaml:"title" json:"title"`
	Slug        string    `yaml:"slug" json:"slug"`
	Date        time.Time `yaml:"date" json:"date"`
	Description string    `yaml:"description" json:"description"`
	Tone        string    `yaml:"tone" json:"tone"`
	Category    string    `yaml:"category" json:"category"`
	Image       string    `yaml:"image,omitempty" json:"image,omitempty"`
	ImageAlt    string    `yaml:"imageAlt,omitempty" json:"imageAlt,omitempty"`
	Draft       bool      `yaml:"draft" json:"draft"`
	WordCount   int       `yaml:"wordCount" json:"wordCount"`
}

type Article struct {
	Frontmatter
	Body string `json:"body"`
}

