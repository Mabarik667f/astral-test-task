# Работа с приложением
1. Скопируйте example.env содержимое в .env (для удобства вставил уже сгенерированный секрет)
2. В postgres создать бд 
```sql
CREATE DATABASE astral_docs;
```
3. Выполнить миграции
```sh
goose up
```
4. Установить зависимости
```sh
task tidy
```
5. Запустить приложение (dev mode)
```sh 
go run ./cmd/main.go
```
## Сборка (Build)
```sh
CGO_ENABLED=0 GOOS=linux go build -o app.out ./cmd/main.go
```

## Запуск тестов
```sh
task fast-test
task test
```

### Прочее

Astral.postman_collection.json - Postman коллекция версии 2.1

- [task](https://github.com/go-task/task) - замена Make
- [mockgen](https://github.com/uber-go/mock) - для моков
- [goose](https://github.com/pressly/goose) - для миграций

Генерация админ-секрета
```sh
openssl rand -base64 64 
```
