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

	bot, err := tgbotapi.NewBotAPI("8902538391:AAFVBZTqLpfZ_FD9nM2Vvk5xsGTjfJIZXZc")
	if err != nil {
		log.Fatal(err)
	}

	updateConfig := tgbotapi.NewUpdate(0)
	updateConfig.Timeout = 60

	updates := bot.GetUpdatesChan(updateConfig)
	selectedWeekByChat := make(map[int64]string)

	for update := range updates {
		if update.Message == nil {
			continue
		}

		chatID := update.Message.Chat.ID
		text := strings.ToLower(strings.TrimSpace(update.Message.Text))

		switch text {
		case "/start", "выбрать неделю", "/выбрать_неделю":
			delete(selectedWeekByChat, chatID)
			sendWeekSelection(bot, chatID, "Выбери неделю:")

		default:
			if week, ok := parseWeek(text); ok {
				selectedWeekByChat[chatID] = week
				sendDaySelection(bot, chatID, "Выбери день недели:")
				continue
			}

			day, ok := parseDay(text)
			if !ok {
				sendWeekSelection(bot, chatID, "Неизвестная команда. Выбери неделю:")
				continue
			}

			week, ok := selectedWeekByChat[chatID]
			if !ok {
				sendWeekSelection(bot, chatID, "Сначала выбери четную или нечетную неделю:")
				continue
			}

			sendSchedule(ctx, bot, conn, chatID, day, week)
		}
	}
}

func parseWeek(text string) (string, bool) {
	switch text {
	case "/четная", "четная", "четная неделя":
		return "четная", true
	case "/нечетная", "нечетная", "нечетная неделя", "нечётная", "нечётная неделя":
		return "нечетная", true
	default:
		return "", false
	}
}

func parseDay(text string) (string, bool) {
	switch text {
	case "/понедельник", "понедельник":
		return "понедельник", true
	case "/вторник", "вторник":
		return "вторник", true
	case "/среда", "среда":
		return "среда", true
	case "/четверг", "четверг":
		return "четверг", true
	case "/пятница", "пятница":
		return "пятница", true
	case "/суббота", "суббота":
		return "суббота", true
	case "/воскресенье", "воскресенье":
		return "воскресенье", true
	default:
		return "", false
	}
}

func sendSchedule(ctx context.Context, bot *tgbotapi.BotAPI, conn *pgx.Conn, chatID int64, day string, week string) {
	answer, err := GetSchedule(ctx, conn, day, week)
	if err != nil {
		sendDaySelection(bot, chatID, "Ошибка при получении расписания")
		return
	}

	sendDaySelection(bot, chatID, answer)
}

func GetSchedule(ctx context.Context, conn *pgx.Conn, day string, week string) (string, error) {
	rows, err := conn.Query(ctx, `
		SELECT time, subject, teacher, room
		FROM schedule
		WHERE day = $1 AND week_type = $2
		ORDER BY time
	`, day, week)
	if err != nil {
		return "", err
	}
	defer rows.Close()

	var result strings.Builder
	result.WriteString("Расписание на " + day + " (" + week + " неделя):\n\n")

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
		return "На " + day + " (" + week + " неделя) расписания нет", nil
	}

	return result.String(), rows.Err()
}

func sendWeekSelection(bot *tgbotapi.BotAPI, chatID int64, text string) {
	sendWithKeyboard(bot, chatID, text, getWeekKeyboard())
}

func sendDaySelection(bot *tgbotapi.BotAPI, chatID int64, text string) {
	sendWithKeyboard(bot, chatID, text, getDayKeyboard())
}

func sendWithKeyboard(bot *tgbotapi.BotAPI, chatID int64, text string, keyboard tgbotapi.ReplyKeyboardMarkup) {
	msg := tgbotapi.NewMessage(chatID, text)
	msg.ReplyMarkup = keyboard

	_, err := bot.Send(msg)
	if err != nil {
		log.Println(err)
	}
}

func getWeekKeyboard() tgbotapi.ReplyKeyboardMarkup {
	keyboard := tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("Четная неделя"),
			tgbotapi.NewKeyboardButton("Нечетная неделя"),
		),
	)

	keyboard.ResizeKeyboard = true
	return keyboard
}

func getDayKeyboard() tgbotapi.ReplyKeyboardMarkup {
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
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("Выбрать неделю"),
		),
	)

	keyboard.ResizeKeyboard = true
	return keyboard
}
