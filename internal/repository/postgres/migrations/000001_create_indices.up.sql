CREATE INDEX IF NOT EXISTS idx_crypto_prices_symbol_timestamp
    ON crypto_prices (symbol, timestamp DESC);