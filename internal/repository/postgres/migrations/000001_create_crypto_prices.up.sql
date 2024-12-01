CREATE TABLE crypto_prices (
                               id VARCHAR(100) NOT NULL,
                               symbol VARCHAR(20) NOT NULL,
                               name VARCHAR(100) NOT NULL,
                               timestamp TIMESTAMP NOT NULL,
                               current_price NUMERIC(30,8) NOT NULL,
                               high_24h NUMERIC(30,8) NOT NULL,
                               low_24h NUMERIC(30,8) NOT NULL,
                               total_volume NUMERIC(40,8) NOT NULL,
                               market_cap NUMERIC(40,8) NOT NULL,
                               market_rank INTEGER NOT NULL,
                               price_change_24h NUMERIC(30,8) NOT NULL,
                               price_change_percentage_24h NUMERIC(30,8) NOT NULL,
                               circulating_supply NUMERIC(40,8) NOT NULL,
                               total_supply NUMERIC(40,8),
                               PRIMARY KEY (id, timestamp)
);

CREATE INDEX idx_crypto_prices_timestamp ON crypto_prices(timestamp);
CREATE INDEX idx_crypto_prices_id ON crypto_prices(id);