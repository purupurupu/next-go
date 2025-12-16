package main

import (
	"fmt"
	"os"
	"time"

	"github.com/joho/godotenv"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"todo-api/internal/config"
	"todo-api/internal/model"
	"todo-api/pkg/database"
)

func main() {
	// Configure zerolog
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	// Load .env file
	if err := godotenv.Load(); err != nil {
		log.Debug().Msg("No .env file found")
	}

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to load configuration")
	}

	// Connect to database
	db, err := database.Connect(cfg.DatabaseURL)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to connect to database")
	}
	defer database.Close(db)

	log.Info().Msg("Starting database seed...")

	// Create test user
	user := model.User{
		Email: "test@example.com",
	}
	if err := user.SetPassword("password123"); err != nil {
		log.Fatal().Err(err).Msg("Failed to hash password")
	}

	// Check if user already exists
	var existingUser model.User
	if err := db.Where("email = ?", user.Email).First(&existingUser).Error; err == nil {
		log.Info().Str("email", user.Email).Msg("User already exists, using existing user")
		user = existingUser
	} else {
		if err := db.Create(&user).Error; err != nil {
			log.Fatal().Err(err).Msg("Failed to create user")
		}
		log.Info().Str("email", user.Email).Msg("Created user")
	}

	// Create categories
	categories := []model.Category{
		{UserID: user.ID, Name: "work", Color: "#3B82F6"},
		{UserID: user.ID, Name: "personal", Color: "#10B981"},
		{UserID: user.ID, Name: "shopping", Color: "#F59E0B"},
		{UserID: user.ID, Name: "health", Color: "#EF4444"},
	}

	for i := range categories {
		var existing model.Category
		if err := db.Where("user_id = ? AND name = ?", user.ID, categories[i].Name).First(&existing).Error; err == nil {
			categories[i] = existing
			log.Info().Str("name", categories[i].Name).Msg("Category already exists")
		} else {
			if err := db.Create(&categories[i]).Error; err != nil {
				log.Error().Err(err).Str("name", categories[i].Name).Msg("Failed to create category")
			} else {
				log.Info().Str("name", categories[i].Name).Msg("Created category")
			}
		}
	}

	// Create tags
	tags := []model.Tag{
		{UserID: user.ID, Name: "urgent", Color: strPtr("#EF4444")},
		{UserID: user.ID, Name: "important", Color: strPtr("#F59E0B")},
		{UserID: user.ID, Name: "later", Color: strPtr("#6B7280")},
		{UserID: user.ID, Name: "meeting", Color: strPtr("#8B5CF6")},
		{UserID: user.ID, Name: "review", Color: strPtr("#06B6D4")},
	}

	for i := range tags {
		var existing model.Tag
		if err := db.Where("user_id = ? AND name = ?", user.ID, tags[i].Name).First(&existing).Error; err == nil {
			tags[i] = existing
			log.Info().Str("name", tags[i].Name).Msg("Tag already exists")
		} else {
			if err := db.Create(&tags[i]).Error; err != nil {
				log.Error().Err(err).Str("name", tags[i].Name).Msg("Failed to create tag")
			} else {
				log.Info().Str("name", tags[i].Name).Msg("Created tag")
			}
		}
	}

	// Create todos
	tomorrow := time.Now().AddDate(0, 0, 1)
	nextWeek := time.Now().AddDate(0, 0, 7)
	yesterday := time.Now().AddDate(0, 0, -1)

	todos := []struct {
		todo model.Todo
		tags []string
	}{
		{
			todo: model.Todo{
				UserID:      user.ID,
				CategoryID:  &categories[0].ID, // work
				Title:       "プロジェクト企画書を作成",
				Description: strPtr("来週のミーティングに向けて企画書を準備する"),
				Priority:    model.PriorityHigh,
				Status:      model.StatusInProgress,
				DueDate:     &tomorrow,
			},
			tags: []string{"urgent", "important"},
		},
		{
			todo: model.Todo{
				UserID:      user.ID,
				CategoryID:  &categories[0].ID, // work
				Title:       "コードレビューを完了",
				Description: strPtr("PRのレビューを行う"),
				Priority:    model.PriorityMedium,
				Status:      model.StatusPending,
				DueDate:     &nextWeek,
			},
			tags: []string{"review"},
		},
		{
			todo: model.Todo{
				UserID:      user.ID,
				CategoryID:  &categories[1].ID, // personal
				Title:       "本を読む",
				Description: strPtr("積読を消化する"),
				Priority:    model.PriorityLow,
				Status:      model.StatusPending,
			},
			tags: []string{"later"},
		},
		{
			todo: model.Todo{
				UserID:      user.ID,
				CategoryID:  &categories[2].ID, // shopping
				Title:       "買い物リスト",
				Description: strPtr("牛乳、卵、パン"),
				Priority:    model.PriorityMedium,
				Status:      model.StatusPending,
				DueDate:     &tomorrow,
			},
			tags: []string{},
		},
		{
			todo: model.Todo{
				UserID:      user.ID,
				CategoryID:  &categories[3].ID, // health
				Title:       "ジムに行く",
				Description: strPtr("週3回を目標に"),
				Priority:    model.PriorityMedium,
				Status:      model.StatusPending,
			},
			tags: []string{"important"},
		},
		{
			todo: model.Todo{
				UserID:     user.ID,
				Title:      "チームミーティング",
				Priority:   model.PriorityHigh,
				Status:     model.StatusCompleted,
				Completed:  true,
				DueDate:    &yesterday,
			},
			tags: []string{"meeting"},
		},
	}

	for _, td := range todos {
		var existing model.Todo
		if err := db.Where("user_id = ? AND title = ?", user.ID, td.todo.Title).First(&existing).Error; err == nil {
			log.Info().Str("title", td.todo.Title).Msg("Todo already exists")
			continue
		}

		if err := db.Create(&td.todo).Error; err != nil {
			log.Error().Err(err).Str("title", td.todo.Title).Msg("Failed to create todo")
			continue
		}

		// Add tags
		for _, tagName := range td.tags {
			for _, tag := range tags {
				if tag.Name == tagName {
					todoTag := model.TodoTag{
						TodoID: td.todo.ID,
						TagID:  tag.ID,
					}
					if err := db.Create(&todoTag).Error; err != nil {
						log.Error().Err(err).Msg("Failed to create todo_tag")
					}
					break
				}
			}
		}

		// Update category todo count
		if td.todo.CategoryID != nil {
			db.Model(&model.Category{}).
				Where("id = ?", td.todo.CategoryID).
				UpdateColumn("todos_count", db.Raw("todos_count + 1"))
		}

		log.Info().Str("title", td.todo.Title).Msg("Created todo")
	}

	// Create notes
	notes := []model.Note{
		{
			UserID: user.ID,
			Title:  strPtr("開発メモ"),
			BodyMD: strPtr(`# 開発メモ

## やること
- [ ] APIの実装
- [ ] テストの追加
- [ ] ドキュメントの更新

## 参考リンク
- [Go公式ドキュメント](https://go.dev/doc/)
- [Echo Framework](https://echo.labstack.com/)
`),
			Pinned: true,
		},
		{
			UserID: user.ID,
			Title:  strPtr("ミーティングノート"),
			BodyMD: strPtr(`# 週次ミーティング

## 議題
1. 進捗報告
2. 課題の共有
3. 来週の予定

## メモ
- 新機能のリリースは来月予定
- パフォーマンス改善が必要
`),
			Pinned: false,
		},
		{
			UserID: user.ID,
			Title:  strPtr("アイデアメモ"),
			BodyMD: strPtr(`# アイデア

- ダークモード対応
- モバイルアプリ
- リマインダー機能
`),
			Pinned: false,
		},
	}

	for i := range notes {
		var existing model.Note
		if notes[i].Title != nil {
			if err := db.Where("user_id = ? AND title = ?", user.ID, *notes[i].Title).First(&existing).Error; err == nil {
				log.Info().Str("title", *notes[i].Title).Msg("Note already exists")
				continue
			}
		}

		notes[i].LastEditedAt = time.Now()
		if err := db.Create(&notes[i]).Error; err != nil {
			log.Error().Err(err).Msg("Failed to create note")
		} else {
			title := "無題"
			if notes[i].Title != nil {
				title = *notes[i].Title
			}
			log.Info().Str("title", title).Msg("Created note")
		}
	}

	fmt.Println("")
	log.Info().Msg("Seed completed successfully!")
	fmt.Println("")
	fmt.Println("テストユーザー:")
	fmt.Println("  Email:    test@example.com")
	fmt.Println("  Password: password123")
}

func strPtr(s string) *string {
	return &s
}
