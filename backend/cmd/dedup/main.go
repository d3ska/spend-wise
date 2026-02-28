package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"

	"backend/config"
	"backend/store"

	"github.com/joho/godotenv"
)

func main() {
	apply := flag.Bool("apply", false, "actually delete duplicates (default is dry-run)")
	flag.Parse()

	if err := run(*apply); err != nil {
		slog.Error("dedup failed", "error", err)
		os.Exit(1)
	}
}

type duplicate struct {
	KeepID      int64
	DeleteID    int64
	WorkspaceID int64
	Description string
	Amount      string
	Currency    string
	KeepDate    string
	DeleteDate  string
}

func run(apply bool) error {
	_ = godotenv.Load()

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	ctx := context.Background()
	pool, err := store.NewPool(ctx, cfg.Database)
	if err != nil {
		return fmt.Errorf("connecting to database: %w", err)
	}
	defer pool.Close()

	// Find duplicate bank transactions:
	// Same workspace, bank account, amount, currency, case-insensitive description,
	// dates within 2 days. Keep the one with the smaller ID (first imported).
	query := `
		SELECT
			t1.id        AS keep_id,
			t2.id        AS delete_id,
			t1.workspace_id,
			t1.description,
			t1.total_amount::text,
			t1.currency,
			t1.date::text AS keep_date,
			t2.date::text AS delete_date
		FROM transactions t1
		JOIN transactions t2
		  ON t1.workspace_id = t2.workspace_id
		 AND t1.bank_account_id IS NOT NULL
		 AND t2.bank_account_id IS NOT NULL
		 AND t1.bank_account_id = t2.bank_account_id
		 AND t1.total_amount = t2.total_amount
		 AND t1.currency = t2.currency
		 AND t1.source = 'bank'
		 AND t2.source = 'bank'
		 AND LOWER(t1.description) = LOWER(t2.description)
		 AND ABS(t1.date - t2.date) BETWEEN 1 AND 2
		 AND t1.id < t2.id
		ORDER BY t1.workspace_id, t1.id
	`

	rows, err := pool.Query(ctx, query)
	if err != nil {
		return fmt.Errorf("querying duplicates: %w", err)
	}
	defer rows.Close()

	var duplicates []duplicate
	for rows.Next() {
		var d duplicate
		if err := rows.Scan(&d.KeepID, &d.DeleteID, &d.WorkspaceID, &d.Description, &d.Amount, &d.Currency, &d.KeepDate, &d.DeleteDate); err != nil {
			return fmt.Errorf("scanning row: %w", err)
		}
		duplicates = append(duplicates, d)
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("iterating rows: %w", err)
	}

	if len(duplicates) == 0 {
		fmt.Println("No duplicate bank transactions found.")
		return nil
	}

	fmt.Printf("Found %d duplicate(s):\n\n", len(duplicates))
	for _, d := range duplicates {
		fmt.Printf("  KEEP   #%-6d  %s  %s %s  %q\n", d.KeepID, d.KeepDate, d.Amount, d.Currency, d.Description)
		fmt.Printf("  DELETE #%-6d  %s  %s %s  %q\n\n", d.DeleteID, d.DeleteDate, d.Amount, d.Currency, d.Description)
	}

	if !apply {
		fmt.Println("Dry run — no changes made. Re-run with --apply to delete duplicates.")
		return nil
	}

	// Delete duplicates (entries cascade-delete automatically).
	deleteIDs := make([]int64, len(duplicates))
	for i, d := range duplicates {
		deleteIDs[i] = d.DeleteID
	}

	tag, err := pool.Exec(ctx,
		`DELETE FROM transactions WHERE id = ANY($1)`,
		deleteIDs,
	)
	if err != nil {
		return fmt.Errorf("deleting duplicates: %w", err)
	}

	fmt.Printf("Deleted %d duplicate transaction(s).\n", tag.RowsAffected())
	return nil
}
