package config

import "time"

type Config struct {
	Env          string      `yaml:"env" env-default:"dev"`
	PublicKey    string      `env:"PUBLIC_KEY" env-required:"true"`
	GRPC_Clients GrpcClients `yaml:"grpc_clients"`
	HTTP_Server  HttpServer  `yaml:"http_server"`
}

type GrpcClients struct {
	AuthService string `yaml:"auth_service_address"`
}

type HttpServer struct {
	Address      string        `yaml:"address"`
	Timeout      time.Duration `yaml:"timeout"`
	Idle_Timeout time.Duration `yaml:"idle_timeout"`
}
