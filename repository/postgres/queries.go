package postgres

var newOrdersTableQuery = `
CREATE TABLE IF NOT EXISTS orderbook_orders (
    id BYTEA PRIMARY KEY NOT NULL,
    maker_id BYTEA NOT NULL,
    token_bid VARCHAR(255) NOT NULL,
    token_ask VARCHAR(255) NOT NULL
);

CREATE INDEX orderbook_orders_maker_id ON orderbook_orders USING hash (maker_id);
`

var addPairTablesQuery = `
CREATE TABLE IF NOT EXISTS $1_$2_min_volume (
    id BYTEA PRIMARY KEY NOT NULL,
    min_volume DECIMAL,
    FOREIGN KEY (id) REFERENCES orderbook_orders (id) ON DELETE CASCADE
);
CREATE INDEX orderbook_orders_tree_$1_$2_min_volume ON $1_$2_min_volume using btree (min_volume);

CREATE TABLE IF NOT EXISTS $1_$2_max_volume (
    id BYTEA PRIMARY KEY NOT NULL,
    max_volume DECIMAL NOT NULL,
    FOREIGN KEY (id) REFERENCES orderbook_orders (id) ON DELETE CASCADE
    );
CREATE INDEX orderbook_orders_tree_$1_$2_max_volume ON $1_$2_max_volume using btree (max_volume);

CREATE TABLE IF NOT EXISTS $1_$2_rate (
    id BYTEA PRIMARY KEY NOT NULL,
    rate DECIMAL NOT NULL,
    FOREIGN KEY (id) REFERENCES orderbook_orders (id) ON DELETE CASCADE
);
CREATE INDEX orderbook_orders_tree_$1_$2_rate ON $1_$2_rate using btree (rate);


CREATE TABLE IF NOT EXISTS $2_$1_min_volume (
    id BYTEA PRIMARY KEY NOT NULL,
    min_volume DECIMAL,
    FOREIGN KEY (id) REFERENCES orderbook_orders (id) ON DELETE CASCADE
);
CREATE INDEX orderbook_orders_tree_asks_$2_$1_min_volume ON $2_$1_min_volume using btree (min_volume);

CREATE TABLE IF NOT EXISTS $2_$1_max_volume (
    id BYTEA PRIMARY KEY NOT NULL,
    max_volume DECIMAL NOT NULL,
    FOREIGN KEY (id) REFERENCES orderbook_orders (id) ON DELETE CASCADE
);
CREATE INDEX orderbook_orders_tree_$2_$1_max_volume ON $2_$1_max_volume using btree (max_volume);

CREATE TABLE IF NOT EXISTS $2_$1_rate (
    id BYTEA PRIMARY KEY NOT NULL,
    rate DECIMAL NOT NULL,
    FOREIGN KEY (id) REFERENCES orderbook_orders (id) ON DELETE CASCADE
);
CREATE INDEX orderbook_orders_tree_$2_$1_rate ON $2_$1_rate using btree (rate);
`

var addOrderQuery = `
	INSERT INTO orderbook_orders VALUES ($1, $2, $3, $4);
	INSERT INTO $3_$4_rate VALUES ($1, $5);
	INSERT INTO $3_$4_max_volume VALUES ($1, $6);
	INSERT INTO $3_$4_min_volume VALUES ($1, $7);
`

var getOrderFromOrdersTableQuery = `
SELECT orderbook_orders.id,
    orderbook_orders.maker_id,
    orderbook_orders.token_bid,
    orderbook_orders.token_ask
FROM orderbook_orders
WHERE orderbook_orders.id = $1;
`

var getOrderByIdAndPairQuery = `
SELECT order.id,
    orderbook_orders.maker_id,
    orderbook_orders.token_bid,
    orderbook_orders.token_ask,
    $2_$3_rate.rate,
    $2_$3_max_volume.max_volume,
    $2_$3_min_volume.min_volume
FROM orderbook_orders
    JOIN $2_$3_max_volume ON $2_$3_max_volume.id = orderbook_orders.id
    JOIN $2_$3_min_volume ON $2_$3_min_volume.id = orderbook_orders.id
    JOIN $2_$3_rate ON $2_$3_rate.id = orderbook_orders.id
WHERE orderbook_orders.id = $1;
`

