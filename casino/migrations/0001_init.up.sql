-- Users

CREATE TABLE IF NOT EXISTS service.users (
  user_id  uuid        PRIMARY KEY,
  username varchar(30) NOT NULL,
  UNIQUE(username)
);

CREATE TABLE IF NOT EXISTS service.user_sessions (
  session_id uuid      PRIMARY KEY,
  user_id    uuid      REFERENCES service.users ON DELETE CASCADE,
  expires_at timestamp NOT NULL
);

CREATE INDEX session_expiry ON service.user_sessions (expires_at);

-- Wallets

CREATE TABLE IF NOT EXISTS service.wallets (
  wallet_id     uuid    PRIMARY KEY,
  balance_minor bigint  DEFAULT 0,
  soft_deleted  boolean DEFAULT false
);

CREATE TABLE IF NOT EXISTS service.user_wallets (
  user_id   uuid REFERENCES service.users   ON DELETE RESTRICT,
  wallet_id uuid REFERENCES service.wallets ON DELETE RESTRICT
);

CREATE TYPE transaction_type AS ENUM ('DEBIT', 'CREDIT');

CREATE TABLE IF NOT EXISTS service.transactions (
  transaction_id   uuid             PRIMARY KEY,
  wallet_id        uuid             REFERENCES service.wallets ON DELETE CASCADE,
  transaction_type transaction_type NOT NULL,
  amount_minor     bigint           NOT NULL,
  time_stamp       timestamp        NOT NULL
);

-- Games

CREATE TYPE game_status_type as ENUM ('playing', 'closed', 'refunded');

CREATE TABLE IF NOT EXISTS service.games (
  game_id           uuid             PRIMARY KEY,
  user_id           uuid             REFERENCES service.users ON DELETE CASCADE,
  engine_id         varchar(50)      NOT NULL,
  bet               bigint           NOT NULL,
  game_status       game_status_type NOT NULL,
  payout_multiplier numeric(9, 2)    NOT NULL,
  private_state     bytea            NOT NULL,
  public_state      bytea            NOT NULL
);

CREATE INDEX games_index ON service.games (user_id, engine_id, game_status);