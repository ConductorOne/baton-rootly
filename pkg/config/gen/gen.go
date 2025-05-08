package main

import (
	cfg "github.com/conductorone/baton-rootly/pkg/config"
	"github.com/conductorone/baton-sdk/pkg/config"
)

func main() {
	config.Generate("rootly", cfg.Config)
}
