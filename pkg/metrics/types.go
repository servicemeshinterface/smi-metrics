package metrics

type Queries struct {
	ResourceQueries map[string]string `yaml:"resourceQueries"`

	EdgeQueries map[string]string `yaml:"edgeQueries"`
}

type errorResponse struct {
	Error string `json:"error"`
}
