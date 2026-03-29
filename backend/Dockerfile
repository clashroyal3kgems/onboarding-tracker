FROM golang:1.26-alpine

WORKDIR /app

# Копируем файлы зависимостей
COPY go.mod go.sum ./
RUN go mod download

# Копируем весь проект (включая папку frontend)
COPY . .

# Собираем приложение
RUN go build -o main ./backend/main.go

EXPOSE 8080
CMD ["./main"]