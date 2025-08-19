package services

import (
	"context"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"friendship/db"
	"friendship/models"
	"friendship/models/sessions"
	statsusers "friendship/models/stats_users"
)

func UpdateStatisticsForFinishedSession(ctx context.Context, sessionID uint) error {
	tx := db.GetDB().WithContext(ctx).Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 1) Узнаём ID статуса "Завершена"
	var st sessions.Status
	if err := tx.Where("status = ?", "Завершена").First(&st).Error; err != nil {
		tx.Rollback()
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("status 'Завершена' not found")
		}
		return err
	}

	// 2) Берём сессию с блокировкой
	var s sessions.Session
	if err := tx.
		Clauses(clause.Locking{Strength: "UPDATE"}).
		Preload("SessionType").
		First(&s, sessionID).Error; err != nil {
		tx.Rollback()
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("session %d not found", sessionID)
		}
		return err
	}

	// 3) Проверки статуса/времени
	now := time.Now()
	if s.StatusID != st.ID {
		tx.Rollback()
		return fmt.Errorf("session %d is not completed", sessionID)
	}
	if s.EndTime.After(now) {
		tx.Rollback()
		return fmt.Errorf("session %d end_time is in the future", sessionID)
	}

	// 4) Идемпотентность:
	// Если строка уже есть — RowsAffected == 0 => ничего не делаем.
	ipe := models.StatsProcessedEvent{
		SessionID: sessionID,
	}

	res := tx.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "session_id"}},
		DoNothing: true,
	}).Create(&ipe)

	if res.Error != nil {
		tx.Rollback()
		return fmt.Errorf("failed to create StatsProcessedEvent: %w", res.Error)
	}

	if res.RowsAffected == 0 {
		return tx.Commit().Error
	}

	// 5) Создатель: только CountCreateSession и MostBigSession
	if err := ensureUserStatsRows(tx, s.UserID); err != nil {
		tx.Rollback()
		return err
	}

	currentUsers := uint16(0)
	if s.CurrentUsers != 0 {
		currentUsers = s.CurrentUsers
	}

	if err := tx.Model(&statsusers.SideStats_users{}).
		Where("user_id = ?", s.UserID).
		Updates(map[string]any{
			"count_create_session": gorm.Expr("count_create_session + 1"),
			"most_big_session":     gorm.Expr("GREATEST(most_big_session, ?)", currentUsers),
		}).Error; err != nil {
		tx.Rollback()
		return err
	}

	// 6) Список пользователей для «участниковой» статистики
	userIDs := make(map[uint]struct{}, 8)

	var joins []sessions.SessionUser
	if err := tx.Where("session_id = ?", s.ID).Find(&joins).Error; err != nil {
		tx.Rollback()
		return err
	}
	for _, j := range joins {
		userIDs[j.UserID] = struct{}{}
	}

	userIDs[s.UserID] = struct{}{}

	// 7) Инкременты по участию для всех: count_all и по типам
	sessionDuration := uint64(0)
	if s.Duration > 0 {
		sessionDuration = uint64(s.Duration)
	} else {
		sessionDuration = uint64(s.EndTime.Sub(s.StartTime).Seconds())
	}

	for uid := range userIDs {
		if err := ensureUserStatsRows(tx, uid); err != nil {
			tx.Rollback()
			return err
		}
		if err := tx.Model(&statsusers.SessionStats_users{}).
			Where("user_id = ?", uid).
			Updates(map[string]any{
				"count_all":  gorm.Expr("count_all + 1"),
				"spent_time": gorm.Expr("spent_time + ?", sessionDuration),
			}).Error; err != nil {
			tx.Rollback()
			return err
		}

		sessionTypeID := uint(0)
		if s.SessionTypeID != 0 {
			sessionTypeID = s.SessionTypeID
		}

		if err := incrementTypeCounter(tx, uid, sessionTypeID); err != nil {
			tx.Rollback()
			return err
		}
	}

	// 8) Жанры из Mongo — привяжем всем пользователям (если есть)
	if err := attachGenresFromMongo(ctx, tx, s.ID, keys(userIDs)); err != nil {
		tx.Rollback()
		return err
	}

	// 9) Пересчёт производных метрик по участию: Streak, MostPopDay, PopSessionType (топ-тип)
	for uid := range userIDs {
		streak, err := computeLongestAttendanceStreak(tx, uid)
		if err != nil {
			tx.Rollback()
			return err
		}

		popDay, err := computeMostPopularDayByAttendance(tx, uid)
		if err != nil {
			tx.Rollback()
			return err
		}

		updates := map[string]any{
			"series_sesion_count": streak,
		}

		// Если popDay не nil, обновляем most_pop_day
		if popDay != nil {
			updates["most_pop_day"] = *popDay
		} else {
			// Явно устанавливаем NULL, если нет данных
			updates["most_pop_day"] = nil
		}

		if err := tx.Model(&statsusers.SideStats_users{}).
			Where("user_id = ?", uid).
			Updates(updates).Error; err != nil {
			tx.Rollback()
			return err
		}

		// Обновляем топ тип сессий (категорию) по посещаемости
		if err := upsertTopSessionTypeByAttendance(tx, uid); err != nil {
			tx.Rollback()
			return err
		}

		// Подсчитываем статистику по категориям/типам сессий
		if err := updateCategoryStats(tx, uid); err != nil {
			tx.Rollback()
			return err
		}
	}

	return tx.Commit().Error
}

