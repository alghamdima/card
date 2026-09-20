package domain

import "time"

// Dynamic Text Field Config for draggable/resizable template editor
type TextFieldConfig struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Label       string  `json:"label"`
	Placeholder string  `json:"placeholder,omitempty"`
	X           float64 `json:"x"`
	Y           float64 `json:"y"`
	Width       float64 `json:"width"`
	Height      float64 `json:"height"`
	FontSize    float64 `json:"fontSize"`
	Color       string  `json:"color"`
	Weight      string  `json:"weight,omitempty"` // "bold" | "regular"
	Align       string  `json:"align,omitempty"`  // "center" | "right" | "left"
	MaxChars    int     `json:"maxChars,omitempty"`
	Required    bool    `json:"required,omitempty"`
	Order       int     `json:"order"`
}

// TemplateVariant represents a single language variant (e.g. AR or EN)
type TemplateVariant struct {
	Image  string            `json:"image"`
	Fields []TextFieldConfig `json:"fields"`
}

type BoxItem struct {
	Top       float64 `json:"top"`
	Bottom    float64 `json:"bottom"`
	CenterX   float64 `json:"centerX"`
	MaxW      float64 `json:"maxW"`
	Size      float64 `json:"size"`
	Color     string  `json:"color,omitempty"`
	Weight    string  `json:"weight,omitempty"`
	HeadColor string  `json:"headColor,omitempty"`
	BodyColor string  `json:"bodyColor,omitempty"`
}

type BoxesConfig struct {
	To      BoxItem `json:"to"`
	Message BoxItem `json:"message"`
	From    BoxItem `json:"from"`
}

type Campaign struct {
	ID         int64            `json:"id"`
	Slug       string           `json:"slug"`
	Title      string           `json:"title"`
	TitleAR    string           `json:"titleAR,omitempty"`
	TitleEN    string           `json:"titleEN,omitempty"`
	Lang       string           `json:"lang"`
	TextColor  string           `json:"textColor"`
	HeadColor  string           `json:"headColor"`
	Boxes      BoxesConfig      `json:"boxes"`
	Image      string           `json:"image"`
	Thumb      string           `json:"thumb,omitempty"`
	TemplateAR *TemplateVariant `json:"templateAR,omitempty"`
	TemplateEN *TemplateVariant `json:"templateEN,omitempty"`
	Active     bool             `json:"active"`
	CreatedAt  time.Time        `json:"createdAt"`
}

type CampaignSummary struct {
	Slug       string    `json:"slug"`
	Title      string    `json:"title"`
	TitleAR    string    `json:"titleAR,omitempty"`
	TitleEN    string    `json:"titleEN,omitempty"`
	Lang       string    `json:"lang"`
	TextColor  string    `json:"textColor"`
	HeadColor  string    `json:"headColor"`
	Thumb      string    `json:"thumb,omitempty"`
	Active     bool      `json:"active"`
	TotalCards int       `json:"totalCards"`
	CreatedAt  time.Time `json:"createdAt"`
}

type Card struct {
	ID           int64                  `json:"id"`
	CampaignSlug string                 `json:"campaignSlug"`
	FromName     string                 `json:"from"`
	ToName       string                 `json:"to"`
	Message      string                 `json:"message"`
	Heading      string                 `json:"heading,omitempty"`
	Lang         string                 `json:"lang,omitempty"`
	FieldValues  map[string]interface{} `json:"fieldValues,omitempty"`
	Device       string                 `json:"device"`
	DateStr      string                 `json:"date"`
	TimeStr      string                 `json:"time"`
	CreatedAt    time.Time              `json:"createdAt"`
}

type DashboardStats struct {
	TotalCampaigns  int            `json:"totalCampaigns"`
	ActiveCampaigns int            `json:"activeCampaigns"`
	TotalCards      int            `json:"totalCards"`
	CardsToday      int            `json:"cardsToday"`
	DeviceStats     map[string]int `json:"deviceStats"`
}

type CampaignAnalytics struct {
	Slug        string `json:"slug"`
	Title       string `json:"title"`
	TotalCards  int    `json:"totalCards"`
	CardsToday  int    `json:"cardsToday"`
	LastCardAt  string `json:"lastCardAt,omitempty"`
}
