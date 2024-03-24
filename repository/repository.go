package repository

import (
	"context"
	"fmt"
	"os"

	"github.com/jackc/pgx/v5"
)

type repository interface {
	GetExpenseType()
	AddValuesDB()
	GetUserExpenseTypes()
	AddExpense()
	AddExpType()
	GetIdTypeExp()
	AddExpValue()
}

type ExpenseRepo struct {
	conn *pgx.Conn
	// tx1  pgx.Tx
}

func NewExpenseRepo(conn *pgx.Conn) *ExpenseRepo {
	return &ExpenseRepo{conn: conn}
}

// GetExpenseType gets one row of type of expenses from DB by name
// func (r *ExpenseRepo) GetExpenseType(name string) (*string, error) {

// 	var typeExpenses string
// 	err := r.conn.QueryRow(context.Background(),
// 		"SELECT title from expense_type, users where expense_type.users_id=users.id and users.name=$1", name).Scan(&typeExpenses)
// 	if err != nil {
// 		err = fmt.Errorf("unable to connect to database: %v", err)
// 		return nil, err
// 	}
// 	return &typeExpenses, nil
// }

// GetUserExpenseTypes gets all rows of type of expenses from DB by name
func (r *ExpenseRepo) GetUserExpenseTypes(name string) ([]string, error) {
	rows, _ := r.conn.Query(context.Background(), "SELECT title from expense_type, users where expense_type.users_id=users.id and users.name=$1", name)
	numbers, err := pgx.CollectRows(rows, pgx.RowTo[string])
	if err != nil {
		err = fmt.Errorf("unable to connect to database: %v", err)
		return nil, err
	}
	return numbers, nil
}

// GetManyRows gets all rows of type of expenses from DB by login
func (r *ExpenseRepo) GetExpenseTypesUser(ctx context.Context, login string) ([]string, error) {
	rows, _ := r.conn.Query(ctx, "SELECT title from expense_type, users where expense_type.users_id=users.id and users.login=$1", login)
	numbers, err := pgx.CollectRows(rows, pgx.RowTo[string])
	if err != nil {
		err = fmt.Errorf("unable to connect to database: %v", err)
		return nil, err
	}
	return numbers, nil
}

// IsExpenseTypeExists checking exist type of expense or not in a database
func (r *ExpenseRepo) IsExpenseTypeExists(ctx context.Context, expType *string) (bool, error) {
	rows, _ := r.conn.Query(ctx, "Select title from expense_type")
	numbers, err := pgx.CollectRows(rows, pgx.RowTo[string])
	existExpType := false
	if err != nil {
		return existExpType, err
	} else {
		for _, v := range numbers {
			if v == *expType {
				existExpType = true
				return existExpType, nil
			}

		}
		existExpType = false
	}
	return existExpType, nil
}

// AddExpType insert a new type of expenses in a table expense_type
func (r *ExpenseRepo) AddExpType(ctx context.Context, tx pgx.Tx, expType *string, userId int) error {
	_, err := tx.Exec(ctx, "Insert into expense_type(users_id,title) values ($1,$2)", userId, *expType)
	if err != nil {
		return err
	}
	return err
}

// GetIdTypeExp gets id of expense type
func (r *ExpenseRepo) GetIdTypeExp(ctx context.Context, tx pgx.Tx, expType *string) (*int, error) {
	var expTypeId int
	err := tx.QueryRow(ctx, "select id from expense_type where title=$1", *expType).Scan(&expTypeId)
	if err != nil {
		err = fmt.Errorf("QueryRow failed: %v", err)
		return nil, err
	}
	return &expTypeId, err
}

// AddExpValue adds new row in a expense table
func (r *ExpenseRepo) AddExpValue(ctx context.Context, tx pgx.Tx, expTypeId *int, timeSpent *string, spent *float64) error {
	// add a new row into table expense

	_, err := tx.Exec(ctx, "Insert into expense(expense_type_id,reated_at, spent_money) values ($1,$2,$3)", *expTypeId, *timeSpent, *spent)
	if err != nil {
		return err
	}
	fmt.Println("was added")
	return err
}

