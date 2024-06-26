package postgres

import (
	"context"
	"github.com/certified-juniors/AtomHack/internal/domain"
	logs "github.com/certified-juniors/AtomHack/internal/logger"

	"github.com/jackc/pgx/v5"
)

const getByEmailQuery = `
	SELECT id, email, password, name, surname, middle_name, role, confirmed
	FROM "user"
	WHERE email = $1
`

const getByIdQuery = `
	SELECT id, email, name, surname, middle_name, role, confirmed
	FROM "user"
	WHERE id = $1
`

const updateConfirmedQuery = `
	UPDATE "user"
	SET confirmed = true
	WHERE id = $1
	RETURNING email
`

const addUserQuery = `
	INSERT INTO "user" (email, password, name, surname, middle_name, role)
	VALUES ($1, $2, $3, $4, $5, $6)
	RETURNING id
`

const userExistsQuery = `
	SELECT EXISTS(SELECT 1
				  FROM "user"
				  WHERE email = $1)
`

type authPostgresqlRepository struct {
	db  domain.PgxPoolIface
	ctx context.Context
}

func NewAuthPostgresqlRepository(pool domain.PgxPoolIface, ctx context.Context) domain.AuthRepository {
	return &authPostgresqlRepository{
		db:  pool,
		ctx: ctx,
	}
}

func (r *authPostgresqlRepository) GetByEmail(email string) (domain.User, error) {
	result := r.db.QueryRow(r.ctx, getByEmailQuery, email)

	logs.Logger.Debug("GetByEmail query result:", result)

	var user domain.User
	err := result.Scan(
		&user.ID,
		&user.Email,
		&user.Password,
		&user.Name,
		&user.Surname,
		&user.MiddleName,
		&user.Role,
		&user.Confirmed,
	)

	if err == pgx.ErrNoRows {
		logs.LogError(logs.Logger, "auth/postgres", "GetByEmail", err, err.Error())
		return domain.User{}, domain.ErrNotFound
	}
	if err != nil {
		logs.LogError(logs.Logger, "auth/postgres", "GetByEmail", err, err.Error())
		return domain.User{}, err
	}

	return user, nil
}

func (r *authPostgresqlRepository) GetByID(id int) (domain.User, error) {
	result := r.db.QueryRow(r.ctx, getByIdQuery, id)

	logs.Logger.Debug("GetByEmail query result:", result)

	var user domain.User
	err := result.Scan(
		&user.ID,
		&user.Email,
		&user.Name,
		&user.Surname,
		&user.MiddleName,
		&user.Role,
		&user.Confirmed,
	)

	if err == pgx.ErrNoRows {
		logs.LogError(logs.Logger, "auth/postgres", "GetByID", err, err.Error())
		return domain.User{}, domain.ErrNotFound
	}
	if err != nil {
		logs.LogError(logs.Logger, "auth/postgres", "GetByID", err, err.Error())
		return domain.User{}, err
	}

	return user, nil
}

func (r *authPostgresqlRepository) AddUser(user domain.User) (int, error) {
	if user.Email == "" || len(user.Password) == 0 {
		return 0, domain.ErrBadRequest
	}

	result := r.db.QueryRow(r.ctx, addUserQuery,
		&user.Email,
		&user.Password,
		&user.Name,
		&user.Surname,
		&user.MiddleName,
		&user.Role,
	)

	logs.Logger.Debug("AddUser queryRow result:", result)

	var id int
	if err := result.Scan(&id); err != nil {
		logs.LogError(logs.Logger, "auth_postgres", "AddUser", err, err.Error())
		return 0, err
	}
	return id, nil
}

func (r *authPostgresqlRepository) UserExists(email string) (bool, error) {
	result := r.db.QueryRow(r.ctx, userExistsQuery, email)

	var exist bool
	if err := result.Scan(&exist); err != nil {
		logs.LogError(logs.Logger, "auth_postgres", "UserExists", err, err.Error())
		return false, err
	}

	return exist, nil
}

func (r *authPostgresqlRepository) ConfirmUser(id int) (string, error) {
	if id == 0 {
		return "", domain.ErrBadRequest
	}

	_, err := r.db.Exec(r.ctx, updateConfirmedQuery, id)
	if err != nil {
		logs.LogError(logs.Logger, "auth/postgresql", "ConfirmUser", err, err.Error())
		return "", err
	}

	result := r.db.QueryRow(r.ctx, updateConfirmedQuery, id)

	logs.Logger.Debug("updateConfirmedQuery queryRow result:", result)

	var email string
	if err := result.Scan(&email); err != nil {
		logs.LogError(logs.Logger, "auth/postgres", "ConfirmUser", err, err.Error())
		return "", err
	}
	return email, nil
}
