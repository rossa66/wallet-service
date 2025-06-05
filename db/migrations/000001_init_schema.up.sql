CREATE TABLE wallets (
    id UUID PRIMARY KEY,
    balance BIGINT NOT NULL DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE TABLE transactions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    wallet_id UUID NOT NULL REFERENCES wallets(id),
    operation_type VARCHAR(10) NOT NULL CHECK (operation_type IN ('DEPOSIT', 'WITHDRAW')),
    amount BIGINT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_transactions_wallet_id ON transactions(wallet_id);
