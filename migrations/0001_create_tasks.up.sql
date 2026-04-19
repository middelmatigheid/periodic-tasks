ALTER TABLE IF EXISTS tasks ADD COLUMN doctor_id BIGINT;
ALTER TABLE IF EXISTS tasks ADD COLUMN due_date TIMESTAMP;
ALTER TABLE IF EXISTS tasks ADD COLUMN generator_id BIGINT;

CREATE TABLE IF NOT EXISTS tasks (
	id BIGSERIAL PRIMARY KEY,
	doctor_id BIGINT,
	title TEXT NOT NULL,
	description TEXT NOT NULL DEFAULT '',
	status TEXT NOT NULL,
	due_date TIMESTAMP,
	generator_id BIGINT,
	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_tasks_status ON tasks (status);
CREATE INDEX IF NOT EXISTS idx_tasks_due_date ON tasks (due_date);
