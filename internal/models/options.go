package models

// Options for the CLI.
type Options struct {
	Debug      bool   `doc:"Enable debug logging" short:"d" default:"true"`
	Host       string `doc:"Hostname to listen on" default:"localhost"`
	Port       int    `doc:"Port to listen on" short:"p" default:"8880"`
	DBHost     string `name:"db-host" doc:"Database hostname" default:"localhost"`
	DBPort     int    `name:"db-port" doc:"Database port" default:"5432"`
	DBUser     string `name:"db-user" doc:"Database username" default:"postgres"`
	DBPassword string `name:"db-password" doc:"Database password" default:"password"`
	DBName     string `name:"db-name" doc:"Database name" default:"postgres"`
	AdminKey   string `name:"admin-key" doc:"Admin API key"`
}
