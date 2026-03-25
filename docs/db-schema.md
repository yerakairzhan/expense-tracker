# Database Schema

Current schema contains 7 tables:

1. `users`
2. `refresh_tokens`
3. `accounts`
4. `categories`
5. `transactions`
6. `budgets`
7. `recurring_payments`

## users

| Column | Type | Constraints |
|---|---|---|
| id | bigserial | PK |
| email | text | NOT NULL, UNIQUE |
| password_hash | text | NOT NULL |
| name | text | NOT NULL |
| currency | text | NOT NULL, default `'USD'` |
| is_active | bool | NOT NULL, default `true` |
| created_at | timestamptz | NOT NULL, default `now()` |
| updated_at | timestamptz | NOT NULL, default `now()` |
| deleted_at | timestamptz | nullable |

## refresh_tokens

| Column | Type | Constraints |
|---|---|---|
| id | bigserial | PK |
| user_id | bigint | NOT NULL, FK → `users.id` |
| token_hash | text | NOT NULL, UNIQUE (bcrypt hash) |
| expires_at | timestamptz | NOT NULL |
| revoked | bool | NOT NULL, default `false` |
| created_at | timestamptz | NOT NULL, default `now()` |

## accounts

| Column | Type | Constraints |
|---|---|---|
| id | bigserial | PK |
| user_id | bigint | NOT NULL, FK → `users.id` |
| name | text | NOT NULL |
| account_type | text | NOT NULL, `cash \| bank_card \| e_wallet` |
| balance | numeric(15,4) | NOT NULL, default `0` |
| currency | text | NOT NULL, default `'USD'` |
| is_active | bool | NOT NULL, default `true` |
| created_at | timestamptz | NOT NULL, default `now()` |
| updated_at | timestamptz | NOT NULL, default `now()` |
| deleted_at | timestamptz | nullable (soft-delete) |

## categories

| Column | Type | Constraints |
|---|---|---|
| id | bigserial | PK |
| user_id | bigint | nullable, FK → `users.id` |
| name | text | NOT NULL |
| type | text | NOT NULL, `income \| expense` |
| color | text | nullable |
| icon | text | nullable |
| is_system | bool | NOT NULL, default `false` |
| created_at | timestamptz | NOT NULL, default `now()` |
| updated_at | timestamptz | NOT NULL, default `now()` |

`user_id` is `NULL` for system categories.

## transactions

| Column | Type | Constraints |
|---|---|---|
| id | bigserial | PK |
| account_id | bigint | NOT NULL, FK → `accounts.id` |
| category_id | bigint | nullable, FK → `categories.id` |
| recurring_id | bigint | nullable, FK → `recurring_payments.id` |
| amount | numeric(15,4) | NOT NULL, must be positive |
| currency | text | NOT NULL |
| type | text | NOT NULL, `income \| expense \| transfer` |
| description | text | NOT NULL |
| notes | text | nullable |
| transacted_at | date | NOT NULL |
| created_at | timestamptz | NOT NULL, default `now()` |
| updated_at | timestamptz | NOT NULL, default `now()` |
| deleted_at | timestamptz | nullable (soft-delete) |

## budgets

| Column | Type | Constraints |
|---|---|---|
| id | bigserial | PK |
| user_id | bigint | NOT NULL, FK → `users.id` |
| category_id | bigint | nullable, FK → `categories.id` |
| limit_amount | numeric(15,4) | NOT NULL |
| currency | text | NOT NULL |
| period | text | NOT NULL, `monthly \| weekly \| yearly \| custom` |
| starts_at | date | NOT NULL |
| ends_at | date | nullable |
| is_active | bool | NOT NULL, default `true` |
| created_at | timestamptz | NOT NULL, default `now()` |
| updated_at | timestamptz | NOT NULL, default `now()` |

## recurring_payments

| Column | Type | Constraints |
|---|---|---|
| id | bigserial | PK |
| user_id | bigint | NOT NULL, FK → `users.id` |
| account_id | bigint | NOT NULL, FK → `accounts.id` |
| category_id | bigint | nullable, FK → `categories.id` |
| title | text | NOT NULL |
| amount | numeric(15,4) | NOT NULL |
| currency | text | NOT NULL |
| frequency | text | NOT NULL, `daily \| weekly \| monthly \| yearly` |
| next_run_at | date | NOT NULL |
| ends_at | date | nullable |
| is_active | bool | NOT NULL, default `true` |
| created_at | timestamptz | NOT NULL, default `now()` |
| updated_at | timestamptz | NOT NULL, default `now()` |

## FK Behavior

| FK | On delete |
|---|---|
| `refresh_tokens.user_id → users.id` | CASCADE |
| `accounts.user_id → users.id` | CASCADE |
| `transactions.account_id → accounts.id` | CASCADE |
| `transactions.category_id → categories.id` | SET NULL |
| `transactions.recurring_id → recurring_payments.id` | SET NULL |
| `budgets.user_id → users.id` | CASCADE |
| `budgets.category_id → categories.id` | SET NULL |
| `recurring_payments.user_id → users.id` | CASCADE |
| `recurring_payments.account_id → accounts.id` | RESTRICT |
| `recurring_payments.category_id → categories.id` | SET NULL |
