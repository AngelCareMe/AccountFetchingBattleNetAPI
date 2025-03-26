GuildTracker API 🚀
  

GuildTracker API — это мощный бэкенд-сервис на Go, который позволяет получать данные аккаунта World of Warcraft через Blizzard API. Сервис авторизует пользователей через Battle.net, запрашивает информацию о персонажах и возвращает такие данные, как Item Level, гильдия и Mythic+ рейтинг. Всё это упаковано в Docker для удобного деплоя! 🐳

🌟 Что умеет API
Авторизация через Blizzard OAuth 🔐
Использует OAuth 2.0 для безопасной авторизации через Battle.net.
Получение данных профиля 📊
Запрашивает профиль аккаунта через Blizzard API (profile/user/wow).
Информация о персонажах 🧙‍♂️
Имя, класс, раса и уровень.
Средний уровень предметов (Item Level).
Название гильдии.
Рейтинг Mythic+.
Демонстрация через HTML 🖥️
Для удобства данные отображаются через HTML-шаблоны (login.html и characters.html).
🛠️ Технологии
Go 1.21 — для бэкенда.
Blizzard API — для получения данных.
Docker — для контейнеризации.
HTML-шаблоны — для демонстрации (временное решение).
🚀 Установка и запуск
Требования
Go (версия 1.21 или выше) 🐹
Docker и Docker Compose 🐳
Учётная запись Blizzard и зарегистрированное приложение в Blizzard API Portal для получения client_id и client_secret.
Шаги
Клонируйте репозиторий 📂
Настройте Blizzard API 🔑
Зарегистрируйте приложение в Blizzard API Portal.
Получите client_id и client_secret.
Обновите файл handlers/handlers.go:

var (
    clientID     = "your_client_id"
    clientSecret = "your_client_secret"
    redirectURI  = "http://localhost:8080/callback"
)
Запустите с помощью Docker 🐳

docker-compose up --build
API будет доступно по адресу: http://localhost:8080.
Запуск без Docker (опционально) 🖥️
bash

go run cmd/main.go
Сервис запустится на http://localhost:8080.
Использование
Перейдите на http://localhost:8080 для авторизации.
Доступные эндпоинты:
GET /login — перенаправляет на авторизацию Blizzard.
GET /callback — получает данные о персонажах и отображает их через HTML-шаблон.
📁 Структура проекта

guildtracker/
├── api/                # 📂 Пустая директория для будущих API-эндпоинтов
├── cmd/                # 🚀 Точка входа приложения
│   ├── env/            # 📂 Пустая директория для конфигурации
│   ├── main.go         # 📜 Главный файл приложения
│   └── handlers/
│       └── handlers.go # 🛠️ Логика обработки запросов и работы с Blizzard API
├── models/
│   └── character.go    # 📋 Модель данных для персонажа
├── templates/          # 🖼️ HTML-шаблоны для демонстрации
│   ├── login.html      # 🔐 Страница входа
│   └── characters.html # 📊 Страница с данными о персонажах
├── Dockerfile          # 🐳 Файл для сборки Docker-образа
├── docker-compose.yml  # 🐳 Конфигурация Docker Compose
└── go.mod              # 📦 Зависимости Go-модуля
⚠️ Примечания
Лимиты Blizzard API ⏳
Сервис добавляет задержки (time.Sleep) между запросами, чтобы не превысить лимиты. Настройте их в handlers.go, если нужно.
Безопасность 🔒
Сейчас clientID и clientSecret пустые. В продакшене используйте переменные окружения.
HTML-шаблоны 📄
Используются для демонстрации. В будущем планируется переход на JSON API.
📅 Планы на будущее
Переделать ответы в JSON для полноценного API. 📡
Добавить поддержку переменных окружения для clientID и clientSecret. 🌍
Подключить базу данных для сохранения данных о персонажах. 💾
Добавить тесты для обработчиков. ✅
📜 Лицензия
Проект распространяется под лицензией MIT. Подробности в файле LICENSE.

📬 Контакты
Есть вопросы или предложения? Пишите: abasr@example.com ✉️
