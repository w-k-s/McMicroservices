CREATE SCHEMA IF NOT EXISTS kitchen;

CREATE TABLE IF NOT EXISTS kitchen.stock(
   name VARCHAR (255) NOT NULL,
   units INTEGER NOT NULL,
   CONSTRAINT uq_stock_name UNIQUE(name)
);