package repository

type ListOptions struct {
	Limit  int
	Offset int
	SortBy string
	Order  string // "ASC" or "DESC"
}

func (o *ListOptions) Validate() error {
	if o.Limit < 0 {
		return ErrInvalidInput
	}
	if o.Offset < 0 {
		return ErrInvalidInput
	}
	if o.Order != "" && o.Order != "ASC" && o.Order != "DESC" {
		return ErrInvalidInput
	}
	return nil
}

func (o *ListOptions) ApplyDefaults() {
	if o.Limit == 0 {
		o.Limit = 50 // Default page size
	}
	if o.Order == "" {
		o.Order = "DESC"
	}
	if o.SortBy == "" {
		o.SortBy = "created_at"
	}
}
