package postgres

var newOrdersTableQuery = `
CREATE TABLE IF NOT EXISTS orders (
    id BYTEA PRIMARY KEY NOT NULL,
    maker_id BYTEA NOT NULL,
    token_bid VARCHAR(255) NOT NULL,
    token_ask VARCHAR(255) NOT NULL
);

CREATE INDEX orders_maker_id ON orders USING hash (maker_id);
`

var addPairTablesQuery = `
CREATE TABLE IF NOT EXISTS $1_$2_min_volume (
    id BYTEA PRIMARY KEY NOT NULL,
    min_volume DECIMAL,
    FOREIGN KEY (id) REFERENCES orders (id) ON DELETE CASCADE
);
CREATE INDEX orders_tree_$1_$2_min_volume ON $1_$2_min_volume using btree (min_volume);

CREATE TABLE IF NOT EXISTS $1_$2_max_volume (
    id BYTEA PRIMARY KEY NOT NULL,
    max_volume DECIMAL NOT NULL,
    FOREIGN KEY (id) REFERENCES orders (id) ON DELETE CASCADE
    );
CREATE INDEX orders_tree_$1_$2_max_volume ON $1_$2_max_volume using btree (max_volume);

CREATE TABLE IF NOT EXISTS $1_$2_rate (
    id BYTEA PRIMARY KEY NOT NULL,
    rate DECIMAL NOT NULL,
    FOREIGN KEY (id) REFERENCES orders (id) ON DELETE CASCADE
);
CREATE INDEX orders_tree_$1_$2_rate ON $1_$2_rate using btree (rate);


CREATE TABLE IF NOT EXISTS $2_$1_min_volume (
    id BYTEA PRIMARY KEY NOT NULL,
    min_volume DECIMAL,
    FOREIGN KEY (id) REFERENCES orders (id) ON DELETE CASCADE
);
CREATE INDEX orders_tree_asks_$2_$1_min_volume ON $2_$1_min_volume using btree (min_volume);

CREATE TABLE IF NOT EXISTS $2_$1_max_volume (
    id BYTEA PRIMARY KEY NOT NULL,
    max_volume DECIMAL NOT NULL,
    FOREIGN KEY (id) REFERENCES orders (id) ON DELETE CASCADE
);
CREATE INDEX orders_tree_$2_$1_max_volume ON $2_$1_max_volume using btree (max_volume);

CREATE TABLE IF NOT EXISTS $2_$1_rate (
    id BYTEA PRIMARY KEY NOT NULL,
    rate DECIMAL NOT NULL,
    FOREIGN KEY (id) REFERENCES orders (id) ON DELETE CASCADE
);
CREATE INDEX orders_tree_$2_$1_rate ON $2_$1_rate using btree (rate);
`

var addOrderQuery = `
	INSERT INTO orders VALUES ($1, $2, $3, $4);
	INSERT INTO $3_$4_rate VALUES ($1, $5);
	INSERT INTO $3_$4_max_volume VALUES ($1, $6);
	INSERT INTO $3_$4_min_volume VALUES ($1, $7);
`

var getOrderFromOrdersTableQuery = `
SELECT orders.id,
    orders.maker_id,
    orders.token_bid,
    orders.token_ask
FROM orders
WHERE orders.id = $1;
`

var getOrderByIdAndPairQuery = `
SELECT order.id,
    orders.maker_id,
    orders.token_bid,
    orders.token_ask,
    $2_$3_rate.rate,
    $2_$3_max_volume.max_volume,
    $2_$3_min_volume.min_volume
FROM orders
    JOIN $2_$3_max_volume ON $2_$3_max_volume.id = orders.id
    JOIN $2_$3_min_volume ON $2_$3_min_volume.id = orders.id
    JOIN $2_$3_rate ON $2_$3_rate.id = orders.id
WHERE orders.id = $1;
`

