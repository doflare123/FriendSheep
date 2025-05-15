package db

import (
	"friendship/models" // Импортирование пакета с определением моделей данных
)

// Функция для получения всех заметок из базы данных
func GetAllUsers() []models.User {
	var Users []models.User
	// Использование GORM для выполнения SQL-запроса SELECT и заполнения среза Users
	db.Find(&Users)
	return Users // Возвращение всех найденных заметок
}

// Функция для получения заметки по ID
func GetUserByID(id int) *models.User {
	var User models.User
	// Использование GORM для выполнения SQL-запроса SELECT с условием WHERE id = id заметки
	if result := db.First(&User, id); result.Error != nil {
		return nil // Возвращение nil, если заметка с указанным ID не найдена
	}
	return &User // Возвращение найденной заметки
}

// Функция для создания новой заметки
func CreateUser(name, email string) models.User {
	User := models.User{
		Name:  name,
		Email: email,
	}
	// Использование GORM для выполнения SQL-запроса INSERT и сохранения новой заметки в базе данных
	db.Create(&User)
	return User // Возвращение созданной заметки
}

// Функция для обновления существующей заметки по ID
func UpdateUser(id int, name, email string) *models.User {
	var User models.User
	// Использование GORM для выполнения SQL-запроса SELECT с условием WHERE id = id заметки
	if result := db.First(&User, id); result.Error != nil {
		return nil // Возвращение nil, если заметка с указанным ID не найдена
	}
	User.Name = name
	User.Email = email
	// Использование GORM для выполнения SQL-запроса UPDATE и сохранения обновленной заметки в базе данных
	db.Save(&User)
	return &User // Возвращение обновленной заметки
}

// Функция для удаления заметки по ID
func DeleteUserByID(id int) bool {
	// Использование GORM для выполнения SQL-запроса DELETE с условием WHERE id = id заметки
	if result := db.Delete(&models.User{}, id); result.Error != nil {
		return false // Возвращение false, если удаление не удалось
	}
	return true // Возвращение true при успешном удалении заметки
}
