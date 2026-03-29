package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

type User struct {
	ID       int    `json:"id" db:"id"`
	Name     string `json:"name" db:"name"`
	Role     string `json:"role" db:"role"`
	Password string `json:"-" db:"password"`
}
type OnboardingItem struct {
	ID         int     `json:"id" db:"id"`
	PlanID     int     `json:"plan_id" db:"plan_id"`
	MaterialID int     `json:"material_id" db:"material_id"`
	Title      string  `json:"title" db:"title"`
	Status     string  `json:"status" db:"status"`
	Comment    *string `json:"comment" db:"comment"`
}
type Material struct {
	ID          int    `json:"id" db:"id"`
	Title       string `json:"title" db:"title"`
	Description string `json:"description" db:"description"`
	Link        string `json:"link" db:"link"`
}

type OnboardingPlan struct {
	ID         int    `json:"id" db:"id"`
	EmployeeID int    `json:"employee_id" db:"employee_id"`
	MentorID   int    `json:"mentor_id" db:"mentor_id"`
	Status     string `json:"status" db:"status"`
}

type PlanRequest struct {
	EmployeeID  int   `json:"employee_id"`
	MentorID    int   `json:"mentor_id"`
	MaterialIDs []int `json:"material_ids"`
}

var db *sqlx.DB

func main() {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = "user=user password=password dbname=onboarding sslmode=disable host=localhost"
	}

	// Цикл ожидания базы данных (Retry logic)
	var err error
	// В func main() твоего main.go измени цикл на этот:
	for i := 1; i <= 10; i++ { // Увеличиваем до 10 попыток
		db, err = sqlx.Connect("postgres", dsn)
		if err == nil {
			err = db.Ping() // Проверяем реальное соединение
			if err == nil {
				break
			}
		}
		log.Printf("Попытка %d: База еще не готова, ждем 5 секунд... (%v)", i, err)
		time.Sleep(5 * time.Second) // Ждем по 5 секунд
	}

	if err != nil {
		log.Fatalln("Не удалось подключиться к БД после 5 попыток:", err)
	}

	r := mux.NewRouter()

	// API эндпоинты
	r.HandleFunc("/api/users", getUsers).Methods("GET")
	r.HandleFunc("/api/login", loginHandler).Methods("POST")
	r.HandleFunc("/api/my-plan", getMyPlan).Methods("GET")
	r.HandleFunc("/api/onboarding-plans/{plan_id}/items/{item_id}", updateItemHandler).Methods("PUT")
	r.HandleFunc("/api/onboarding-plans/{id}/status", updatePlanStatus).Methods("POST")

	r.HandleFunc("/api/onboarding-plans/{id}/items", getPlanItems).Methods("GET")

	r.HandleFunc("/api/onboarding-plans", createPlan).Methods("POST")
	r.HandleFunc("/api/onboarding-plans", getPlans).Methods("GET")

	r.HandleFunc("/api/my-plan", getMyPlan).Methods("GET")
	r.HandleFunc("/api/onboarding-items/{id}/status", updateItemStatus).Methods("POST")

	r.HandleFunc("/api/onboarding-plans/{id}/complete", completePlan).Methods("POST")

	r.HandleFunc("/api/materials", getMaterials).Methods("GET")
	r.HandleFunc("/api/materials", createMaterial).Methods("POST")
	r.HandleFunc("/api/materials/{id}", deleteMaterial).Methods("DELETE")

	// В func main() перед PathPrefix("/"):
	r.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./frontend/login.html")
	})
	r.HandleFunc("/onboarding/my", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./frontend/onboarding-my.html")
	})
	// И так далее для остальных страниц

	// Статика: фронтенд теперь ищется в корневой папке frontend внутри контейнера
	r.PathPrefix("/").Handler(http.FileServer(http.Dir("./frontend/")))

	fmt.Println("Server started at :8080")
	log.Fatal(http.ListenAndServe(":8080", r))
}

func ProtectedHandler(w http.ResponseWriter, r *http.Request) {
	// Запрещаем кэширование
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate") // HTTP 1.1
	w.Header().Set("Pragma", "no-cache")                                   // HTTP 1.0
	w.Header().Set("Expires", "0")                                         // Proxies

	// Далее ваша логика...
}

