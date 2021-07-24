package worker

import "os"

type Config struct {
	WorkerGRPCAddr    string
	SchedulerGRPCAddr string
}

// TODO: load config from env/json/yaml/toml etcd.
func InitConfig() *Config {
	return &Config{
		WorkerGRPCAddr:    os.Getenv("WORKER_GRPC_ADDR"),
		SchedulerGRPCAddr: os.Getenv("SCHEDULER_GRPC_ADDR"),
	}
}
