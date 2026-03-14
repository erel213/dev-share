package sqlite

func joinColumns(cols []string) string {
	result := ""
	for i, c := range cols {
		if i > 0 {
			result += ", "
		}
		result += c
	}
	return result
}

func nilIfEmpty(s string) interface{} {
	if s == "" {
		return nil
	}
	return s
}