// === Всякая штука для приколов =================================================================

func ensureUserStatsRows(tx *gorm.DB, userID uint) error {
	// Проверяем существование записи SideStats_users
	var existingSideStats statsusers.SideStats_users
	err := tx.Where("user_id = ?", userID).First(&existingSideStats).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			userIDPtr := userID
			countCreateSession := uint16(0)
			seriesSesionCount := uint16(0)
			mostBigSession := uint16(0)

			newSideStats := statsusers.SideStats_users{
				UserID:             &userIDPtr,
				CountCreateSession: &countCreateSession,
				SeriesSesionCount:  &seriesSesionCount,
				MostPopDay:         nil, // явно nil для NULL в БД
				MostBigSession:     &mostBigSession,
			}
			if err := tx.Create(&newSideStats).Error; err != nil {
				return fmt.Errorf("failed to create SideStats_users: %w", err)
			}
		} else {
			return fmt.Errorf("failed to query SideStats_users: %w", err)
		}
	}

	var existingSessionStats statsusers.SessionStats_users
	err = tx.Where("user_id = ?", userID).First(&existingSessionStats).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			newSessionStats := statsusers.SessionStats_users{
				UserID:          userID,
				CountAll:        0,
				CountFilms:      0,
				CountGames:      0,
				CountTableGames: 0,
				CountAnother:    0,
				SpentTime:       0,
			}
			if err := tx.Create(&newSessionStats).Error; err != nil {
				return fmt.Errorf("failed to create SessionStats_users: %w", err)
			}
		} else {
			return fmt.Errorf("failed to query SessionStats_users: %w", err)
		}
	}

	return nil
}

func incrementTypeCounter(tx *gorm.DB, userID, sessionTypeID uint) error {
	switch sessionTypeID {
	case 1:
		return tx.Model(&statsusers.SessionStats_users{}).
			Where("user_id = ?", userID).
			UpdateColumn("count_films", gorm.Expr("count_films + 1")).Error
	case 2:
		return tx.Model(&statsusers.SessionStats_users{}).
			Where("user_id = ?", userID).
			UpdateColumn("count_games", gorm.Expr("count_games + 1")).Error
	case 3:
		return tx.Model(&statsusers.SessionStats_users{}).
			Where("user_id = ?", userID).
			UpdateColumn("count_table_games", gorm.Expr("count_table_games + 1")).Error
	default:
		return tx.Model(&statsusers.SessionStats_users{}).
			Where("user_id = ?", userID).
			UpdateColumn("count_another", gorm.Expr("count_another + 1")).Error
	}
}

func computeLongestAttendanceStreak(tx *gorm.DB, userID uint) (uint16, error) {
	type drow struct{ D time.Time }
	var dates []drow

	err := tx.
		Raw(`
			select date_trunc('day', s.start_time) as d
			from session_users su
			join sessions s on s.id = su.session_id
			join statuses st on st.id = s.status_id
			where su.user_id = ?
			  and st.status = 'Завершена'
			group by 1
			order by 1 asc
		`, userID).Scan(&dates).Error
	if err != nil {
		return 0, err
	}

	var best, cur int
	var prev *time.Time
	for _, r := range dates {
		d := r.D.UTC()
		if prev == nil {
			cur, best = 1, 1
			prev = &d
			continue
		}
		diffDays := int(d.Sub(prev.UTC()).Hours() / 24)
		if diffDays == 1 {
			cur++
		} else if diffDays >= 0 {
			cur = 1
		}
		if cur > best {
			best = cur
		}
		*prev = d
	}
	return uint16(best), nil
}

