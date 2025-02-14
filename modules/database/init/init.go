package init

import (
	l "log"
	"os"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"userstyles.world/models"
	"userstyles.world/modules/config"
	"userstyles.world/modules/database"
	"userstyles.world/modules/log"
	"userstyles.world/utils"
)

var tables = []struct {
	name  string
	model interface{}
}{
	{"users", &models.User{}},
	{"styles", &models.Style{}},
	{"stats", &models.Stats{}},
	{"oauths", &models.OAuth{}},
	{"histories", &models.History{}},
	{"logs", &models.Log{}},
	{"reviews", &models.Review{}},
	{"notifications", &models.Notification{}},
}

func connect() (*gorm.DB, error) {
	newLogger := logger.New(
		l.New(os.Stdout, "\r\n", l.LstdFlags),
		logger.Config{
			SlowThreshold: time.Second,
			LogLevel:      logLevel(),
			Colorful:      config.DBColor,
		},
	)

	conn, err := gorm.Open(sqlite.Open(config.DB), &gorm.Config{
		Logger: newLogger,
	})
	if err != nil {
		return nil, err
	}

	return conn, nil
}

// Initialize the database connection.
func Initialize() {
	conn, err := connect()
	if err != nil {
		log.Warn.Fatal("Failed to connect database:", err.Error())
	}

	database.Conn = conn
	log.Info.Println("Database successfully connected.")

	shouldSeed := false
	// Generate data for development.
	if config.DBDrop && !config.Production {
		for _, table := range tables {
			if err := drop(table.model); err != nil {
				log.Warn.Fatalf("Failed to drop %s, err: %s\n", table.name, err.Error())
			}
			log.Info.Println("Dropped database table", table.name)
		}
		shouldSeed = true
	}

	if _, ok := os.LookupEnv("MAGIC"); ok {
		log.Info.Println("Started stats table migration!")
		migrateStats()
		log.Info.Println("Finished stats table migration!")
	}

	// Migrate tables.
	if config.DBMigrate {
		for _, table := range tables {
			if err := migrate(table.model); err != nil {
				log.Warn.Fatalf("Failed to migrate %s, err: %s\n", table.name, err.Error())
			}

			log.Info.Println("Migrated database table", table.name)
		}
	}

	if shouldSeed {
		seed()
	}
}

func migrateStats() {
	var err error

	log.Info.Println("Dropping stats indices")
	stmt := `
DROP INDEX IF EXISTS idx_stats_weekly_installs;
DROP INDEX IF EXISTS idx_stats_deleted_at;
DROP INDEX IF EXISTS idx_stats_installed;
DROP INDEX IF EXISTS idx_stats_style_id;
DROP INDEX IF EXISTS idx_stats_viewed;
DROP INDEX IF EXISTS idx_stats;`
	if err = database.Conn.Exec(stmt).Error; err != nil {
		log.Warn.Fatalln("Failed to drop stats indices:", err)
	}

	log.Info.Println("Renaming stats to stats_old")
	stmt = `ALTER TABLE stats RENAME TO stats_old;`
	if err = database.Conn.Exec(stmt).Error; err != nil {
		log.Warn.Fatalln("Failed to rename stats to stats_old:", err)
	}

	log.Info.Println("Creating new stats table")
	if err = migrate(models.Stats{}); err != nil {
		log.Warn.Fatalln("Failed to create new stats table:", err)
	}

	log.Info.Println("Inserting original data")
	stmt = `INSERT INTO stats(id, created_at, updated_at, deleted_at, hash, install, style_id, view) SELECT * FROM stats_old;`
	if err = database.Conn.Exec(stmt).Error; err != nil {
		log.Warn.Fatalln("Failed to insert original data:", err)
	}

	log.Info.Println("Dropping stats_old table")
	stmt = `DROP TABLE stats_old;`
	if err = database.Conn.Exec(stmt).Error; err != nil {
		log.Warn.Fatalln("Failed to drop stats_old table:", err)
	}

	log.Info.Println("Updating columns with no data")
	stmt = `
UPDATE stats SET install = NULL WHERE install = 0;
UPDATE stats SET view = NULL WHERE view = 0;`
	if err = database.Conn.Exec(stmt).Error; err != nil {
		log.Warn.Fatalln("Failed to update columns with no data:", err)
	}

	log.Info.Println("Cleaning up database file")
	stmt = `VACUUM;`
	if err = database.Conn.Exec(stmt).Error; err != nil {
		log.Warn.Fatalln("Failed to clean up database file:", err)
	}
}

