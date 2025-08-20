package services

import (
	"errors"
	"friendship/db"
	"friendship/models"
	"friendship/models/news"
	"strings"
	"time"
)

type CreateNewsInput struct {
	Title       string `json:"title" binding:"required"`
	Description string `json:"description" binding:"required"`
	Image       string `json:"image" binding:"required"`
	Text        string `json:"text" binding:"required"`
}

type NewsPage struct {
	Items   []news.News
	Page    int
	Total   int64
	HasMore bool
}

type CommentDTO struct {
	ID     int    `json:"id"`
	UserID uint   `json:"user_id"`
	Name   string `json:"name"`
	Image  string `json:"image"`
	Text   string `json:"text"`
}

type NewsDTO struct {
	ID          int          `json:"id"`
	Title       string       `json:"title"`
	Description string       `json:"description"`
	Image       string       `json:"image"`
	CreatedTime string       `json:"created_time"`
	Content     string       `json:"content"`
	Comments    []CommentDTO `json:"comments"`
}

type CreateCommentInput struct {
	Text string `json:"text" binding:"required"`
}

type CommentResponse struct {
	ID     uint   `json:"id"`
	NewsID uint   `json:"news_id"`
	Text   string `json:"text"`
	User   struct {
		Name  string `json:"name"`
		Image string `json:"image"`
	} `json:"user"`
}

func CreateNews(input CreateNewsInput) (*news.News, error) {
	if strings.TrimSpace(input.Title) == "" ||
		strings.TrimSpace(input.Description) == "" ||
		strings.TrimSpace(input.Image) == "" ||
		strings.TrimSpace(input.Text) == "" {
		return nil, errors.New("все поля обязательны")
	}

	database := db.GetDB()

	newsItem := &news.News{
		Title:       input.Title,
		Description: input.Description,
		Image:       input.Image,
		CreatedTime: time.Now().Format("2006-01-02 15:04:05"),
	}

	if err := database.Create(newsItem).Error; err != nil {
		return nil, err
	}

	content := &news.ContentNews{
		NewsID: uint(newsItem.ID),
		Text:   input.Text,
	}
	if err := database.Model(newsItem).Association("Content").Append(content); err != nil {
		return nil, err
	}

	if err := database.Preload("Content").First(newsItem, newsItem.ID).Error; err != nil {
		return nil, err
	}

	return newsItem, nil
}

func GetNews(page int) (*NewsPage, error) {
	const limit = 6
	offset := (page - 1) * limit

	var items []news.News
	var total int64
	database := db.GetDB()

	if err := database.Model(&news.News{}).Count(&total).Error; err != nil {
		return nil, err
	}

	if err := database.Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&items).Error; err != nil {
		return nil, err
	}

	hasMore := int64(page*limit) < total

	return &NewsPage{
		Items:   items,
		Page:    page,
		Total:   total,
		HasMore: hasMore,
	}, nil
}

func GetNewsByID(id int) (*NewsDTO, error) {
	database := db.GetDB()

	var item news.News
	if err := database.
		Preload("Content").
		Preload("Comments").
		Preload("Comments.User").
		First(&item, id).Error; err != nil {
		return nil, err
	}

	content := ""
	if len(item.Content) > 0 {
		content = item.Content[0].Text
	}

	var comments []CommentDTO
	for _, c := range item.Comments {
		comments = append(comments, CommentDTO{
			ID:     c.ID,
			UserID: c.User.ID,
			Name:   c.User.Name,
			Image:  c.User.Image,
			Text:   c.Text,
		})
	}

	result := &NewsDTO{
		ID:          item.ID,
		Title:       item.Title,
		Description: item.Description,
		Image:       item.Image,
		CreatedTime: item.CreatedTime,
		Content:     content,
		Comments:    comments,
	}

	return result, nil
}

func CreateComment(newsID uint, email string, input CreateCommentInput) (*CommentResponse, error) {
	database := db.GetDB()

	if input.Text == "" {
		return nil, errors.New("текст комментария обязателен")
	}

	var user models.User
	if err := database.Where("email = ?", email).First(&user).Error; err != nil {
		return nil, err
	}

	comment := &news.Comments{
		NewsID: newsID,
		UserID: user.ID,
		Text:   input.Text,
	}

	if err := database.Create(comment).Error; err != nil {
		return nil, err
	}

	resp := &CommentResponse{
		ID:     uint(comment.ID),
		NewsID: comment.NewsID,
		Text:   comment.Text,
		User: struct {
			Name  string `json:"name"`
			Image string `json:"image"`
		}{
			Name:  user.Name,
			Image: user.Image,
		},
	}

	return resp, nil
}

func DeleteComment(commentID int) error {
	database := db.GetDB()

	if err := database.Delete(&news.Comments{}, commentID).Error; err != nil {
		return err
	}

	return nil
}
