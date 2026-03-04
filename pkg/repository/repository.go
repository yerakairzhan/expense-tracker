package repository

import (
	"context"
	"database/sql"
	"github.com/example/financial-intelligence-platform/pkg/models"
)

// UserRepository handles user-related database operations
type UserRepository struct {
	db *sql.DB
}

// NewUserRepository creates a new UserRepository instance
func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

// CreateUser inserts a new user into the database
func (ur *UserRepository) CreateUser(ctx context.Context, email, name string) (*models.User, error) {
	var user models.User
	
	err := ur.db.QueryRowContext(
		ctx,
		`INSERT INTO users (email, name) VALUES ($1, $2) 
		 RETURNING id, email, name, created_at, updated_at`,
		email, name,
	).Scan(&user.ID, &user.Email, &user.Name, &user.CreatedAt, &user.UpdatedAt)
	
	if err != nil {
		return nil, err
	}
	
	return &user, nil
}

// GetUserByID retrieves a user by ID
func (ur *UserRepository) GetUserByID(ctx context.Context, id int) (*models.User, error) {
	var user models.User
	
	err := ur.db.QueryRowContext(
		ctx,
		`SELECT id, email, name, created_at, updated_at FROM users WHERE id = $1`,
		id,
	).Scan(&user.ID, &user.Email, &user.Name, &user.CreatedAt, &user.UpdatedAt)
	
	if err != nil {
		return nil, err
	}
	
	return &user, nil
}

// AccountRepository handles account-related database operations
type AccountRepository struct {
	db *sql.DB
}

// NewAccountRepository creates a new AccountRepository instance
func NewAccountRepository(db *sql.DB) *AccountRepository {
	return &AccountRepository{db: db}
}

// CreateAccount inserts a new account into the database
func (ar *AccountRepository) CreateAccount(ctx context.Context, userID int, accountType, balance, currency string) (*models.Account, error) {
	var account models.Account
	
	err := ar.db.QueryRowContext(
		ctx,
		`INSERT INTO accounts (user_id, account_type, balance, currency) 
		 VALUES ($1, $2, $3, $4) 
		 RETURNING id, user_id, account_type, balance, currency, created_at, updated_at`,
		userID, accountType, balance, currency,
	).Scan(&account.ID, &account.UserID, &account.AccountType, &account.Balance, &account.Currency, &account.CreatedAt, &account.UpdatedAt)
	
	if err != nil {
		return nil, err
	}
	
	return &account, nil
}

// GetAccountsByUserID retrieves all accounts for a user
func (ar *AccountRepository) GetAccountsByUserID(ctx context.Context, userID int) ([]models.Account, error) {
	rows, err := ar.db.QueryContext(
		ctx,
		`SELECT id, user_id, account_type, balance, currency, created_at, updated_at 
		 FROM accounts WHERE user_id = $1 ORDER BY created_at DESC`,
		userID,
	)
	
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var accounts []models.Account
	for rows.Next() {
		var acc models.Account
		err := rows.Scan(&acc.ID, &acc.UserID, &acc.AccountType, &acc.Balance, &acc.Currency, &acc.CreatedAt, &acc.UpdatedAt)
		if err != nil {
			return nil, err
		}
		accounts = append(accounts, acc)
	}
	
	return accounts, rows.Err()
}

// TransactionRepository handles transaction-related database operations
type TransactionRepository struct {
	db *sql.DB
}

// NewTransactionRepository creates a new TransactionRepository instance
func NewTransactionRepository(db *sql.DB) *TransactionRepository {
	return &TransactionRepository{db: db}
}

// CreateTransaction inserts a new transaction into the database
func (tr *TransactionRepository) CreateTransaction(ctx context.Context, accountID int, amount, description, txType string) (*models.Transaction, error) {
	var tx models.Transaction
	
	err := tr.db.QueryRowContext(
		ctx,
		`INSERT INTO transactions (account_id, amount, description, transaction_type) 
		 VALUES ($1, $2, $3, $4) 
		 RETURNING id, account_id, amount, description, transaction_type, created_at`,
		accountID, amount, description, txType,
	).Scan(&tx.ID, &tx.AccountID, &tx.Amount, &tx.Description, &tx.TransactionType, &tx.CreatedAt)
	
	if err != nil {
		return nil, err
	}
	
	return &tx, nil
}

// ListTransactionsByAccountID retrieves transactions for an account
func (tr *TransactionRepository) ListTransactionsByAccountID(ctx context.Context, accountID, limit, offset int) ([]models.Transaction, error) {
	rows, err := tr.db.QueryContext(
		ctx,
		`SELECT id, account_id, amount, description, transaction_type, created_at 
		 FROM transactions WHERE account_id = $1 
		 ORDER BY created_at DESC LIMIT $2 OFFSET $3`,
		accountID, limit, offset,
	)
	
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var transactions []models.Transaction
	for rows.Next() {
		var tx models.Transaction
		err := rows.Scan(&tx.ID, &tx.AccountID, &tx.Amount, &tx.Description, &tx.TransactionType, &tx.CreatedAt)
		if err != nil {
			return nil, err
		}
		transactions = append(transactions, tx)
	}
	
	return transactions, rows.Err()
}
