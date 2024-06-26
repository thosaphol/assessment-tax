package postgres

import (
	"context"
	"time"
)

var dbTimeout = time.Second * 3

type Deduction struct {
	Personal    float64
	MaxKReceipt float64
}

func (p *Postgres) SetPersonalDeduction(amount float64) error {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	var dCount int

	row := p.Db.QueryRowContext(ctx, "SELECT count(*) as count FROM deductions;")
	err := row.Scan(&dCount)
	if err != nil {
		return err
	}

	var stmt string
	if dCount == 0 {
		stmt = "INSERT INTO deductions(personal) VALUES($1);"
	} else {
		stmt = "UPDATE deductions SET personal=$1;"
	}

	r, err := p.Db.ExecContext(ctx, stmt, amount)
	if err != nil {
		return err
	}

	if _, err := r.RowsAffected(); err != nil {
		return err
	}
	return nil
}

func (p *Postgres) PersonalDeduction() (float64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	row := p.Db.QueryRowContext(ctx, "SELECT personal,maximum_k_receipt FROM deductions")

	var d Deduction
	err := row.Scan(&d.Personal,
		&d.MaxKReceipt)

	if err != nil {
		return 0, err
	}

	return d.Personal, nil
}

func (p *Postgres) SetKReceiptDeduction(amount float64) error {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	var dCount int

	row := p.Db.QueryRowContext(ctx, "SELECT count(*) as count FROM deductions;")
	err := row.Scan(&dCount)
	if err != nil {
		return err
	}

	var stmt string
	if dCount == 0 {
		stmt = "INSERT INTO deductions(maximum_k_receipt) VALUES($1);"
	} else {
		stmt = "UPDATE deductions SET maximum_k_receipt=$1;"
	}

	r, err := p.Db.ExecContext(ctx, stmt, amount)
	if err != nil {
		return err
	}

	if _, err := r.RowsAffected(); err != nil {
		return err
	}
	return nil
}
func (p *Postgres) KReceiptDeduction() (float64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	row := p.Db.QueryRowContext(ctx, "SELECT maximum_k_receipt FROM deductions")

	var d float64
	err := row.Scan(&d)

	if err != nil {
		return 0, err
	}

	return d, nil
}
