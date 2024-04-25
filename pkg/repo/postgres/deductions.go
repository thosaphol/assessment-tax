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

// func (p *Postgres) Deduction() (*deduction.Deduction, error) {
// 	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
// 	defer cancel()

// 	row := p.Db.QueryRowContext(ctx, "SELECT personal,maximum_k_receipt FROM deductions")

// 	var d Deduction
// 	err := row.Scan(&d.Personal,
// 		&d.MaxKReceipt)

// 	if err != nil {
// 		return nil, err
// 	}

// 	return &deduction.Deduction{Personal: d.Personal, MaxKReceipt: d.MaxKReceipt}, nil
// }
