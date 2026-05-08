package notify

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/JonathanVil/kultured/calc"
	"github.com/JonathanVil/kultured/models"
)

type Config struct {
	URL   string
	Topic string
	User  string
	Pass  string
}

func (c Config) Enabled() bool {
	return c.URL != "" && c.Topic != ""
}

func (c Config) send(message string) error {
	url := strings.TrimRight(c.URL, "/") + "/" + c.Topic
	req, err := http.NewRequest("POST", url, strings.NewReader(message))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "text/plain")
	if c.User != "" {
		req.SetBasicAuth(c.User, c.Pass)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		return fmt.Errorf("ntfy returned status %d", resp.StatusCode)
	}
	return nil
}

// StartScheduler launches a background goroutine that fires reminder
// notifications once per minute at the configured time.
func StartScheduler(db *sql.DB, cfg Config) {
	go func() {
		// Align to next minute boundary before entering the loop.
		now := time.Now()
		time.Sleep(time.Until(now.Truncate(time.Minute).Add(time.Minute)))
		for {
			checkAndSend(db, cfg)
			time.Sleep(time.Minute)
		}
	}()
}

func checkAndSend(db *sql.DB, cfg Config) {
	batches, err := models.GetBatchesWithReminders(db)
	if err != nil {
		log.Println("reminder: failed to fetch batches:", err)
		return
	}

	now := time.Now()
	nowHHMM := now.Format("15:04")

	for _, b := range batches {
		if b.ReminderTime != nowHHMM {
			continue
		}

		// For weekly reminders with a day-of-week set, skip other days.
		if b.ReminderIntervalDays == 7 && b.ReminderDayOfWeek.Valid {
			// DB stores Monday=0…Sunday=6; Go's Weekday() is Sunday=0…Saturday=6.
			todayMon := (int(now.Weekday()) + 6) % 7
			if todayMon != int(b.ReminderDayOfWeek.Int64) {
				continue
			}
		}

		if b.LastRemindedAt.Valid {
			last, err := time.Parse("2006-01-02T15:04:05Z", b.LastRemindedAt.String)
			if err == nil {
				// Compare calendar dates so the interval is in whole days.
				lastDate := last.UTC().Truncate(24 * time.Hour)
				todayDate := now.UTC().Truncate(24 * time.Hour)
				daysSince := int(todayDate.Sub(lastDate).Hours() / 24)
				if daysSince < b.ReminderIntervalDays {
					continue
				}
			}
		}

		day := currentDay(b)
		msg := fmt.Sprintf("%s in %s is on day %d", b.Name, stageLabel(b.Stage), day)

		if err := cfg.send(msg); err != nil {
			log.Printf("reminder: failed to send for batch %d: %v", b.ID, err)
			continue
		}
		log.Printf("reminder: sent for batch %d (%s)", b.ID, b.Name)

		ts := now.UTC().Format("2006-01-02T15:04:05Z")
		if err := models.UpdateReminderSent(db, b.ID, ts); err != nil {
			log.Printf("reminder: failed to update last_reminded_at for batch %d: %v", b.ID, err)
		}
	}
}

func currentDay(b models.Batch) int {
	if b.StartF2.Valid {
		if b.DoneAt.Valid {
			d, _ := calc.DaysBetween(b.StartF2.String, b.DoneAt.String)
			return d
		}
		d, _ := calc.DaysSince(b.StartF2.String)
		return d
	}
	d, _ := calc.FermentationDays(b.StartedAt)
	return d
}

func stageLabel(stage string) string {
	switch stage {
	case "f1":
		return "F1"
	case "f2":
		return "F2"
	case "bottled":
		return "Bottled"
	default:
		return stage
	}
}