var getOrderWithMaxRateQuery = `
SELECT orders.id,
    orders.maker_id,
    orders.token_bid,
    orders.token_ask,
    $1_$2_rate.rate
    $1_$2_max_volume.max_volume,
    $1_$2_min_volume.min_volume
FROM $1_$2_rate
    JOIN $1_$2_min_volume ON $1_$2_min_volume.id = $1_$2_rate.id
    JOIN $1_$2_max_volume ON $1_$2_max_volume.id = $1_$2_rate.id
    JOIN orders ON orders.id = $1_$2_rate.id
WHERE $1_$2_rate.rate = (SELECT MAX($1_$2_rate.rate) FROM $1_$2_rate) LIMIT 1;
`

var getOrderWithMinRateQuery = `
SELECT orders.id,
    orders.maker_id,
    orders.token_bid,
    orders.token_ask,
    $1_$2_rate.rate
    $1_$2_max_volume.max_volume,
    $1_$2_min_volume.min_volume
FROM $1_$2_rate
    JOIN $1_$2_min_volume ON $1_$2_min_volume.id = $1_$2_rate.id
    JOIN $1_$2_max_volume ON $1_$2_max_volume.id = $1_$2_rate.id
    JOIN orders ON orders.id = $1_$2_rate.id
WHERE $1_$2_rate.rate = (SELECT MIN($1_$2_rate.rate) FROM $1_$2_rate) LIMIT 1;
`

var getOrderWithMaxVolumeQuery = `
SELECT orders.id,
    orders.maker_id,
    orders.token_bid,
    orders.token_ask,
    $1_$2_rate.rate,
    $1_$2_max_volume.max_volume,
    $1_$2_min_volume.min_volume
FROM $1_$2_max_volume
    JOIN $1_$2_rate ON $1_$2_rate.id = $1_$2_min_volume.id
    JOIN $1_$2_max_volume ON $1_$2_max_volume.id = $1_$2_min_volume.id
    JOIN orders ON orders.id = $1_$2_min_volume.id
WHERE $1_$2_max_volume.max_volume = (SELECT MAX($1_$2_max_volume.max_volume) FROM $1_$2_max_volume) LIMIT 1;
`

var getOrderWithMinVolumeQuery = `
SELECT orders.id,
    orders.maker_id,
    orders.token_bid,
    orders.token_ask,
    $1_$2_rate.rate,
    $1_$2_max_volume.max_volume,
    $1_$2_min_volume.min_volume
FROM $1_$2_min_volume
    JOIN $1_$2_rate ON $1_$2_rate.id = $1_$2_min_volume.id
    JOIN $1_$2_max_volume ON $1_$2_max_volume.id = $1_$2_min_volume.id
    JOIN orders ON orders.id = $1_$2_min_volume.id
WHERE $1_$2_min_volume.min_volume = (SELECT MIN($1_$2_min_volume.min_volume) FROM $1_$2_min_volume) LIMIT 1;
`

var listOrdersByPairQuery = `
SELECT order.id,
    orders.maker_id,
    orders.token_bid,
    orders.token_ask,
    $1_$2_rate.rate,
    $1_$2_max_volume.max_volume,
    $1_$2_min_volume.min_volume
FROM orders
    JOIN $1_$2_max_volume ON $1_$2_max_volume.id = orders.id
    JOIN $1_$2_min_volume ON $1_$2_min_volume.id = orders.id
    JOIN $1_$2_rate ON $1_$2_rate.id = orders.id
WHERE orders.token_bid = $1 AND orders.token_ask = $2 ORDER BY order.id $3;
`

var listOrdersByMakerIdFromOrdersTableQuery = `
SELECT orders.id,
    orders.maker_id,
    orders.token_bid,
    orders.token_ask
FROM orders
WHERE orders.maker_id = $1$2;
`

