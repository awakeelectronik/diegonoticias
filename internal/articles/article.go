package articles

import "time"

type Frontmatter struct {
	Title       string    `yaml:"title"`
	Slug        string    `yaml:"slug"`
	Date        time.Time `yaml:"date"`
	Description string    `yaml:"description"`
	Tone        string    `yaml:"tone"`
	Category    string    `yaml:"category"`
	Image       string    `yaml:"image,omitempty"`
	ImageAlt    string    `yaml:"imageAlt,omitempty"`
	Draft       bool      `yaml:"draft"`
	WordCount   int       `yaml:"wordCount"`
}

type Article struct {
	Frontmatter
	Body string `json:"body"`
}

