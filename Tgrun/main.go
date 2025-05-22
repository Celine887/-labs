package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	_ "github.com/mattn/go-sqlite3"
)

var db *sql.DB

func initDB() {
	var err error
	db, err = sql.Open("sqlite3", "./runs.db")
	if err != nil {
		log.Fatal(err)
	}

	createTable := `
	CREATE TABLE IF NOT EXISTS runs (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id INTEGER,
		distance REAL,
		duration INTEGER,
		date DATETIME DEFAULT CURRENT_TIMESTAMP
	);`

	_, err = db.Exec(createTable)
	if err != nil {
		log.Fatal(err)
	}
}

func main() {
	bot, err := tgbotapi.NewBotAPI(os.Getenv("TELEGRAM_BOT_TOKEN"))
	if err != nil {
		log.Fatal(err)
	}

	initDB()

	bot.Debug = true
	log.Printf("Authorized on account %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message == nil {
			continue
		}

		if !update.Message.IsCommand() {
			continue
		}

		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "")

		switch update.Message.Command() {
		case "start":
			msg.Text = "Привет! Я бот для трекинга пробежек. Используйте следующие команды:\n" +
				"/run <расстояние> <время> - записать пробежку (например: /run 5 30 - 5 км за 30 минут)\n" +
				"/stats - показать статистику\n" +
				"/help - показать это сообщение"
		case "run":
			args := strings.Fields(update.Message.CommandArguments())
			if len(args) != 2 {
				msg.Text = "Используйте формат: /run <расстояние> <время>"
				break
			}

			distance, err := strconv.ParseFloat(args[0], 64)
			if err != nil {
				msg.Text = "Неверный формат расстояния"
				break
			}

			duration, err := strconv.Atoi(args[1])
			if err != nil {
				msg.Text = "Неверный формат времени"
				break
			}

			_, err = db.Exec("INSERT INTO runs (user_id, distance, duration) VALUES (?, ?, ?)",
				update.Message.From.ID, distance, duration)
			if err != nil {
				log.Printf("Error saving run: %v", err)
				msg.Text = "Произошла ошибка при сохранении пробежки"
				break
			}

			msg.Text = fmt.Sprintf("Пробежка записана: %.1f км за %d минут", distance, duration)

		case "stats":
			var totalDistance float64
			var totalDuration int
			var count int

			err := db.QueryRow("SELECT COUNT(*), SUM(distance), SUM(duration) FROM runs WHERE user_id = ?",
				update.Message.From.ID).Scan(&count, &totalDistance, &totalDuration)
			if err != nil {
				log.Printf("Error getting stats: %v", err)
				msg.Text = "Произошла ошибка при получении статистики"
				break
			}

			if count == 0 {
				msg.Text = "У вас пока нет записанных пробежек"
				break
			}

			avgPace := float64(totalDuration) / totalDistance
			msg.Text = fmt.Sprintf("Статистика пробежек:\n"+
				"Всего пробежек: %d\n"+
				"Общее расстояние: %.1f км\n"+
				"Общее время: %d минут\n"+
				"Средний темп: %.1f мин/км",
				count, totalDistance, totalDuration, avgPace)

		case "help":
			msg.Text = "Доступные команды:\n" +
				"/run <расстояние> <время> - записать пробежку\n" +
				"/stats - показать статистику\n" +
				"/help - показать это сообщение"
		}

		if _, err := bot.Send(msg); err != nil {
			log.Printf("Error sending message: %v", err)
		}
	}
}
