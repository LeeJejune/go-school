package main

import (
	"github.com/LeeJejune/go-school/go-practice/internal/app"
	"go.uber.org/fx"
)

func main() { fx.New(app.Gateway()).Run() }
