CREATE TABLE IF NOT EXISTS task_generators (
	id BIGSERIAL PRIMARY KEY,
    doctor_id BIGINT NOT NULL,
	title TEXT NOT NULL,
	description TEXT NOT NULL DEFAULT '',
	status TEXT NOT NULL,
	every_n_days SMALLINT,
	every_ith_day SMALLINT,
	parity BOOLEAN,
	next_due_date TIMESTAMP NOT NULL,
	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	processed_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_task_generators_next_due_date ON task_generators (next_due_date);
CREATE INDEX IF NOT EXISTS idx_task_generators_processed_at_next_due_date ON task_generators (processed_at, next_due_date);
