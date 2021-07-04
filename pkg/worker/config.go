package worker

import "os"

type Config struct {
	WorkerGRPCAddr string
}

// TODO: load config from env/json/yaml/toml etcd.
func InitConfig() *Config {
	return &Config{
		WorkerGRPCAddr: os.Getenv("WORKER_GRPC_ADDR"),
	}
}
