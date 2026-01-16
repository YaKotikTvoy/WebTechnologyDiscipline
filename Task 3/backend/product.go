package main

type Product struct {
	ID          int     `json:"id"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Price       float64 `json:"price"`
	Image       string  `json:"image"`
	Stock       int     `json:"stock"`
	CreatedAt   string  `json:"created_at,omitempty"`
}

type Stats struct {
	Total      int `json:"total"`
	InStock    int `json:"inStock"`
	LowStock   int `json:"lowStock"`
	OutOfStock int `json:"outOfStock"`
}