func generateData(amount int) ([]models.Style, []models.User) {
	randomData := utils.RandomString(amount * 7 * 4)
	var styleStructs []models.Style
	for i := 0; i < amount; i++ {
		startData := randomData[(i * 7 * 4):]
		styleStructs = append(styleStructs, models.Style{
			UserID:      uint(amount),
			Category:    startData[:4],
			Name:        startData[4:8],
			Description: startData[8:12],
			Notes:       startData[12:16],
			Preview:     startData[16:20],
			Code:        startData[20:24],
			Homepage:    startData[24:28],
			Featured:    true,
		})
	}

	var userStructs []models.User
	randomData = utils.RandomString(amount * 4 * 4)
	for i := 0; i < amount; i++ {
		startData := randomData[(i * 4 * 4):]
		userStructs = append(userStructs, models.User{
			Username:  startData[:4],
			Email:     startData[4:8],
			Biography: startData[8:12],
			Password:  startData[12:16],
		})
	}

	return styleStructs, userStructs
}

func seed() {
	log.Info.Println("Seeding database mock data.")
	defer log.Info.Println("Finished seeding mock data.")

	users := []models.User{
		{
			Username:  "admin",
			Email:     "admin@usw.local",
			Biography: "Admin of USw.",
			Password:  utils.GenerateHashedPassword("admin123"),
			Role:      models.Admin,
		},
		{
			Username:  "moderator",
			Email:     "moderator@usw.local",
			Biography: "I'm a moderator.",
			Password:  utils.GenerateHashedPassword("moderator"),
		},
		{
			Username: "regular",
			Email:    "regular@usw.local",
			Password: utils.GenerateHashedPassword("regular"),
		},
	}

	styles := []models.Style{
		{
			UserID:      1,
			Name:        "Dark-GitHub",
			Description: "Customizable dark theme for GitHub.",
			Notes:       "Some notes go here.",
			Preview:     "https://userstyles.world/api/style/preview/2.webp",
			Original:    "https://raw.githubusercontent.com/vednoc/dark-github/main/github.user.styl",
			Homepage:    "https://github.com/vednoc/dark-github",
			Category:    "github.com",
			MirrorCode:  true,
			Featured:    true,
		},
		{
			UserID:      1,
			Name:        "Dark-GitLab",
			Description: "Customizable dark theme for GitLab.",
			Notes:       "Some notes go here.",
			Preview:     "https://userstyles.world/api/style/preview/3.webp",
			Original:    "https://gitlab.com/vednoc/dark-gitlab/raw/master/gitlab.user.styl",
			Homepage:    "https://gitlab.com/vednoc/dark-gitlab",
			Category:    "gitlab.com",
			MirrorCode:  true,
			Featured:    true,
		},
		{
			UserID:      1,
			Name:        "Dark-WhatsApp",
			Description: "Customizable dark theme for WhatsApp.",
			Notes:       "Some notes go here.",
			Preview:     "https://userstyles.world/api/style/preview/4.webp",
			Original:    "https://raw.githubusercontent.com/vednoc/dark-whatsapp/master/wa.user.styl",
			Homepage:    "https://github.com/vednoc/dark-whatsapp",
			Category:    "web.whatsapp.com",
			MirrorCode:  true,
			Featured:    true,
		},
		{
			UserID:   2,
			Name:     "Archived userstyle",
			Archived: true,
		},
		{
			UserID:   3,
			Name:     "Featured userstyle",
			Featured: true,
		},
		{
			UserID: 3,
			Name:   "Temporary userstyle",
		},
	}

	oauths := []models.OAuth{
		{
			UserID:       1,
			Name:         "USw integration",
			Description:  "Just some integration",
			Scopes:       []string{"user", "style"},
			ClientID:     "publicccc_client",
			ClientSecret: "secreettUwU",
			RedirectURI:  "https://gusted.xyz/callback_helper",
		},
	}

	logs := []models.Log{
		{
			UserID:         1,
			Reason:         "I like to abuse powers.",
			Kind:           models.LogBanUser,
			TargetUserName: "gusted",
		},
		{
			UserID:         1,
			Reason:         "My style is superior",
			Kind:           models.LogRemoveStyle,
			TargetUserName: "gusted",
			TargetData:     "Black-Discord",
		},
	}

	if config.DBRandomData {
		s, u := generateData(config.DBRandomDataAmount)
		styles = append(styles, s...)
		users = append(users, u...)
	}

	for i := range users {
		database.Conn.Create(&users[i])
	}
	for i := range styles {
		database.Conn.Create(&styles[i])
	}
	for i := range oauths {
		database.Conn.Create(&oauths[i])
	}
	for i := range logs {
		database.Conn.Create(&logs[i])
	}
}
