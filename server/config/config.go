package config




type Config struct {
	Endpoints map[string]string `json:"endpoints"`
	JWTSecret string `json:"jwt_secret"`
	JWTRefreshSecret string `json:"jwt_refresh_secret"`
	Port  string `json:"port"`
}

