-- Align schema with docs/db-schema (v1 core + forward-compatible tables)

-- users
ALTER TABLE users
    ALTER COLUMN id TYPE BIGINT,
    ALTER COLUMN email TYPE TEXT,
    ALTER COLUMN name TYPE TEXT,
    ALTER COLUMN created_at TYPE TIMESTAMPTZ USING created_at AT TIME ZONE 'UTC',
    ALTER COLUMN updated_at TYPE TIMESTAMPTZ USING updated_at AT TIME ZONE 'UTC';

ALTER TABLE users
    ADD COLUMN IF NOT EXISTS password_hash TEXT NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS currency TEXT NOT NULL DEFAULT 'USD',
    ADD COLUMN IF NOT EXISTS is_active BOOLEAN NOT NULL DEFAULT TRUE,
    ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMPTZ;

-- accounts
ALTER TABLE accounts
    ALTER COLUMN id TYPE BIGINT,
    ALTER COLUMN user_id TYPE BIGINT,
    ALTER COLUMN account_type TYPE TEXT,
    ALTER COLUMN balance TYPE NUMERIC(15,4),
    ALTER COLUMN currency TYPE TEXT,
    ALTER COLUMN currency SET DEFAULT 'USD',
    ALTER COLUMN created_at TYPE TIMESTAMPTZ USING created_at AT TIME ZONE 'UTC',
    ALTER COLUMN updated_at TYPE TIMESTAMPTZ USING updated_at AT TIME ZONE 'UTC';

ALTER TABLE accounts
    ADD COLUMN IF NOT EXISTS name TEXT NOT NULL DEFAULT 'Main account',
    ADD COLUMN IF NOT EXISTS is_active BOOLEAN NOT NULL DEFAULT TRUE,
    ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMPTZ;

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint
        WHERE conname = 'accounts_account_type_check'
    ) THEN
        ALTER TABLE accounts
            ADD CONSTRAINT accounts_account_type_check
            CHECK (account_type IN ('cash', 'bank_card', 'e_wallet'));
    END IF;
END $$;

-- categories
CREATE TABLE IF NOT EXISTS categories (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT REFERENCES users(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    type TEXT NOT NULL CHECK (type IN ('income', 'expense')),
    color TEXT,
    icon TEXT,
    is_system BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- recurring_payments (v3 table included by schema doc)
CREATE TABLE IF NOT EXISTS recurring_payments (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    account_id BIGINT NOT NULL REFERENCES accounts(id) ON DELETE RESTRICT,
    category_id BIGINT REFERENCES categories(id) ON DELETE SET NULL,
    title TEXT NOT NULL,
    amount NUMERIC(15,4) NOT NULL CHECK (amount > 0),
    currency TEXT NOT NULL,
    frequency TEXT NOT NULL CHECK (frequency IN ('daily', 'weekly', 'monthly', 'yearly')),
    next_run_at DATE NOT NULL,
    ends_at DATE,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- transactions
ALTER TABLE transactions
    ALTER COLUMN id TYPE BIGINT,
    ALTER COLUMN account_id TYPE BIGINT,
    ALTER COLUMN amount TYPE NUMERIC(15,4),
    ALTER COLUMN description TYPE TEXT,
    ALTER COLUMN created_at TYPE TIMESTAMPTZ USING created_at AT TIME ZONE 'UTC';

DO $$
BEGIN
    IF EXISTS (
        SELECT 1
        FROM information_schema.columns
        WHERE table_name = 'transactions' AND column_name = 'transaction_type'
    ) THEN
        ALTER TABLE transactions RENAME COLUMN transaction_type TO type;
    END IF;
END $$;

ALTER TABLE transactions
    ADD COLUMN IF NOT EXISTS category_id BIGINT REFERENCES categories(id) ON DELETE SET NULL,
    ADD COLUMN IF NOT EXISTS recurring_id BIGINT REFERENCES recurring_payments(id) ON DELETE SET NULL,
    ADD COLUMN IF NOT EXISTS currency TEXT NOT NULL DEFAULT 'USD',
    ADD COLUMN IF NOT EXISTS type TEXT NOT NULL DEFAULT 'expense',
    ADD COLUMN IF NOT EXISTS notes TEXT,
    ADD COLUMN IF NOT EXISTS transacted_at DATE NOT NULL DEFAULT CURRENT_DATE,
    ADD COLUMN IF NOT EXISTS updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMPTZ;

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint
        WHERE conname = 'transactions_type_check'
    ) THEN
        ALTER TABLE transactions
            ADD CONSTRAINT transactions_type_check
            CHECK (type IN ('income', 'expense', 'transfer'));
    END IF;
END $$;

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint
        WHERE conname = 'transactions_amount_positive_check'
    ) THEN
        ALTER TABLE transactions
            ADD CONSTRAINT transactions_amount_positive_check
            CHECK (amount > 0);
    END IF;
END $$;

-- refresh_tokens
CREATE TABLE IF NOT EXISTS refresh_tokens (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_hash TEXT NOT NULL UNIQUE,
    expires_at TIMESTAMPTZ NOT NULL,
    revoked BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- budgets (v2 table included by schema doc)
CREATE TABLE IF NOT EXISTS budgets (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    category_id BIGINT REFERENCES categories(id) ON DELETE SET NULL,
    limit_amount NUMERIC(15,4) NOT NULL,
    currency TEXT NOT NULL,
    period TEXT NOT NULL CHECK (period IN ('monthly', 'weekly', 'yearly', 'custom')),
    starts_at DATE NOT NULL,
    ends_at DATE,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Useful indexes
CREATE INDEX IF NOT EXISTS idx_users_email_active ON users(email) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_accounts_user_active ON accounts(user_id) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_transactions_account_active ON transactions(account_id) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_transactions_transacted_at ON transactions(transacted_at);
CREATE INDEX IF NOT EXISTS idx_refresh_tokens_user ON refresh_tokens(user_id);

-- Seed common system categories if empty
INSERT INTO categories (name, type, color, icon, is_system)
SELECT s.name, s.type, s.color, s.icon, TRUE
FROM (
    VALUES
        ('Salary', 'income', '#2E7D32', 'wallet'),
        ('Investments', 'income', '#1565C0', 'chart'),
        ('Food', 'expense', '#EF6C00', 'utensils'),
        ('Transport', 'expense', '#6A1B9A', 'car'),
        ('Utilities', 'expense', '#455A64', 'bolt')
) AS s(name, type, color, icon)
WHERE NOT EXISTS (SELECT 1 FROM categories WHERE is_system = TRUE);
