# ООО "Обувь" — Спецификация проекта

Информационная система для магазина обуви. Бекенд на Go, фронтенд на Next.js + Ant Design, база данных SQLite.

## Стек технологий

- **Backend**: Go, chi (роутер), modernc.org/sqlite (pure Go), golang-jwt/jwt, rs/cors
- **Frontend**: Next.js (App Router), Ant Design, TypeScript
- **Database**: SQLite (файл data.db)
- **Repo**: монорепо — `backend/` и `frontend/` в одном репозитории

## Архитектура

```
Browser (Next.js :3000) --fetch()--> Go API (:8080) --SQL--> SQLite (data.db)
                                         |
                                    uploads/ (фото товаров, http.FileServer)
```

Go API — классический REST. Слоёная структура: handler → service → repository → model. Middleware: JWT auth, CORS, role check.

Next.js — App Router. Серверные компоненты для начальной загрузки, клиентские для интерактива (фильтры, поиск). Ant Design компоненты для UI.

## Структура монорепо

```
backend/
  ├── cmd/server/main.go
  ├── internal/
  │   ├── handler/        # HTTP handlers
  │   ├── service/        # бизнес-логика
  │   ├── repository/     # SQL запросы
  │   ├── model/          # структуры данных
  │   ├── middleware/      # JWT, CORS, роли
  │   └── database/       # миграции, подключение, seed
  ├── uploads/            # фото товаров
  ├── go.mod
  └── go.sum

frontend/
  ├── src/app/            # App Router pages
  │   ├── login/
  │   ├── products/
  │   ├── orders/
  │   └── layout.tsx
  ├── src/components/     # переиспользуемые компоненты
  ├── src/lib/            # API клиент, auth утилиты
  ├── src/types/          # TypeScript типы
  ├── public/             # статика фронта
  ├── package.json
  └── next.config.js
```

## Схема базы данных

10 таблиц, 3NF, ссылочная целостность через FK с ON DELETE RESTRICT (кроме order_items).

### users
| Поле | Тип | Описание |
|------|-----|----------|
| id | INTEGER PK | Автоинкремент |
| login | TEXT NOT NULL | Логин (email) |
| password | TEXT NOT NULL | bcrypt-хеш пароля |
| last_name | TEXT | Фамилия |
| first_name | TEXT | Имя |
| patronymic | TEXT | Отчество |
| role_id | INTEGER FK | → roles.id |

### roles
| Поле | Тип | Описание |
|------|-----|----------|
| id | INTEGER PK | Автоинкремент |
| name | TEXT NOT NULL | client / manager / admin |

Роль "гость" — это отсутствие аутентификации, в таблице roles не хранится. Импортные роли: "Администратор", "Менеджер", "Авторизированный клиент" → admin, manager, client.

### products
| Поле | Тип | Описание |
|------|-----|----------|
| id | INTEGER PK | Автоинкремент |
| article | TEXT NOT NULL UNIQUE | Артикул (напр. "А112Т4") |
| name | TEXT NOT NULL | Наименование |
| description | TEXT | Описание |
| price | REAL NOT NULL | Цена (≥ 0, с копейками) |
| discount | REAL DEFAULT 0 | Скидка в % |
| quantity | INTEGER DEFAULT 0 | Количество на складе (≥ 0) |
| image | TEXT | Путь к файлу в uploads/ |
| category_id | INTEGER FK | → categories.id |
| manufacturer_id | INTEGER FK | → manufacturers.id |
| supplier_id | INTEGER FK | → suppliers.id |
| unit_id | INTEGER FK | → units.id |

### categories
| Поле | Тип |
|------|-----|
| id | INTEGER PK |
| name | TEXT NOT NULL |

### manufacturers
| Поле | Тип |
|------|-----|
| id | INTEGER PK |
| name | TEXT NOT NULL |

### suppliers
| Поле | Тип |
|------|-----|
| id | INTEGER PK |
| name | TEXT NOT NULL |

### units
| Поле | Тип |
|------|-----|
| id | INTEGER PK |
| name | TEXT NOT NULL |

### orders
| Поле | Тип | Описание |
|------|-----|----------|
| id | INTEGER PK | Автоинкремент |
| order_date | TEXT NOT NULL | Дата заказа |
| delivery_date | TEXT | Дата доставки/выдачи (может быть NULL) |
| pickup_code | TEXT | Код для получения (напр. "901") |
| status_id | INTEGER FK | → order_statuses.id |
| pickup_point_id | INTEGER FK | → pickup_points.id |
| user_id | INTEGER FK | → users.id (клиент, оформивший заказ) |

Примечание: в ТЗ модуль 4 использует термин "дата выдачи", в импортных данных — "дата доставки". Это одно и то же поле, в БД `delivery_date`.