var listOrdersByMakerIdQuery = `
SELECT order.id,
    orders.maker_id,
    orders.token_bid,
    orders.token_ask,
    $2_$3_rate.rate,
    $2_$3_max_volume.max_volume,
    $2_$3_min_volume.min_volume
FROM orders
    JOIN $2_$3_max_volume ON $2_$3_max_volume.id = orders.id
    JOIN $2_$3_min_volume ON $2_$3_min_volume.id = orders.id
    JOIN $2_$3_rate ON $2_$3_rate.id = orders.id
WHERE maker_id = $1$4;
`

var listMaxRateOrdersQuery = `
SELECT orders.id,
    orders.maker_id,
    orders.token_bid,
    orders.token_ask,
    $1_$2_rate.rate
    $1_$2_max_volume.max_volume,
    $1_$2_min_volume.min_volume
FROM $1_$2_rate
    JOIN $1_$2_min_volume ON $1_$2_min_volume.id = $1_$2_rate.id
    JOIN $1_$2_max_volume ON $1_$2_max_volume.id = $1_$2_rate.id
    JOIN orders ON orders.id = $1_$2_rate.id
WHERE $1_$2_rate.rate = (SELECT $1_$2_rate.rate FROM $1_$2_rate ORDER BY $1_$2_rate.rate DESC$3);
`

var listMinRateOrdersQuery = `
SELECT orders.id,
    orders.maker_id,
    orders.token_bid,
    orders.token_ask,
    $1_$2_rate.rate
    $1_$2_max_volume.max_volume,
    $1_$2_min_volume.min_volume
FROM $1_$2_rate
    JOIN $1_$2_min_volume ON $1_$2_min_volume.id = $1_$2_rate.id
    JOIN $1_$2_max_volume ON $1_$2_max_volume.id = $1_$2_rate.id
    JOIN orders ON orders.id = $1_$2_rate.id
WHERE $1_$2_rate.rate = (SELECT $1_$2_rate.rate FROM $1_$2_rate ORDER BY $1_$2_rate.rate ASK$3);
`

var listMaxVolumeOrdersQuery = `
SELECT orders.id,
    orders.maker_id,
    orders.token_bid,
    orders.token_ask,
    $1_$2_rate.rate,
    $1_$2_max_volume.max_volume,
    $1_$2_min_volume.min_volume
FROM $1_$2_max_volume
    JOIN $1_$2_rate ON $1_$2_rate.id = $1_$2_min_volume.id
    JOIN $1_$2_max_volume ON $1_$2_max_volume.id = $1_$2_min_volume.id
    JOIN orders ON orders.id = $1_$2_min_volume.id
WHERE $1_$2_max_volume.max_volume = (SELECT $1_$2_max_volume.max_volume FROM $1_$2_max_volume ORDER BY $1_$2_max_volume.max_volume DESK$3);
`

var listMinVolumeOrdersQuery = `
SELECT orders.id,
    orders.maker_id,
    orders.token_bid,
    orders.token_ask,
    $1_$2_rate.rate,
    $1_$2_max_volume.max_volume,
    $1_$2_min_volume.min_volume
FROM $1_$2_min_volume
    JOIN $1_$2_rate ON $1_$2_rate.id = $1_$2_min_volume.id
    JOIN $1_$2_max_volume ON $1_$2_max_volume.id = $1_$2_min_volume.id
    JOIN orders ON orders.id = $1_$2_min_volume.id
WHERE $1_$2_min_volume.min_volume = (SELECT $1_$2_min_volume.min_volume FROM $1_$2_min_volume ORDER BY $1_$2_min_volume.min_volume ASK$3);
`

var removePairQuery = `
DROP TABLE IF EXISTS $1_$2_rate;
DROP TABLE IF EXISTS $1_$2_max_volume;
DROP TABLE IF EXISTS $1_$2_min_volume;

DROP TABLE IF EXISTS $2_$1_rate;
DROP TABLE IF EXISTS $2_$1_max_volume;
DROP TABLE IF EXISTS $2_$1_min_volume;
`

var removeOrderQuery = `
DELETE FROM orders WHERE id = $1;
`