func computeMostPopularDayByAttendance(tx *gorm.DB, userID uint) (*uint16, error) {
	type row struct{ D, C int }
	var rows []row
	err := tx.
		Raw(`
            select extract(isodow from s.start_time)::int as d, count(*) as c
            from session_users su
            join sessions s on s.id = su.session_id
            join statuses st on st.id = s.status_id
            where su.user_id = ?
              and st.status = 'Завершена'
            group by 1
        `, userID).Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return nil, nil
	}
	bestD, bestC := 1, -1
	for _, r := range rows {
		if r.C > bestC || (r.C == bestC && r.D < bestD) {
			bestD, bestC = r.D, r.C
		}
	}
	result := uint16(bestD)
	return &result, nil
}

func upsertTopSessionTypeByAttendance(tx *gorm.DB, userID uint) error {
	type row struct {
		TID uint
		C   int
	}
	var top row

	if err := tx.
		Raw(`
            select s.session_type_id as tid, count(*) as c
            from session_users su
            join sessions s on s.id = su.session_id
            join statuses st on st.id = s.status_id
            where su.user_id = ?
              and st.status = 'Завершена'
              and s.session_type_id IS NOT NULL
            group by 1
            order by c desc, tid asc
            limit 1
        `, userID).
		Scan(&top).Error; err != nil {
		return fmt.Errorf("failed to get top session type: %w", err)
	}

	if top.TID == 0 {
		return nil
	}

	rec := statsusers.PopSessionType{
		UserID:        userID,
		SessionTypeID: top.TID,
	}

	err := tx.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "user_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"session_type_id"}),
	}).Create(&rec).Error

	if err != nil {
		return fmt.Errorf("failed to upsert PopSessionType: %w", err)
	}

	return nil
}

func attachGenresFromMongo(ctx context.Context, tx *gorm.DB, sessionID uint, userIDs []uint) error {
	metadata, err := db.GetSessionMetadataId(&sessionID)
	if err != nil {
		return err
	}

	if metadata == nil {
		return nil
	}

	if len(metadata.Genres) == 0 {
		return nil
	}

	for _, gname := range metadata.Genres {
		if gname == "" {
			continue
		}

		var g statsusers.Genre
		if err := tx.Where("name = ?", gname).FirstOrCreate(&g, statsusers.Genre{Name: gname}).Error; err != nil {
			return err
		}

		for _, uid := range userIDs {
			genreID := g.ID
			userIDPtr := uid

			var existingLink statsusers.SessionsStatsGenres_users
			err := tx.Where("user_id = ? AND genre_id = ?", uid, genreID).First(&existingLink).Error

			if err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					count := uint16(1)
					link := statsusers.SessionsStatsGenres_users{
						UserID:  &userIDPtr,
						GenreID: &genreID,
						Count:   &count,
					}
					if err := tx.Create(&link).Error; err != nil {
						return fmt.Errorf("failed to create genre link for user %d and genre %d: %w", uid, genreID, err)
					}
				} else {
					return fmt.Errorf("failed to check existing genre link: %w", err)
				}
			} else {
				if err := tx.Model(&existingLink).UpdateColumn("count", gorm.Expr("count + 1")).Error; err != nil {
					return fmt.Errorf("failed to increment genre count for user %d and genre %d: %w", uid, genreID, err)
				}
			}
		}
	}
	return nil
}

func updateCategoryStats(tx *gorm.DB, userID uint) error {
	type categoryCount struct {
		SessionTypeID uint
		Count         int
	}

	var counts []categoryCount
	err := tx.Raw(`
		select s.session_type_id, count(*) as count
		from session_users su
		join sessions s on s.id = su.session_id
		join statuses st on st.id = s.status_id
		where su.user_id = ?
		  and st.status = 'Завершена'
		group by s.session_type_id
	`, userID).Scan(&counts).Error

	if err != nil {
		return err
	}

	for _, count := range counts {
		if err := incrementTypeCategoryCount(tx, userID, count.SessionTypeID, count.Count); err != nil {
			return err
		}
	}

	return nil
}

func incrementTypeCategoryCount(tx *gorm.DB, userID, sessionTypeID uint, totalCount int) error {
	switch sessionTypeID {
	case 1:
		return tx.Model(&statsusers.SessionStats_users{}).
			Where("user_id = ?", userID).
			Update("count_films", totalCount).Error
	case 2:
		return tx.Model(&statsusers.SessionStats_users{}).
			Where("user_id = ?", userID).
			Update("count_games", totalCount).Error
	case 3:
		return tx.Model(&statsusers.SessionStats_users{}).
			Where("user_id = ?", userID).
			Update("count_table_games", totalCount).Error
	default:
		return tx.Model(&statsusers.SessionStats_users{}).
			Where("user_id = ?", userID).
			Update("count_another", totalCount).Error
	}
}

func keys(m map[uint]struct{}) []uint {
	out := make([]uint, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	return out
}