func getUsers(w http.ResponseWriter, r *http.Request) {
	var users []User
	// Добавляем password, чтобы sqlx мог заполнить структуру целиком
	err := db.Select(&users, "SELECT id, name, role, password FROM users")
	if err != nil {
		log.Printf("ОШИБКА В GET_USERS: %v", err) // Посмотри это сообщение в терминале!
		http.Error(w, "Ошибка базы данных: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(users)
}
func loginHandler(w http.ResponseWriter, r *http.Request) {
	var creds struct {
		ID       int    `json:"id"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&creds); err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	var user User
	// Ищем пользователя и его пароль
	err := db.Get(&user, "SELECT id, role, password FROM users WHERE id=$1", creds.ID)
	if err != nil || user.Password != creds.Password {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Отправляем данные для сохранения в localStorage
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"user_id": user.ID,
		"role":    user.Role,
	})
}

func getMaterials(w http.ResponseWriter, r *http.Request) {
	var materials []Material
	err := db.Select(&materials, "SELECT id, title, description, link FROM materials ORDER BY id DESC")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(materials)
}

func createMaterial(w http.ResponseWriter, r *http.Request) {
	// Временно убираем проверку роли для теста, чтобы исключить эту причину
	// role := r.Header.Get("X-User-Role")

	var m Material
	if err := json.NewDecoder(r.Body).Decode(&m); err != nil {
		log.Printf("Ошибка декодирования JSON: %v", err)
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	// Логируем, что пришло с фронта
	log.Printf("Попытка создать материал: %+v", m)

	_, err := db.Exec("INSERT INTO materials (title, description, link) VALUES ($1, $2, $3)",
		m.Title, m.Description, m.Link)

	if err != nil {
		log.Printf("ОШИБКА БАЗЫ ДАННЫХ: %v", err) // Это появится в терминале Docker
		http.Error(w, "Database error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func deleteMaterial(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("X-User-Role") != "admin" {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	vars := mux.Vars(r)
	id := vars["id"]
	_, err := db.Exec("DELETE FROM materials WHERE id = $1", id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func createPlan(w http.ResponseWriter, r *http.Request) {
	var req PlanRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// 1. Создаем сам план
	var planID int
	// Было 'active', стало 'todo'
	err := db.QueryRow("INSERT INTO onboarding_plans (employee_id, mentor_id, status) VALUES ($1, $2, 'todo') RETURNING id",
		req.EmployeeID, req.MentorID).Scan(&planID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// 2. Добавляем шаги (материалы) в этот план
	for _, mID := range req.MaterialIDs {
		// Получаем название материала для копирования в title шага
		var title string
		db.Get(&title, "SELECT title FROM materials WHERE id=$1", mID)

		db.Exec("INSERT INTO onboarding_items (plan_id, material_id, title) VALUES ($1, $2, $3)",
			planID, mID, title)
	}

	w.WriteHeader(http.StatusCreated)
}

func getPlans(w http.ResponseWriter, r *http.Request) {
	var plans []OnboardingPlan
	// Явно перечисляем поля, чтобы исключить ошибки
	err := db.Select(&plans, "SELECT id, employee_id, mentor_id, status FROM onboarding_plans ORDER BY id DESC")

	if err != nil {
		log.Printf("ОШИБКА SELECT PLANS: %v", err) // Посмотри это в docker logs
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Логируем для себя, сколько планов нашли
	log.Printf("Найдено планов в базе: %d", len(plans))

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(plans)
}

func getMyPlan(w http.ResponseWriter, r *http.Request) {
	// Получаем ID из заголовка, который передает фронтенд
	userID := r.Header.Get("X-User-Id")
	if userID == "" {
		userID = r.URL.Query().Get("userId") // Оставим как запасной вариант
	}
	if userID == "" {
		http.Error(w, "userId is required", http.StatusBadRequest)
		return
	}

	// 1. Находим активный план сотрудника
	var plan OnboardingPlan
	// Добавляем ORDER BY id DESC, чтобы брать последний созданный план
	err := db.Get(&plan, `
    SELECT id, employee_id, mentor_id, status 
    FROM onboarding_plans 
    WHERE employee_id = $1 AND status = 'active' 
    ORDER BY id DESC LIMIT 1`, userID)
	if err != nil {
		log.Printf("План не найден для пользователя %s: %v", userID, err)
		http.Error(w, "План не найден", http.StatusNotFound)
		return
	}

	// 2. Получаем все задачи (items) этого плана
	type Item struct {
		ID     int    `json:"id" db:"id"`
		Title  string `json:"title" db:"title"`
		Status string `json:"status" db:"status"`
	}
	var items []Item
	err = db.Select(&items, "SELECT id, title, status FROM onboarding_items WHERE plan_id = $1 ORDER BY id ASC", plan.ID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"plan":  plan,
		"items": items,
	})
}

func updateItemStatus(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	itemID := vars["id"]

	var req struct {
		Status  string `json:"status"`
		Comment string `json:"comment"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	// УДАЛИЛИ updated_at = NOW(), так как это вызывает ошибку, если колонки нет
	_, err := db.Exec(`UPDATE onboarding_items 
                       SET status = $1, comment = $2 
                       WHERE id = $3`, req.Status, req.Comment, itemID)
	if err != nil {
		log.Printf("Ошибка обновления задачи: %v", err)
		http.Error(w, err.Error(), 500)
		return
	}

	// Автоматика завершения плана
	var planID int
	db.Get(&planID, "SELECT plan_id FROM onboarding_items WHERE id = $1", itemID)

	var unfinishedCount int
	// Проверяем, сколько шагов НЕ в статусе 'done'
	db.Get(&unfinishedCount, "SELECT COUNT(*) FROM onboarding_items WHERE plan_id = $1 AND status != 'done'", planID)

	if unfinishedCount == 0 {
		// Обновляем статус на 'done' (чтобы фильтры на фронте его видели)
		db.Exec("UPDATE onboarding_plans SET status = 'done' WHERE id = $1", planID)
		log.Printf("План %d автоматически завершен!", planID)
	} else {
		// Если начали что-то делать, но не закончили — ставим 'in_progress'
		db.Exec("UPDATE onboarding_plans SET status = 'in_progress' WHERE id = $1", planID)
	}

	w.WriteHeader(http.StatusOK)
}
func completePlan(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	planID := vars["id"]

	_, err := db.Exec("UPDATE onboarding_plans SET status = 'done' WHERE id = $1", planID)
	if err != nil {
		log.Printf("Ошибка при завершении плана: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, "Поздравляем! Онбординг завершен.")
}

func getPlanItems(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	planID := vars["id"]

	var items []OnboardingItem
	// COALESCE вернет пустую строку, если в базе NULL
	query := `
        SELECT id, plan_id, material_id, title, status, COALESCE(comment, '') as comment 
        FROM onboarding_items 
        WHERE plan_id = $1 
        ORDER BY id ASC`

	err := db.Select(&items, query, planID)
	if err != nil {
		log.Printf("Ошибка базы: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(items)
}
func updatePlanStatus(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	planID := vars["id"]

	var req struct {
		Status string `json:"status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	// Обновляем статус плана. Используем "done", чтобы совпадало с фронтом.
	_, err := db.Exec("UPDATE onboarding_plans SET status = $1 WHERE id = $2", req.Status, planID)
	if err != nil {
		log.Printf("Ошибка обновления статуса плана %s: %v", planID, err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
func updateItemHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	planID := vars["plan_id"]
	itemID := vars["item_id"]

	var updateData struct {
		Status  string `json:"status"`
		Comment string `json:"comment"`
	}

	if err := json.NewDecoder(r.Body).Decode(&updateData); err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	// 1. Обновляем саму задачу
	_, err := db.Exec(`
        UPDATE onboarding_items 
        SET status = $1, comment = $2 
        WHERE id = $3 AND plan_id = $4`,
		updateData.Status, updateData.Comment, itemID, planID)

	if err != nil {
		log.Printf("Error updating item: %v", err)
		http.Error(w, "DB Error", 500)
		return
	}

	// 2. АВТОМАТИКА: Обновляем статус всего плана
	// Считаем количество выполненных и общее количество задач
	var stats struct {
		Total int `db:"total"`
		Done  int `db:"done"`
	}

	err = db.Get(&stats, `
        SELECT 
            count(*) as total, 
            count(*) FILTER (WHERE status = 'done') as done 
        FROM onboarding_items 
        WHERE plan_id = $1`, planID)

	if err == nil && stats.Total > 0 {
		newPlanStatus := "in_progress" // СТРОГО ТАК, как в <option value="...">

		if stats.Done == stats.Total {
			newPlanStatus = "done"
		}
		// Если есть хоть один выполненный пункт или просто план начат — это in_progress

		_, errExec := db.Exec("UPDATE onboarding_plans SET status = $1 WHERE id = $2", newPlanStatus, planID)
		if errExec != nil {
			log.Printf("Ошибка обновления плана %s: %v", planID, errExec)
		} else {
			log.Printf("План %s переведен в статус: %s", planID, newPlanStatus)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})

}
