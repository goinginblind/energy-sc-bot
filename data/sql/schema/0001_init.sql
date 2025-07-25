-- +goose Up
-- users table
CREATE TABLE users (
    id              BIGSERIAL PRIMARY KEY,
    telegram_id     BIGINT    NOT NULL UNIQUE,
    phone           TEXT,
    email           TEXT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- OTP requests (одноразовый пароль для входа)
CREATE TABLE otp_requests (
    id          BIGSERIAL PRIMARY KEY,
    user_id     BIGINT        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    otp_code    TEXT          NOT NULL,
    expires_at  TIMESTAMPTZ   NOT NULL,
    created_at  TIMESTAMPTZ   NOT NULL DEFAULT now()
);
CREATE INDEX ON otp_requests(user_id);

-- sessions (трек истечения сессии)
CREATE TABLE sessions (
    id          BIGSERIAL PRIMARY KEY,
    user_id     BIGINT        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token       TEXT          NOT NULL UNIQUE,
    expires_at  TIMESTAMPTZ   NOT NULL,
    created_at  TIMESTAMPTZ   NOT NULL DEFAULT now()
);
CREATE INDEX ON sessions(token);

-- bills table
CREATE TABLE bills (
    id          BIGSERIAL PRIMARY KEY,
    user_id     BIGINT        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    pdf_url     TEXT,
    amount      NUMERIC(12,2) NOT NULL,
    status      TEXT          NOT NULL CHECK (status IN ('paid','unpaid','overdue')),
    issued_at   TIMESTAMPTZ   NOT NULL DEFAULT now(),
    due_date    DATE
);
CREATE INDEX ON bills(user_id);
