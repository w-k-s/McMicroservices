CREATE SCHEMA IF NOT EXISTS "order";
CREATE TABLE IF NOT EXISTS "order"."order"(
    id VARCHAR(14) PRIMARY KEY,
    toppings VARCHAR(255) NOT NULL,
    status VARCHAR(50) NOT NULL,
    failure_reason VARCHAR(255),
    created_at timestamp with time zone NOT NULL,
    updated_at timestamp with time zone,
    version BIGINT NOT NULL
);
