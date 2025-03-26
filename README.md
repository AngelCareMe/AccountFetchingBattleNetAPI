Понял, ты прав, твой проект больше ориентирован на бэкенд, так как основная функциональность — это API для получения данных из Blizzard API, а HTML-шаблоны скорее вспомогательные. Давай переделаю README, убрав акцент на фронтенд и сосредоточившись на том, что это бэкенд API для получения данных аккаунта Battle.net из World of Warcraft.

GuildTracker API
GuildTracker API — это бэкенд-сервис на Go, который предоставляет доступ к данным аккаунта World of Warcraft через Blizzard API. Проект использует OAuth 2.0 для авторизации через Battle.net и позволяет получить информацию о персонажах пользователя, включая их уровень предметов (Item Level), гильдию и рейтинг Mythic+. Сервис настроен для работы в Docker, что упрощает деплой.

Что умеет API
Авторизация через Blizzard OAuth:
Использует OAuth 2.0 для получения токена доступа через Battle.net.
Получение данных профиля:
Запрашивает профиль аккаунта через Blizzard API (profile/user/wow).
Данные о персонажах:
Имя, класс, раса и уровень персонажа.
Средний уровень предметов (Item Level).
Название гильдии.
Рейтинг Mythic+.
Рендеринг данных:
Для демонстрации возвращает данные через HTML-шаблоны (login.html для входа, characters.html для списка персонажей).
Технологии
Go: Бэкенд написан на Go с использованием стандартной библиотеки (net/http, html/template и др.).
Blizzard API: Для получения данных о профилях и персонажах.
Docker: Для контейнеризации и упрощения деплоя.
Установка и запуск
Требования
Установленный Go (версия 1.21 или выше).
Установленный Docker и Docker Compose.
Учётная запись Blizzard и зарегистрированное приложение в Blizzard API Portal для получения client_id и client_secret.
Шаги для запуска
Клонируйте репозиторий:
bash

Collapse

Wrap

Copy
git clone https://github.com/abasr/guildtracker.git
cd guildtracker
Настройте Blizzard API:
Зарегистрируйте приложение в Blizzard API Portal.
Получите client_id и client_secret.
В файле handlers/handlers.go замените значения переменных clientID и clientSecret на свои:
go

Collapse

Wrap

Copy
var (
    clientID     = "your_client_id"
    clientSecret = "your_client_secret"
    redirectURI  = "http://localhost:8080/callback"
)
Запустите с помощью Docker:
Убедитесь, что Docker и Docker Compose установлены.
Запустите сервис:
bash

Collapse

Wrap

Copy
docker-compose up --build
API будет доступно по адресу http://localhost:8080.
Запуск без Docker (опционально):
Убедитесь, что Go установлен.
Запустите сервер:
bash

Collapse

Wrap

Copy
go run cmd/main.go
Сервис будет доступен по адресу http://localhost:8080.
Использование
Перейдите на http://localhost:8080 для авторизации.
Используйте эндпоинты:
GET /login: Перенаправляет на страницу авторизации Blizzard.
GET /callback: Обрабатывает ответ от Blizzard, получает данные о персонажах и возвращает их через HTML-шаблон.
