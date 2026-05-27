package main

import (
	"context"
	"fmt"
	"log"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/jackc/pgx/v5"
)

func main() {
	ctx := context.Background()

	conn, err := pgx.Connect(ctx, "postgres://andruhablatnoy:1234@localhost:5432/postgres")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close(ctx)

	bot, err := tgbotapi.NewBotAPI("8902538391:AAHgbmZ6G38eg5gF4mmhByG2hjDyIhRAJ-M")
	if err != nil {
		log.Fatal(err)
	}

	updateConfig := tgbotapi.NewUpdate(0)
	updateConfig.Timeout = 60

	updates := bot.GetUpdatesChan(updateConfig)

	for update := range updates {
		if update.Message == nil {
			continue
		}

		chatID := update.Message.Chat.ID
		text := strings.ToLower(update.Message.Text)

		switch text {
		case "/start":
			sendWithKeyboard(bot, chatID, "Выбери день недели:")

		case "/понедельник", "понедельник":
			sendSchedule(ctx, bot, conn, chatID, "понедельник")

		case "/вторник", "вторник":
			sendSchedule(ctx, bot, conn, chatID, "вторник")

		case "/среда", "среда":
			sendSchedule(ctx, bot, conn, chatID, "среда")

		case "/четверг", "четверг":
			sendSchedule(ctx, bot, conn, chatID, "четверг")

		case "/пятница", "пятница":
			sendSchedule(ctx, bot, conn, chatID, "пятница")

		case "/суббота", "суббота":
			sendSchedule(ctx, bot, conn, chatID, "суббота")

		case "/воскресенье", "воскресенье":
			sendSchedule(ctx, bot, conn, chatID, "воскресенье")

		default:
			sendWithKeyboard(bot, chatID, "Неизвестная команда. Выбери день недели:")
		}
	}
}

func sendSchedule(ctx context.Context, bot *tgbotapi.BotAPI, conn *pgx.Conn, chatID int64, day string) {
	answer, err := GetScheduleByDay(ctx, conn, day)
	if err != nil {
		sendWithKeyboard(bot, chatID, "Ошибка при получении расписания")
		return
	}

	sendWithKeyboard(bot, chatID, answer)
}

func GetScheduleByDay(ctx context.Context, conn *pgx.Conn, day string) (string, error) {
	rows, err := conn.Query(ctx, `
		SELECT time, subject, teacher, room
		FROM schedule
		WHERE day = $1
		ORDER BY time
	`, day)
	if err != nil {
		return "", err
	}
	defer rows.Close()

	var result strings.Builder
	result.WriteString("Расписание на " + day + ":\n\n")

	found := false

	for rows.Next() {
		found = true

		var time string
		var subject string
		var teacher *string
		var room *string

		err := rows.Scan(&time, &subject, &teacher, &room)
		if err != nil {
			return "", err
		}

		result.WriteString(fmt.Sprintf("%s — %s", time, subject))

		if teacher != nil {
			result.WriteString(fmt.Sprintf(", %s", *teacher))
		}

		if room != nil {
			result.WriteString(fmt.Sprintf(", кабинет %s", *room))
		}

		result.WriteString("\n")
	}

	if !found {
		return "На " + day + " расписания нет", nil
	}

	return result.String(), rows.Err()
}

func sendWithKeyboard(bot *tgbotapi.BotAPI, chatID int64, text string) {
	msg := tgbotapi.NewMessage(chatID, text)
	msg.ReplyMarkup = getMainKeyboard()

	_, err := bot.Send(msg)
	if err != nil {
		log.Println(err)
	}
}

func getMainKeyboard() tgbotapi.ReplyKeyboardMarkup {
	keyboard := tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("Понедельник"),
			tgbotapi.NewKeyboardButton("Вторник"),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("Среда"),
			tgbotapi.NewKeyboardButton("Четверг"),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("Пятница"),
			tgbotapi.NewKeyboardButton("Суббота"),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("Воскресенье"),
		),
	)

	keyboard.ResizeKeyboard = true
	return keyboard
}
