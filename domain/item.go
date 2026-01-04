package domain

type Item struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type Items []Item
