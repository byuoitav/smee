package redis

func defaultQueries() map[string]query {
	return map[string]query{
		"": func(key string, val []byte) bool {
			return false
		},
	}
}
