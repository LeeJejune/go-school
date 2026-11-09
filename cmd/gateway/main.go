package main

import (
	"github.com/LeeJejune/go-school/internal/app"
	"go.uber.org/fx"
)

func main() { fx.New(app.Gateway()).Run() }
