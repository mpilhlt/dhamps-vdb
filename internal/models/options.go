package models

// Options for the CLI.
type Options struct {
  Debug bool        `doc:"Enable debug logging" short:"d" default:"true"`
  Host  string      `doc:"Hostname to listen on" default:"localhost"`
  Port  int         `doc:"Port to listen on" short:"p" default:"8888"`
  DBHost string     `doc:"Database hostname" default:"localhost"`
  DBPort int        `doc:"Database port" default:"5432"`
  DBUser string     `doc:"Database username" default:"postgres"`
  DBPassword string `doc:"Database password" default:"password"`
  DBName string     `doc:"Database name" default:"postgres"`
  AdminKey string   `doc:"Admin API key"`
}
