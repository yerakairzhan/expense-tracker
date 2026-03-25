# API Endpoints (v1)

## Conventions

- Base URL: `http://localhost:8080`
- API prefix: `/api/v1`
- Auth header for protected endpoints: `Authorization: Bearer <access_token>`
- Content type: `application/json`
- Date format: `YYYY-MM-DD`
- Amount format: decimal string with up to 4 fraction digits, e.g. `"1250.0000"`

## Auth

| Method | Endpoint | Auth | Description |
|---|---|---|---|
| POST | `/api/v1/auth/register` | public | Register and return access/refresh tokens |
| POST | `/api/v1/auth/login` | public | Login and return access/refresh tokens |
| POST | `/api/v1/auth/refresh` | public | Rotate refresh token |
| POST | `/api/v1/auth/logout` | JWT | Revoke refresh token |

### Register request

```json
{
  "email": "user@example.com",
  "password": "secure_password",
  "name": "Alex",
  "currency": "USD"
}
```

### Auth response

```json
{
  "access_token": "<jwt>",
  "refresh_token": "<token>",
  "expires_in": 900
}
```

## Users

| Method | Endpoint | Auth | Description |
|---|---|---|---|
| GET | `/api/v1/users/me` | JWT | Get own profile |
| PATCH | `/api/v1/users/me` | JWT | Update name/currency |
| PATCH | `/api/v1/users/me/password` | JWT | Change password |

## Accounts

| Method | Endpoint | Auth | Description |
|---|---|---|---|
| GET | `/api/v1/accounts` | JWT | List user accounts |
| POST | `/api/v1/accounts` | JWT | Create account |
| GET | `/api/v1/accounts/:id` | JWT | Get account |
| PATCH | `/api/v1/accounts/:id` | JWT | Update account |
| DELETE | `/api/v1/accounts/:id` | JWT | Soft-delete account |

### Create account request

```json
{
  "name": "Kaspi card",
  "account_type": "bank_card",
  "currency": "KZT",
  "balance": "50000.0000"
}
```

Allowed `account_type`: `cash`, `bank_card`, `e_wallet`.

## Transactions

| Method | Endpoint | Auth | Description |
|---|---|---|---|
| GET | `/api/v1/transactions` | JWT | List transactions with filters |
| POST | `/api/v1/transactions` | JWT | Create transaction |
| GET | `/api/v1/transactions/:id` | JWT | Get transaction |
| PATCH | `/api/v1/transactions/:id` | JWT | Update amount/category/notes |
| DELETE | `/api/v1/transactions/:id` | JWT | Soft-delete transaction |

### List query params

| Param | Type | Description |
|---|---|---|
| `account_id` | int | Filter by account |
| `category_id` | int | Filter by category |
| `type` | string | `income`, `expense`, `transfer` |
| `from` | date | Start date inclusive |
| `to` | date | End date inclusive |
| `page` | int | Default `1` |
| `limit` | int | Default `20`, max `100` |

### Create transaction request

```json
{
  "account_id": 1,
  "category_id": 3,
  "amount": "4500.0000",
  "currency": "KZT",
  "type": "expense",
  "description": "Groceries",
  "notes": "Weekend shopping",
  "transacted_at": "2026-03-25"
}
```

Rules:
- `amount` is always positive.
- Direction comes from `type`.

## Health

| Method | Endpoint | Auth | Description |
|---|---|---|---|
| GET | `/health` | public | Liveness |
| GET | `/health/ready` | public | Readiness (DB ping) |

## Error Envelope

All errors follow:

```json
{
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "amount is required"
  }
}
```

### Error codes

| HTTP | Code |
|---|---|
| 400 | `VALIDATION_ERROR` |
| 401 | `UNAUTHORIZED` |
| 403 | `FORBIDDEN` |
| 404 | `NOT_FOUND` |
| 409 | `CONFLICT` |
| 422 | `INSUFFICIENT_FUNDS` |
| 500 | `INTERNAL_ERROR` |
