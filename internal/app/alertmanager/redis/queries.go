package redis

func defaultQueries() map[string]query {
	return map[string]query{
		"": func() bool {
			return false
		},
	}
}
