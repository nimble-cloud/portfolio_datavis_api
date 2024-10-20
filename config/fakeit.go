package config

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jaswdr/faker/v2"
)

// const U = "UPDATE public.company_order_details SET account_name = $1 WHERE account_name = $2"

func FakeNames() error {
	fmt.Println("fakin")
	f := faker.New()
	// ctx, cancel := context.WithTimeout(context.Background(), 29*time.Second)
	// defer cancel()
	ctx := context.Background()
	tx, err := DB.Begin(ctx)
	if err != nil {
		tx.Rollback(ctx)
		return err
	}

	_, err = tx.Exec(ctx, "CREATE TEMPORARY TABLE faker (cur character varying(300), new character varying(300))")
	if err != nil {
		tx.Rollback(ctx)
		return err
	}

	copyRows := [][]any{}
	rows, err := DB.Query(ctx, "SELECT DISTINCT(account_name) FROM public.company_order_details")
	if err != nil {
		tx.Rollback(ctx)
		return err
	}

	for rows.Next() {
		account_name := ""
		err = rows.Scan(&account_name)
		if err != nil {
			tx.Rollback(ctx)
			return err
		}

		copyRows = append(copyRows, []any{account_name, f.Company().Name()})
	}

	copyCount, err := tx.CopyFrom(
		ctx,
		pgx.Identifier{"faker"},
		[]string{"cur", "new"},
		pgx.CopyFromRows(copyRows),
	)
	if err != nil {
		tx.Rollback(ctx)
		return err
	}

	fmt.Printf("copied %d rows\n", copyCount)

	// update the values
	_, err = tx.Exec(ctx, `
	UPDATE public.company_order_details AS d
	SET account_name = f.new
	FROM faker f
	WHERE d.account_name = f.cur
	`)

	if err != nil {
		tx.Rollback(ctx)
		return err
	}

	err = tx.Commit(ctx)
	if err != nil {
		tx.Rollback(ctx)
		return err
	}

	fmt.Println("faked")
	return nil
}
