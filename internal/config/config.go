package config

import "time"

type Config struct {
	Env          string      `yaml:"env" env-default:"dev"`
	Domain       string      `yaml:"domain"`
	PublicKey    string      `env:"PUBLIC_KEY" env-required:"true"`
	GRPC_Clients GrpcClients `yaml:"grpc_clients"`
	HTTP_Server  HttpServer  `yaml:"http_server"`
	CORS         CORS        `yaml:"cors"`
}

type GrpcClients struct {
	AuthService         string `yaml:"auth_service_address"`
	UserService         string `yaml:"user_service_address"`
	MatcherService      string `yaml:"matcher_service_address"`
	FileStorageService  string `yaml:"file_storage_service_address"`
	ChatService         string `yaml:"chat_service_address"`
	NotificationService string `yaml:"notification_service_address"`
}

type HttpServer struct {
	Address      string        `yaml:"address"`
	Timeout      time.Duration `yaml:"timeout"`
	Idle_Timeout time.Duration `yaml:"idle_timeout"`
}

type CORS struct {
	AllowedOrigins   []string `yaml:"allowed_origins"`
	AllowedMethods   []string `yaml:"allowed_methods"`
	AllowedHeaders   []string `yaml:"allowed_headers"`
	AllowCredentials bool     `yaml:"allow_credentials"`
}