### order_items
| Поле | Тип | Описание |
|------|-----|----------|
| id | INTEGER PK | Автоинкремент |
| order_id | INTEGER FK | → orders.id (ON DELETE CASCADE) |
| product_id | INTEGER FK | → products.id (ON DELETE RESTRICT) |
| quantity | INTEGER NOT NULL | Количество в заказе |

### order_statuses
| Поле | Тип |
|------|-----|
| id | INTEGER PK |
| name | TEXT NOT NULL |

### pickup_points
| Поле | Тип |
|------|-----|
| id | INTEGER PK |
| address | TEXT NOT NULL |

## REST API

### Аутентификация

| Метод | Путь | Описание | Доступ |
|-------|------|----------|--------|
| POST | /api/auth/login | Логин → JWT токен | public |

Тело запроса: `{ "login": "...", "password": "..." }`
Ответ: `{ "token": "jwt...", "user": { id, login, fullName, role } }`

### Товары

| Метод | Путь | Описание | Доступ |
|-------|------|----------|--------|
| GET | /api/products | Список товаров | public |
| GET | /api/products/{id} | Один товар | public |
| POST | /api/products | Создать товар | admin |
| PUT | /api/products/{id} | Обновить товар | admin |
| DELETE | /api/products/{id} | Удалить товар | admin |
| POST | /api/products/{id}/image | Загрузить/заменить фото | admin |

Query-параметры GET /api/products:
- `?search=текст` — поиск по всем текстовым полям (LIKE)
- `?sort=quantity_asc|quantity_desc` — сортировка по количеству
- `?supplier_id=N` — фильтр по поставщику

### Заказы

| Метод | Путь | Описание | Доступ |
|-------|------|----------|--------|
| GET | /api/orders | Список заказов | manager+ |
| GET | /api/orders/{id} | Один заказ с позициями | manager+ |
| POST | /api/orders | Создать заказ | admin |
| PUT | /api/orders/{id} | Обновить заказ | admin |
| DELETE | /api/orders/{id} | Удалить заказ | admin |

### Справочники (GET only)

| Путь | Доступ |
|------|--------|
| /api/categories | auth |
| /api/manufacturers | auth |
| /api/suppliers | manager+ |
| /api/units | auth |
| /api/pickup-points | auth |
| /api/order-statuses | auth |

### Статика

| Путь | Описание |
|------|----------|
| /uploads/{filename} | Фото товаров (http.FileServer) |

### Уровни доступа

- **public** — без токена (гость)
- **auth** — любой авторизованный пользователь
- **manager+** — менеджер или администратор
- **admin** — только администратор

## Роли и права

| Роль | Каталог товаров | Фильтр/Поиск/Сортировка | CRUD товаров | Заказы | CRUD заказов |
|------|----------------|--------------------------|--------------|--------|--------------|
| Гость | Просмотр | Нет | Нет | Нет | Нет |
| Клиент | Просмотр | Нет | Нет | Нет | Нет |
| Менеджер | Просмотр | Да | Нет | Просмотр | Нет |
| Админ | Просмотр | Да | Да | Просмотр | Да |

## Фронтенд: страницы

| Роут | Описание | Доступ |
|------|----------|--------|
| /login | Страница входа | public |
| /products | Каталог товаров | public |
| /products/new | Добавить товар | admin |
| /products/[id]/edit | Редактировать товар | admin |
| /orders | Список заказов | manager+ |
| /orders/new | Добавить заказ | admin |
| /orders/[id]/edit | Редактировать заказ | admin |

Каждая страница задаёт `<title>` через Next.js metadata, соответствующий назначению (по ТЗ: "Заголовок окна должен соответствовать назначению").

## Стиль (по ТЗ)

