# ООО "Обувь" — Информационная система

Веб-приложение для управления товарами и заказами магазина обуви.

**Стек:** Go (chi, SQLite) + Next.js (Ant Design, TypeScript)

## Запуск через Docker (рекомендуется)

```bash
docker-compose up -d
```

- Frontend: http://localhost:3000
- Backend API: http://localhost:8080

Seed данных из Excel запускается автоматически при первом старте.

## Запуск без Docker

**Backend:**

```bash
cd backend
go run ./cmd/seed      # импорт данных (один раз)
go run ./cmd/server    # запуск сервера на :8080
```

**Frontend:**

```bash
cd frontend
npm install
npm run dev            # запуск на :3000
```

## Линтеры

```bash
cd backend && golangci-lint run ./...
cd frontend && npx eslint src/
```

## Учётные записи (из seed)

| Роль | Логин | Пароль |
|------|-------|--------|
| Администратор | 94d5ous@gmail.com | uzWC67 |
| Менеджер | 2g1asu@mail.com | 2wNtAn |
| Клиент | 6dcsmq@outlook.com | 8ntwUp |
