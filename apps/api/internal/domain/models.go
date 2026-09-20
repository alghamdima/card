package domain

import "time"

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
	ID        int64       `json:"id"`
	Slug      string      `json:"slug"`
	Title     string      `json:"title"`
	Lang      string      `json:"lang"`
	TextColor string      `json:"textColor"`
	HeadColor string      `json:"headColor"`
	Boxes     BoxesConfig `json:"boxes"`
	Image     string      `json:"image"`
	Thumb     string      `json:"thumb,omitempty"`
	Active    bool        `json:"active"`
	CreatedAt time.Time   `json:"createdAt"`
}

type CampaignSummary struct {
	Slug       string    `json:"slug"`
	Title      string    `json:"title"`
	Lang       string    `json:"lang"`
	TextColor  string    `json:"textColor"`
	HeadColor  string    `json:"headColor"`
	Thumb      string    `json:"thumb,omitempty"`
	Active     bool      `json:"active"`
	TotalCards int       `json:"totalCards"`
	CreatedAt  time.Time `json:"createdAt"`
}

type Card struct {
	ID           int64     `json:"id"`
	CampaignSlug string    `json:"campaignSlug"`
	FromName     string    `json:"from"`
	ToName       string    `json:"to"`
	Message      string    `json:"message"`
	Heading      string    `json:"heading,omitempty"`
	Device       string    `json:"device"`
	DateStr      string    `json:"date"`
	TimeStr      string    `json:"time"`
	CreatedAt    time.Time `json:"createdAt"`
}

type DashboardStats struct {
	TotalCampaigns int            `json:"totalCampaigns"`
	ActiveCampaigns int           `json:"activeCampaigns"`
	TotalCards     int            `json:"totalCards"`
	CardsToday     int            `json:"cardsToday"`
	DeviceStats    map[string]int `json:"deviceStats"`
}
