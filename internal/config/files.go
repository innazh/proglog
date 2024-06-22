package config

import (
	"os"
	"path/filepath"
)

var (
	//issuing cert authority
	CAFile = configFile("ca.pem")

	//server
	ServerCertFile = configFile("server.pem")
	ServerKeyFile  = configFile("server-key.pem")

	//clients
	RootClientCertFile   = configFile("root-client.pem")
	RootClientKeyFile    = configFile("root-client-key.pem")
	NobodyClientCertFile = configFile("nobody-client.pem")
	NobodyClientKeyFile  = configFile("nobody-client-key.pem")

	//authorization config files
	ACLModelFile  = configFile("model.conf")
	ACLPolicyFile = configFile("policy.csv")
)

// configFile returns the full path of a config file
func configFile(filename string) string {
	if dir := os.Getenv("CONFIG_DIR"); dir != "" {
		return filepath.Join(dir, filename)
	}
	homeDir, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}
	return filepath.Join(homeDir, "projects", ".proglog", filename)
}