var getOrderWithMaxRateQuery = `
SELECT orderbook_orders.id,
    orderbook_orders.maker_id,
    orderbook_orders.token_bid,
    orderbook_orders.token_ask,
    $1_$2_rate.rate
    $1_$2_max_volume.max_volume,
    $1_$2_min_volume.min_volume
FROM $1_$2_rate
    JOIN $1_$2_min_volume ON $1_$2_min_volume.id = $1_$2_rate.id
    JOIN $1_$2_max_volume ON $1_$2_max_volume.id = $1_$2_rate.id
    JOIN orderbook_orders ON orderbook_orders.id = $1_$2_rate.id
WHERE $1_$2_rate.rate = (SELECT MAX($1_$2_rate.rate) FROM $1_$2_rate) LIMIT 1;
`

var getOrderWithMinRateQuery = `
SELECT orderbook_orders.id,
    orderbook_orders.maker_id,
    orderbook_orders.token_bid,
    orderbook_orders.token_ask,
    $1_$2_rate.rate
    $1_$2_max_volume.max_volume,
    $1_$2_min_volume.min_volume
FROM $1_$2_rate
    JOIN $1_$2_min_volume ON $1_$2_min_volume.id = $1_$2_rate.id
    JOIN $1_$2_max_volume ON $1_$2_max_volume.id = $1_$2_rate.id
    JOIN orderbook_orders ON orderbook_orders.id = $1_$2_rate.id
WHERE $1_$2_rate.rate = (SELECT MIN($1_$2_rate.rate) FROM $1_$2_rate) LIMIT 1;
`

var getOrderWithMaxVolumeQuery = `
SELECT orderbook_orders.id,
    orderbook_orders.maker_id,
    orderbook_orders.token_bid,
    orderbook_orders.token_ask,
    $1_$2_rate.rate,
    $1_$2_max_volume.max_volume,
    $1_$2_min_volume.min_volume
FROM $1_$2_max_volume
    JOIN $1_$2_rate ON $1_$2_rate.id = $1_$2_min_volume.id
    JOIN $1_$2_max_volume ON $1_$2_max_volume.id = $1_$2_min_volume.id
    JOIN orderbook_orders ON orderbook_orders.id = $1_$2_min_volume.id
WHERE $1_$2_max_volume.max_volume = (SELECT MAX($1_$2_max_volume.max_volume) FROM $1_$2_max_volume) LIMIT 1;
`

var getOrderWithMinVolumeQuery = `
SELECT orderbook_orders.id,
    orderbook_orders.maker_id,
    orderbook_orders.token_bid,
    orderbook_orders.token_ask,
    $1_$2_rate.rate,
    $1_$2_max_volume.max_volume,
    $1_$2_min_volume.min_volume
FROM $1_$2_min_volume
    JOIN $1_$2_rate ON $1_$2_rate.id = $1_$2_min_volume.id
    JOIN $1_$2_max_volume ON $1_$2_max_volume.id = $1_$2_min_volume.id
    JOIN orderbook_orders ON orderbook_orders.id = $1_$2_min_volume.id
WHERE $1_$2_min_volume.min_volume = (SELECT MIN($1_$2_min_volume.min_volume) FROM $1_$2_min_volume) LIMIT 1;
`

var listOrdersByPairQuery = `
SELECT order.id,
    orderbook_orders.maker_id,
    orderbook_orders.token_bid,
    orderbook_orders.token_ask,
    $1_$2_rate.rate,
    $1_$2_max_volume.max_volume,
    $1_$2_min_volume.min_volume
FROM orderbook_orders
    JOIN $1_$2_max_volume ON $1_$2_max_volume.id = orderbook_orders.id
    JOIN $1_$2_min_volume ON $1_$2_min_volume.id = orderbook_orders.id
    JOIN $1_$2_rate ON $1_$2_rate.id = orderbook_orders.id
WHERE orderbook_orders.token_bid = $1 AND orderbook_orders.token_ask = $2 ORDER BY order.id $3;
`

