package main

import (
	"embed"

	"userstyles.world/handlers"
	"userstyles.world/modules/cache"
	database "userstyles.world/modules/database/init"
	"userstyles.world/modules/images"
	"userstyles.world/modules/log"
	"userstyles.world/search"
	"userstyles.world/services/cron"
	"userstyles.world/utils"
)

var (
	//go:embed views/*
	views embed.FS

	//go:embed static/*
	static embed.FS
)

func main() {
	log.Initialize()
	cache.Initialize()
	utils.InitalizeCrypto()
	utils.InitializeValidator()
	database.Initialize()
	cron.Initialize()
	search.Initialize()
	images.Initialize()
	handlers.Initialize(views, static)
}