// AddExpTransaction checks existing type of expenses from command-line in a table, and adds new row to expense table by transaction
func (r *ExpenseRepo) AddExpTransaction(ctx context.Context, login *string, expType *string, timeSpent *string, spent *float64) error {
	// checking expType exists in a table expense_type or not
	existExpType, err := r.IsExpenseTypeExists(ctx, expType)
	if err != nil {
		return err
	}

	// begin transaction
	tx, err := r.conn.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	if !existExpType {
		var userId int
		loginValue := *login
		// by QueryRow gets user's id from table users by login
		err = r.conn.QueryRow(ctx,
			`SELECT id FROM users where login=$1`, loginValue).Scan(&userId)
		if err != nil {
			err = fmt.Errorf("QueryRow failed: %v", err)
			return err
		}
		// adding new type of expense to expense_type table
		err = r.AddExpType(ctx, tx, expType, userId)
		if err != nil {
			return err
		}
		// getting Id of new expense_type
		expId, err1 := r.GetIdTypeExp(ctx, tx, expType)
		if err1 != nil {
			return err1
		}
		// adding a new row into expense table
		err = r.AddExpValue(ctx, tx, expId, timeSpent, spent)
		if err != nil {
			return err
		}

	} else if existExpType {
		// getting Id of expType from expense_type
		expId, err1 := r.GetIdTypeExp(ctx, tx, expType)
		if err1 != nil {
			return err1
		}
		// adding a new row into expense table
		err = r.AddExpValue(ctx, tx, expId, timeSpent, spent)
		if err != nil {
			return err
		}

	}

	err = tx.Commit(ctx)
	if err != nil {
		return err
	}
	return err
}

// // AddValuesDB insert row to the table user
// func (r *ExpenseRepo) AddValuesDB(ctx context.Context) error {

// 	_, err := r.conn.Exec(ctx, "Insert into users(name,surname,login,pass,email) VALUES ('kolya','Bon','spiman','1243','er1@23.ru')")
// 	if err != nil {
// 		return err
// 	}
// 	return nil
// }

// ConnectToDB connects to DB
func ConnectToDB(ctx context.Context, myurl string) (*pgx.Conn, error) {
	conn, err := pgx.Connect(ctx, os.Getenv(myurl))
	if err != nil {
		err = fmt.Errorf("unable to connect to database: %v", err)
		return nil, err
	}
	return conn, nil
}

// AddExpense checking type of expenses in a database, if not exists - adds new type of expenses to a table expense_type
// after this adds a new row to a table expense
func (r *ExpenseRepo) AddExpense(ctx context.Context, login *string, expType *string, timeSpent *string, spent *float64) error {
	// getting all typies of expenses by login and checking expType there or not
	numbers, err := r.GetExpenseTypesUser(ctx, *login)
	if err != nil {
		return err
	} else {
		existExpType := false
		for _, v := range numbers {
			if v == *expType {
				existExpType = true
				break
			}
		}
		// if expType not exists in a table expense_type
		if !existExpType {
			var userId int
			loginValue := *login
			// by QueryRow gets user's id from table users by login
			err = r.conn.QueryRow(ctx,
				`SELECT id 
			FROM 
			users 
			where login=$1`, loginValue).Scan(&userId)
			if err != nil {
				err = fmt.Errorf("QueryRow failed: %v", err)
				return err
			}
			// begin transaction
			tx, err := r.conn.Begin(ctx)
			if err != nil {
				return err
			}
			defer tx.Rollback(ctx)
			// insert a new type of expenses in a table expense_type
			_, err = tx.Exec(ctx, "Insert into expense_type(users_id,title) values ($1,$2)", userId, *expType)
			if err != nil {
				return err
			}
			// by QueryRow gets id expType from a table expense_type
			var expTypeId int
			err = r.conn.QueryRow(ctx, "select id from expense_type where title=$1", *expType).Scan(&expTypeId)
			if err != nil {
				err = fmt.Errorf("QueryRow failed: %v", err)
				return err
			}
			_, err = tx.Exec(ctx, "Insert into expense(expense_type_id,reated_at, spent_money) values ($1,$2,$3)", expTypeId, *timeSpent, *spent)
			if err != nil {
				return err
			}

			err = tx.Commit(ctx)
			if err != nil {
				return err
			}
			// fmt.Println("new expense was added")
			// if expType exist in a table expense_type
		} else if existExpType {
			var expTypeId int
			// by QueryRow gets id of expType from type of expense
			err = r.conn.QueryRow(ctx, "select id from expense_type where title=$1", *expType).Scan(&expTypeId)
			if err != nil {
				err = fmt.Errorf("QueryRow failed: %v", err)
				return err
			}
			_, err := r.conn.Exec(ctx, "Insert into expense(expense_type_id,reated_at, spent_money) values ($1,$2,$3)", expTypeId, *timeSpent, *spent)
			return err
		}

	}
	return nil
}