var listOrdersByMakerIdFromOrdersTableQuery = `
SELECT orderbook_orders.id,
    orderbook_orders.maker_id,
    orderbook_orders.token_bid,
    orderbook_orders.token_ask
FROM orderbook_orders
WHERE orderbook_orders.maker_id = $1$2;
`

var listOrdersByMakerIdQuery = `
SELECT order.id,
    orderbook_orders.maker_id,
    orderbook_orders.token_bid,
    orderbook_orders.token_ask,
    $2_$3_rate.rate,
    $2_$3_max_volume.max_volume,
    $2_$3_min_volume.min_volume
FROM orderbook_orders
    JOIN $2_$3_max_volume ON $2_$3_max_volume.id = orderbook_orders.id
    JOIN $2_$3_min_volume ON $2_$3_min_volume.id = orderbook_orders.id
    JOIN $2_$3_rate ON $2_$3_rate.id = orderbook_orders.id
WHERE maker_id = $1$4;
`

var listMaxRateOrdersQuery = `
SELECT orderbook_orders.id,
    orderbook_orders.maker_id,
    orderbook_orders.token_bid,
    orderbook_orders.token_ask,
    $1_$2_rate.rate
    $1_$2_max_volume.max_volume,
    $1_$2_min_volume.min_volume
FROM $1_$2_rate
    JOIN $1_$2_min_volume ON $1_$2_min_volume.id = $1_$2_rate.id
    JOIN $1_$2_max_volume ON $1_$2_max_volume.id = $1_$2_rate.id
    JOIN orderbook_orders ON orderbook_orders.id = $1_$2_rate.id
WHERE $1_$2_rate.rate = (SELECT $1_$2_rate.rate FROM $1_$2_rate ORDER BY $1_$2_rate.rate DESC$3);
`

var listMinRateOrdersQuery = `
SELECT orderbook_orders.id,
    orderbook_orders.maker_id,
    orderbook_orders.token_bid,
    orderbook_orders.token_ask,
    $1_$2_rate.rate
    $1_$2_max_volume.max_volume,
    $1_$2_min_volume.min_volume
FROM $1_$2_rate
    JOIN $1_$2_min_volume ON $1_$2_min_volume.id = $1_$2_rate.id
    JOIN $1_$2_max_volume ON $1_$2_max_volume.id = $1_$2_rate.id
    JOIN orderbook_orders ON orderbook_orders.id = $1_$2_rate.id
WHERE $1_$2_rate.rate = (SELECT $1_$2_rate.rate FROM $1_$2_rate ORDER BY $1_$2_rate.rate ASK$3);
`

var listMaxVolumeOrdersQuery = `
SELECT orderbook_orders.id,
    orderbook_orders.maker_id,
    orderbook_orders.token_bid,
    orderbook_orders.token_ask,
    $1_$2_rate.rate,
    $1_$2_max_volume.max_volume,
    $1_$2_min_volume.min_volume
FROM $1_$2_max_volume
    JOIN $1_$2_rate ON $1_$2_rate.id = $1_$2_min_volume.id
    JOIN $1_$2_max_volume ON $1_$2_max_volume.id = $1_$2_min_volume.id
    JOIN orderbook_orders ON orderbook_orders.id = $1_$2_min_volume.id
WHERE $1_$2_max_volume.max_volume = (SELECT $1_$2_max_volume.max_volume FROM $1_$2_max_volume ORDER BY $1_$2_max_volume.max_volume DESK$3);
`

var listMinVolumeOrdersQuery = `
SELECT orderbook_orders.id,
    orderbook_orders.maker_id,
    orderbook_orders.token_bid,
    orderbook_orders.token_ask,
    $1_$2_rate.rate,
    $1_$2_max_volume.max_volume,
    $1_$2_min_volume.min_volume
FROM $1_$2_min_volume
    JOIN $1_$2_rate ON $1_$2_rate.id = $1_$2_min_volume.id
    JOIN $1_$2_max_volume ON $1_$2_max_volume.id = $1_$2_min_volume.id
    JOIN orderbook_orders ON orderbook_orders.id = $1_$2_min_volume.id
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
DELETE FROM orderbook_orders WHERE id = $1;
`
