-- =============================================
-- RECURRING TRANSACTION DETECTION
-- =============================================
-- Replace 1 with your workspace_id before running.
-- =============================================

WITH txn_with_gaps AS (
    SELECT
        t.id,
        t.counterparty_iban,
        ABS(t.total_amount)              AS amount,
        t.currency,
        t.description,
        t.date,
        t.transaction_type,
        t.bank_account_id,
        t.date - LAG(t.date) OVER w      AS gap_days,
        ROW_NUMBER() OVER (
            PARTITION BY t.counterparty_iban, t.currency
            ORDER BY t.date DESC, t.id DESC
        )                                 AS recency_rank
    FROM transactions t
    WHERE t.workspace_id = 1  -- ← your workspace_id here
      AND t.counterparty_iban IS NOT NULL
      AND t.counterparty_iban <> ''
      AND t.date >= CURRENT_DATE - INTERVAL '6 months'
    WINDOW w AS (
        PARTITION BY t.counterparty_iban, t.currency
        ORDER BY t.date, t.id
    )
),

stats AS (
    SELECT
        counterparty_iban,
        currency,
        COUNT(*)                                                AS occurrence_count,
        MIN(date)                                               AS first_seen,
        MAX(date)                                               AS last_seen,
        ROUND(AVG(amount), 2)                                   AS avg_amount,
        ROUND(MIN(amount), 2)                                   AS min_amount,
        ROUND(MAX(amount), 2)                                   AS max_amount,
        CASE WHEN AVG(amount) > 0
             THEN ROUND(COALESCE(STDDEV(amount), 0) / AVG(amount), 4)
             ELSE 0
        END                                                     AS amount_cv,
        ROUND(AVG(gap_days))::int                               AS avg_days_between,
        ROUND(COALESCE(STDDEV(gap_days), 0))::int               AS interval_stddev_days,
        (array_agg(id               ORDER BY recency_rank))[1]  AS latest_tx_id,
        (array_agg(description      ORDER BY recency_rank))[1]  AS latest_description,
        (array_agg(transaction_type  ORDER BY recency_rank))[1]  AS latest_type,
        (array_agg(amount            ORDER BY recency_rank))[1]  AS latest_amount,
        (array_agg(bank_account_id   ORDER BY recency_rank))[1]  AS latest_bank_account_id
    FROM txn_with_gaps
    GROUP BY counterparty_iban, currency
    HAVING COUNT(*) >= 2
)

SELECT
    s.counterparty_iban,
    s.currency,
    s.occurrence_count,
    s.first_seen,
    s.last_seen,
    s.avg_amount,
    s.min_amount,
    s.max_amount,
    s.amount_cv,
    s.avg_days_between,
    s.interval_stddev_days,
    s.latest_tx_id,
    s.latest_description,
    s.latest_type,
    s.latest_amount,
    s.latest_bank_account_id,

    CASE
        WHEN s.avg_days_between BETWEEN 5   AND 9    THEN 'weekly'
        WHEN s.avg_days_between BETWEEN 12  AND 18   THEN 'biweekly'
        WHEN s.avg_days_between BETWEEN 25  AND 35   THEN 'monthly'
        WHEN s.avg_days_between BETWEEN 55  AND 65   THEN 'bimonthly'
        WHEN s.avg_days_between BETWEEN 85  AND 100  THEN 'quarterly'
        WHEN s.avg_days_between BETWEEN 170 AND 195  THEN 'semiannual'
        WHEN s.avg_days_between BETWEEN 350 AND 380  THEN 'yearly'
        ELSE                                              'irregular'
    END AS estimated_frequency,

    s.last_seen + s.avg_days_between AS next_expected_date,

    ROUND(
        s.avg_amount * 30.0 / NULLIF(s.avg_days_between, 0),
        2
    ) AS monthly_cost,

    CASE
        WHEN s.occurrence_count >= 4
             AND s.interval_stddev_days <= GREATEST(s.avg_days_between * 0.3, 3)
        THEN 'high'
        WHEN s.occurrence_count >= 3
             AND s.interval_stddev_days <= GREATEST(s.avg_days_between * 0.5, 5)
        THEN 'medium'
        WHEN s.avg_days_between BETWEEN 5 AND 380
        THEN 'low'
        ELSE 'low'
    END AS confidence,

    CASE WHEN s.amount_cv < 0.05 THEN TRUE ELSE FALSE END AS is_fixed_amount,

    (
        SELECT e.category_id
        FROM entries e
        JOIN transactions t2 ON t2.id = e.transaction_id
        WHERE t2.counterparty_iban = s.counterparty_iban
          AND t2.currency          = s.currency
          AND t2.workspace_id      = 1  -- ← same workspace_id here
        ORDER BY t2.date DESC, t2.id DESC
        LIMIT 1
    ) AS latest_category_id

FROM stats s
WHERE s.avg_days_between IS NOT NULL
ORDER BY
    CASE
        WHEN s.occurrence_count >= 4
             AND s.interval_stddev_days <= GREATEST(s.avg_days_between * 0.3, 3)
        THEN 1
        WHEN s.occurrence_count >= 3
             AND s.interval_stddev_days <= GREATEST(s.avg_days_between * 0.5, 5)
        THEN 2
        ELSE 3
    END,
    ROUND(s.avg_amount * 30.0 / NULLIF(s.avg_days_between, 0), 2) DESC NULLS LAST;
