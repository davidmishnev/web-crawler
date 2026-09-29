package document

type Metadata struct {
	Author      string            `json:"author,omitempty"`
	Language    string            `json:"language,omitempty"`
	Description string            `json:"description,omitempty"`
	Extra       map[string]string `json:"extra,omitempty"`
}
