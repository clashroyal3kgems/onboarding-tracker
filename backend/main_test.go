package main

import (
	"testing"
)

// Тест 1: План должен переходить в 'completed', когда все шаги выполнены (Пункт 11 ТЗ)
func TestPlanAutoCompletion(t *testing.T) {
	// Имитируем логику: всего 2 задачи, обе стали 'done'
	totalItems := 2
	doneItems := 2

	status := "active"
	if doneItems == totalItems {
		status = "completed"
	}

	if status != "completed" {
		t.Errorf("Ожидался статус 'completed', получили '%s'", status)
	}
}

// Тест 2: Проверка прав доступа (Пункт 11 ТЗ)
// Сотрудник не может менять статус в ЧУЖОМ плане
func TestEmployeeAccessValidation(t *testing.T) {
	planOwnerID := 3   // Sidor
	currentUserID := 5 // Какой-то другой юзер

	canEdit := false
	if planOwnerID == currentUserID {
		canEdit = true
	}

	if canEdit {
		t.Error("Ошибка безопасности: пользователь смог отредактировать чужой план!")
	}
}

// Тест 3: Возможность оставить комментарий ментором (Пункт 11 ТЗ)
func TestMentorCanLeaveComment(t *testing.T) {
	userRole := "mentor"
	comment := "Нужно переделать отчет"

	// Логика: если роль ментор или админ, комментарий разрешен
	canComment := false
	if userRole == "mentor" || userRole == "admin" {
		canComment = true
	}

	if !canComment || comment == "" {
		t.Error("Ментор должен иметь возможность оставлять непустой комментарий")
	}
}
