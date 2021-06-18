package snapshot

import (
	"log"
	"time"

	"userstyles.world/models"
	"userstyles.world/modules/database"
)

func getViews(id int64) (i int64) {
	day := time.Now().AddDate(0, 0, -1)
	database.Conn.
		Model(models.Stats{}).
		Where("style_id = ? and created_at > ? and view > ?", id, day, day).
		Count(&i)

	return i
}

func getUpdates(id int64) (i int64) {
	day := time.Now().AddDate(0, 0, -1)
	q := "style_id = ? and install > ? and updated_at > ?"
	database.Conn.
		Model(models.Stats{}).
		Where(q, id, day, day).
		Count(&i)

	return i
}

func getInstalls(id int64) (i int64) {
	day := time.Now().AddDate(0, 0, -1)
	q := "style_id = ? and install > ? and created_at > ?"
	database.Conn.
		Model(models.Stats{}).
		Where(q, id, day, day).
		Count(&i)

	return i
}

func StyleStatistics() {
	styles, err := models.GetAllStyleIDs()
	if err != nil {
		log.Fatalln(err)
	}

	// Store style stats.
	stats := new([]models.History)

	// Iterate over styles and collect their stats.
	for _, v := range styles {
		item := models.History{
			StyleID:       v.ID,
			DailyViews:    getViews(int64(v.ID)),
			DailyInstalls: getInstalls(int64(v.ID)),
			DailyUpdates:  getUpdates(int64(v.ID)),
		}

		*stats = append(*stats, item)
	}

	log.Println("Stats history.")
	database.Conn.Debug().Create(stats)
}
