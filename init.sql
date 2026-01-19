CREATE TABLE IF NOT EXISTS histories (
    message_order BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    user_id TEXT NOT NULL,
    content TEXT NOT NULL,
    id TEXT NOT NULL,
    role TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS accounts (
    email TEXT NOT NULL,
    password TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    instaddr_id TEXT NOT NULL,
    instaddr_password TEXT NOT NULL,
    activate_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