- **Шрифт**: Times New Roman
- **Основной фон**: #FFFFFF
- **Дополнительный фон (хедер)**: #7FFF00
- **Акцент (кнопки действий)**: #00FA9A
- **Скидка > 15% — фон строки**: #2E8B57
- **Скидка — старая цена**: зачёркнутая, красный шрифт; новая цена — чёрный шрифт
- **Нет на складе — фон строки**: голубой (#e0f2fe)
- **Иконка приложения**: import/Icon.ico
- **Логотип на хедере**: import/Icon.png (не искажать пропорции)
- **Заглушка для фото**: import/picture.png

## UI-компоненты (Ant Design)

- **Layout**: хедер (#7FFF00) с логотипом, навигацией, ФИО + кнопка выхода
- **Навигация**: кнопка "Назад" на формах добавления/редактирования (по ТЗ: "перемещаться между существующими окнами, в том числе обратно")
- **Каталог**: кастомный список (не Table) — карточки товаров с условным стилем
- **Тулбар**: Input.Search + Select (поставщик) + кнопки сортировки — видны менеджеру и админу
- **Форма товара**:
  - name → Input
  - article → Input (при добавлении скрыт, при редактировании read-only)
  - category → Select (dropdown, данные из /api/categories)
  - manufacturer → Select (dropdown, данные из /api/manufacturers)
  - supplier → Select (dropdown, данные из /api/suppliers)
  - description → TextArea
  - price → InputNumber (≥ 0, step 0.01)
  - discount → InputNumber (≥ 0)
  - quantity → InputNumber (≥ 0, целое)
  - unit → Select (dropdown, данные из /api/units)
  - image → Upload (фото товара)
- **Список заказов**: Table с колонками (артикулы товаров, статус, пункт выдачи, дата заказа, дата доставки)
- **Форма заказа**:
  - status → Select (dropdown, данные из /api/order-statuses)
  - pickup_point → Select (dropdown, данные из /api/pickup-points)
  - order_date → DatePicker
  - delivery_date → DatePicker
  - pickup_code → Input (auto или ручной)
  - **Позиции заказа** → динамический список (Form.List): выбор товара по артикулу (Select) + количество (InputNumber). Можно добавлять/удалять строки.
- **Ошибки**: notification / Modal с заголовком, иконкой, информативным текстом

## Валидация и бизнес-правила

### Валидация на фронте (Ant Form rules)
- Цена ≥ 0, допускаются копейки (десятичная дробь)
- Количество ≥ 0, целое число
- Обязательные поля отмечены, визуальные подсказки при ошибках

### Валидация на бекенде
- Дублирует фронтовую валидацию
- Возвращает 400 с `{ "error": "описание" }`

### Удаление товара
- Бекенд проверяет наличие в order_items
- Если товар в заказе → 409 Conflict
- Фронт показывает Modal.warning с объяснением

### Удаление заказа
- Без ограничений, ON DELETE CASCADE на order_items

### Загрузка фото
- Бекенд проверяет MIME (jpeg/png)
- Ресайз до 300×200 пикселей
- Сохранение в uploads/, в БД — путь к файлу
- При замене старый файл удаляется

### Аутентификация
- JWT в httpOnly cookie
- Пароли хешируются через bcrypt (seed-скрипт хеширует пароли из Excel при импорте)
- JWT истёк → 401 → фронт редиректит на /login
- Роль не позволяет → 403, UI скрывает недоступные элементы заранее

### Форма редактирования товара
- Одновременно открыта только одна (реализовано через отдельный роут, не модалка)
- При редактировании ID read-only
- При добавлении ID скрыт (автоинкремент +1)

### Сообщения об ошибках (по ТЗ)
- Ant notification / Modal с заголовком, иконкой типа (error, warning, info) и информативным текстом
- "Полная информация о совершенных ошибках и порядок действий для их исправления"

## Импорт данных

Данные из Excel-файлов (`import/`) импортируются при инициализации БД. Реализуется как Go seed-команда, запускается однократно.

### Порядок импорта (с учётом FK-зависимостей)

1. **roles** — извлечь уникальные значения из колонки "Роль сотрудника" в user_import.xlsx. Маппинг: "Администратор" → admin, "Менеджер" → manager, "Авторизированный клиент" → client.
2. **users** ← user_import.xlsx — ФИО разбивается на last_name / first_name / patronymic. Пароли хешируются через bcrypt перед вставкой в БД.
3. **categories** — извлечь уникальные из колонки "Категория товара" в Tovar.xlsx.
4. **manufacturers** — извлечь уникальные из колонки "Производитель" в Tovar.xlsx.
5. **suppliers** — извлечь уникальные из колонки "Поставщик" в Tovar.xlsx.
6. **units** — извлечь уникальные из колонки "Единица измерения" в Tovar.xlsx (в текущих данных только "шт.").
7. **products** ← Tovar.xlsx — article из колонки "Артикул", FK по имени через справочники. Колонка "Фото" → путь к файлу в uploads/.
8. **pickup_points** ← Пункты выдачи_import.xlsx — файл не имеет заголовка, первая строка — это уже адрес. ID назначается по порядку строк (строка 1 → id=1, строка 2 → id=2, ...).
9. **order_statuses** — извлечь уникальные из колонки "Статус заказа" в Заказ_import.xlsx.
10. **orders** ← Заказ_import.xlsx — pickup_point_id = значение из колонки "Адрес пункта выдачи" (это числовой индекс, соответствует id в pickup_points). user_id определяется по ФИО через поиск в users. pickup_code из колонки "Код для получения".
11. **order_items** — парсинг колонки "Артикул заказа": формат `"АРТИКУЛ1, КОЛ-ВО1, АРТИКУЛ2, КОЛ-ВО2, ..."`. Разбить по запятым, взять пары (артикул, количество), найти product_id по артикулу.
12. **Фото** — `1.jpg` — `10.jpg` и `picture.png` копируются из `import/` в `backend/uploads/`.

### Обработка ошибок импорта
- Невалидные даты (например 30.02.2025 в order_date строки 7 заказов) → записать NULL в соответствующее поле, логировать предупреждение.
- Дублирующиеся артикулы → пропустить дубликат, логировать.
- Несуществующий артикул в заказе → пропустить позицию, логировать.
