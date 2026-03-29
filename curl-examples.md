1. Авторизация (Login)
Проверяем получение ID и роли.


curl -X POST http://localhost:8080/api/login \
     -H "Content-Type: application/json" \
     -d '{"id": 1, "password": "123"}'


2. Работа с планами (Onboarding Plans)
Создать новый план (назначить материалы сотруднику):

curl -X POST http://localhost:8080/api/onboarding-plans \
     -H "Content-Type: application/json" \
     -d '{
           "employee_id": 3,
           "mentor_id": 1,
           "material_ids": [1, 2, 5]
         }'

Получить план текущего пользователя:

curl -X GET "http://localhost:8080/api/my-plan?userId=3" \
     -H "X-User-Id: 3" \
     -H "X-User-Role: employee"

3. Работа с материалами (Items)
Обновить статус задачи (тот самый POST, который у тебя в коде):
По ТЗ ты используешь POST для обновления (хотя по канонам REST тут подошел бы PUT или PATCH, но мы следуем твоему HandleFunc).

curl -X POST http://localhost:8080/api/onboarding-items/10/status \
     -H "Content-Type: application/json" \
     -H "X-User-Id: 3" \
     -H "X-User-Role: employee" \
     -d '{"status": "done", "comment": "Изучил основы Go, всё понятно!"}'

4. Завершение плана
Принудительно закрыть план:

curl -X POST http://localhost:8080/api/onboarding-plans/1/complete \
     -H "X-User-Id: 1" \
     -H "X-User-Role: mentor"
     
5. Управление материалами (Admin)
Создать новый учебный материал:

curl -X POST http://localhost:8080/api/materials \
     -H "Content-Type: application/json" \
     -H "X-User-Role: admin" \
     -d '{
           "title": "Docker для новичков",
           "description": "Как запускать контейнеры",
           "link": "https://docker.com/guide"
         }'
Удалить материал:

curl -X DELETE http://localhost:8080/api/materials/5 \
     -H "X-User-Role: admin"