package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/SashaBokov/orderbook"
	"github.com/pkg/errors"
)

// Check that Database implements orderbook.OrderBook
var _ = orderbook.OrderBook(&Database{})

// Database is a wrapper around sql.DB with orderbook methods.
type Database struct {
	conn *sql.DB
}

func New(databaseURL string) (*Database, error) {
	conn, err := sql.Open("postgres", databaseURL)
	if err != nil {
		return nil, errors.Wrap(err, "connecting to database")
	}

	if err := conn.Ping(); err != nil {
		return nil, errors.Wrap(err, "pinging database")
	}

	db := &Database{conn: conn}
	if err := db.initOrdersTable(); err != nil {
		return nil, errors.Wrap(err, "initializing orders table")
	}

	return db, nil
}

// initOrdersTable creating orders table
func (db *Database) initOrdersTable() error {
	if _, err := db.conn.Exec(newOrdersTableQuery); err != nil {
		return errors.Wrap(err, "creating orders table")
	}

	return nil
}

// AddNewPair adding new pair to orderbook
func (db *Database) AddNewPair(tokenBid, tokenAsk string) error {
	tx, err := db.conn.BeginTx(context.Background(), nil)
	if err != nil {
		return errors.Wrap(err, "beginning transaction")
	}

	if _, err := tx.Exec(addPairTablesQuery, tokenBid, tokenAsk); err != nil {
		err = errors.Wrap(err, "creating pair tables")
		if errR := tx.Rollback(); err != nil {
			return errors.Wrap(errors.Wrap(errR, "rolling back transaction"), err.Error())
		}

		return err
	}

	return nil
}

// AddOrder adding new order to orderbook
func (db *Database) AddOrder(order orderbook.Order) error {
	tx, err := db.conn.BeginTx(context.Background(), nil)
	if err != nil {
		return errors.Wrap(err, "beginning transaction")
	}

	if _, err := tx.Exec(addOrderQuery, order.Id, order.MakerId, order.TokenBid, order.TokenAsk, order.Rate, order.MaxVolume, order.MaxVolume); err != nil {
		err = errors.Wrap(err, "inserting order")
		if errR := tx.Rollback(); err != nil {
			return errors.Wrap(errors.Wrap(errR, "rolling back transaction"), err.Error())
		}

		return err
	}

	if err := tx.Commit(); err != nil {
		return errors.Wrap(err, "committing transaction")
	}

	return nil
}

// GetOrderById getting order from orderbook
func (db *Database) GetOrderById(orderId string) (orderbook.Order, error) {
	rows, err := db.conn.Query(getOrderFromOrdersTableQuery, orderId)
	if err != nil {
		return orderbook.Order{}, errors.Wrap(err, "getting order by id")
	}

	ordersFromOrdersTable, err := db.parseSQLRowsFromOrdersTable(rows)
	if err != nil {
		return orderbook.Order{}, errors.Wrap(err, "parsing sql rows from orders table")
	}

	if len(ordersFromOrdersTable) == 0 {
		return orderbook.Order{}, errors.Wrap(err, "no order with this id")
	}

	order, err := db.getOrderByPairAndId(orderId, ordersFromOrdersTable[0].TokenBid, ordersFromOrdersTable[0].TokenAsk)
	if err != nil {
		return orderbook.Order{}, errors.Wrap(err, "getting order by pair and id")
	}

	return order, nil
}

// GetOrderWithMaxRate getting order from orderbook with max rate
func (db *Database) GetOrderWithMaxRate(tokenBid, tokenAsk string) (orderbook.Order, error) {
	rows, err := db.conn.Query(getOrderWithMaxRateQuery, tokenBid, tokenAsk)
	if err != nil {
		return orderbook.Order{}, errors.Wrap(err, "getting order with max rate")
	}

	orders, err := db.parseSQLRowsToOrders(rows)
	if err != nil {
		return orderbook.Order{}, errors.Wrap(err, "parsing sql rows to orders")
	}

	if len(orders) == 0 {
		return orderbook.Order{}, fmt.Errorf("no orders with this pair")
	}

	return orders[0], nil
}

// GetOrderWithMinRate getting order from orderbook with min rate
func (db *Database) GetOrderWithMinRate(tokenBid, tokenAsk string) (orderbook.Order, error) {
	rows, err := db.conn.Query(getOrderWithMinRateQuery, tokenBid, tokenAsk)
	if err != nil {
		return orderbook.Order{}, errors.Wrap(err, "getting order with min rate")
	}

	orders, err := db.parseSQLRowsToOrders(rows)
	if err != nil {
		return orderbook.Order{}, errors.Wrap(err, "parsing sql rows to orders")
	}

	if len(orders) == 0 {
		return orderbook.Order{}, fmt.Errorf("no orders with this pair")
	}

	return orders[0], nil
}

// GetOrderWithMaxVolume getting order from orderbook with max volume
func (db *Database) GetOrderWithMaxVolume(tokenBid, tokenAsk string) (orderbook.Order, error) {
	rows, err := db.conn.Query(getOrderWithMaxVolumeQuery, tokenBid, tokenAsk)
	if err != nil {
		return orderbook.Order{}, errors.Wrap(err, "getting order with max volume")
	}

	orders, err := db.parseSQLRowsToOrders(rows)
	if err != nil {
		return orderbook.Order{}, errors.Wrap(err, "parsing sql rows to orders")
	}

	if len(orders) == 0 {
		return orderbook.Order{}, fmt.Errorf("no orders with this pair")
	}

	return orders[0], nil
}

// GetOrderWithMinVolume getting order from orderbook with min volume
func (db *Database) GetOrderWithMinVolume(tokenBid, tokenAsk string) (orderbook.Order, error) {
	rows, err := db.conn.Query(getOrderWithMinVolumeQuery, tokenBid, tokenAsk)
	if err != nil {
		return orderbook.Order{}, errors.Wrap(err, "getting order with min volume")
	}

	orders, err := db.parseSQLRowsToOrders(rows)
	if err != nil {
		return orderbook.Order{}, errors.Wrap(err, "parsing sql rows to orders")
	}

	if len(orders) == 0 {
		return orderbook.Order{}, fmt.Errorf("no orders with this pair")
	}

	return orders[0], nil
}

// ListOrdersByPair getting orders from orderbook by pair
func (db *Database) ListOrdersByPair(tokenBid, tokenAsk string, limit, offset int) ([]orderbook.Order, error) {
	rows, err := db.conn.Query(listOrdersByPairQuery, tokenBid, tokenAsk, db.convertLimitOffset(limit, offset))
	if err != nil {
		return nil, errors.Wrap(err, "getting orders by pair")
	}

	orders, err := db.parseSQLRowsToOrders(rows)
	if err != nil {
		return nil, errors.Wrap(err, "parsing sql rows to orders")
	}

	if len(orders) == 0 {
		return nil, fmt.Errorf("no orders with this pair")
	}

	return orders, nil
}

// ListOrdersByMakerId getting order from orderbook
func (db *Database) ListOrdersByMakerId(makerId string, limit, offset int) ([]orderbook.Order, error) {
	rows, err := db.conn.Query(listOrdersByMakerIdQuery, makerId, db.convertLimitOffset(limit, offset))
	if err != nil {
		return nil, errors.Wrap(err, "getting order by maker id")
	}

	orders, err := db.parseSQLRowsToOrders(rows)
	if err != nil {
		return nil, errors.Wrap(err, "parsing sql rows to orders")
	}

	if len(orders) == 0 {
		return nil, fmt.Errorf("no orders with this maker id")
	}

	return orders, nil
}

// ListMaxRateOrders getting orders from orderbook with max rate
func (db *Database) ListMaxRateOrders(tokenBid, tokenAsk string, limit, offset int) ([]orderbook.Order, error) {
	rows, err := db.conn.Query(listMaxRateOrdersQuery, tokenBid, tokenAsk, db.convertLimitOffset(limit, offset))
	if err != nil {
		return nil, errors.Wrap(err, "getting orders with max rate")
	}

	orders, err := db.parseSQLRowsToOrders(rows)
	if err != nil {
		return nil, errors.Wrap(err, "parsing sql rows to orders")
	}

	if len(orders) == 0 {
		return nil, fmt.Errorf("no orders with this pair")
	}

	return orders, nil
}

// ListMinRateOrders getting orders from orderbook with min rate
func (db *Database) ListMinRateOrders(tokenBid, tokenAsk string, limit, offset int) ([]orderbook.Order, error) {
	rows, err := db.conn.Query(listMinRateOrdersQuery, tokenBid, tokenAsk, db.convertLimitOffset(limit, offset))
	if err != nil {
		return nil, errors.Wrap(err, "getting orders with min rate")
	}

	orders, err := db.parseSQLRowsToOrders(rows)
	if err != nil {
		return nil, errors.Wrap(err, "parsing sql rows to orders")
	}

	if len(orders) == 0 {
		return nil, fmt.Errorf("no orders with this pair")
	}

	return orders, nil
}

// ListMaxVolumeOrders getting orders from orderbook with max volume
func (db *Database) ListMaxVolumeOrders(tokenBid, tokenAsk string, limit, offset int) ([]orderbook.Order, error) {
	rows, err := db.conn.Query(listMaxVolumeOrdersQuery, tokenBid, tokenAsk, db.convertLimitOffset(limit, offset))
	if err != nil {
		return nil, errors.Wrap(err, "getting orders with max volume")
	}

	orders, err := db.parseSQLRowsToOrders(rows)
	if err != nil {
		return nil, errors.Wrap(err, "parsing sql rows to orders")
	}

	if len(orders) == 0 {
		return nil, fmt.Errorf("no orders with this pair")
	}

	return orders, nil
}

// ListMinVolumeOrders getting orders from orderbook with min volume
func (db *Database) ListMinVolumeOrders(tokenBid, tokenAsk string, limit, offset int) ([]orderbook.Order, error) {
	rows, err := db.conn.Query(listMinVolumeOrdersQuery, tokenBid, tokenAsk, db.convertLimitOffset(limit, offset))
	if err != nil {
		return nil, errors.Wrap(err, "getting orders with min volume")
	}

	orders, err := db.parseSQLRowsToOrders(rows)
	if err != nil {
		return nil, errors.Wrap(err, "parsing sql rows to orders")
	}

	if len(orders) == 0 {
		return nil, fmt.Errorf("no orders with this pair")
	}

	return orders, nil
}

// RemovePair removing pair from orderbook
func (db *Database) RemovePair(tokenBid, tokenAsk string) error {
	tx, err := db.conn.BeginTx(context.Background(), nil)
	if err != nil {
		return errors.Wrap(err, "beginning transaction")
	}

	_, err = tx.Exec(removePairQuery, tokenBid, tokenAsk)
	if err != nil {
		err = errors.Wrap(err, "exec remove pair query")
		if errR := tx.Rollback(); err != nil {
			return errors.Wrap(errors.Wrap(errR, "rolling back transaction"), err.Error())
		}
		return err
	}

	err = tx.Commit()
	if err != nil {
		return errors.Wrap(err, "committing transaction")
	}

	return nil
}

// RemoveOrder removing order from orderbook
func (db *Database) RemoveOrder(orderId string) error {
	_, err := db.conn.Exec(removeOrderQuery, orderId)
	if err != nil {
		return errors.Wrap(err, "exec remove order query")
	}

	return nil
}

// getOrderByPairAndId getting order from orderbook by pair and id
func (db *Database) getOrderByPairAndId(orderId, tokenBid, tokenAsk string) (orderbook.Order, error) {
	rows, err := db.conn.Query(getOrderByIdAndPairQuery, orderId, tokenBid, tokenAsk)
	if err != nil {
		return orderbook.Order{}, errors.Wrap(err, "getting order by pair and id")
	}

	orderIdWithPair, err := db.parseSQLRowsFromOrdersTable(rows)
	if err != nil {
		return orderbook.Order{}, errors.Wrap(err, "parsing sql rows from orders table")
	}

	if len(orderIdWithPair) == 0 {
		return orderbook.Order{}, errors.Wrap(err, "no order with this id")
	}

	return orderbook.Order{}, nil
}

// parseSQLRowsToOrders parsing sql.Rows to []orderbook.Order
func (db *Database) parseSQLRowsToOrders(rows *sql.Rows) ([]orderbook.Order, error) {
	orders := make([]orderbook.Order, 0)
	for rows.Next() {
		var order orderbook.Order
		if err := rows.Scan(&order.Id, &order.MakerId, &order.TokenBid, &order.TokenAsk, &order.Rate, &order.MaxVolume, &order.MinVolume); err != nil {
			return nil, errors.Wrap(err, "scanning rows")
		}
		orders = append(orders, order)
	}

	return orders, nil
}

// parseSQLRowsFromOrdersTable parsing sql.Rows form orders table to []orderbook.Order
func (db *Database) parseSQLRowsFromOrdersTable(rows *sql.Rows) ([]orderbook.Order, error) {
	orders := make([]orderbook.Order, 0)
	for rows.Next() {
		var order orderbook.Order
		if err := rows.Scan(&order.Id, &order.MakerId, &order.TokenBid, &order.TokenAsk); err != nil {
			return nil, errors.Wrap(err, "scanning rows")
		}
		orders = append(orders, order)
	}

	return orders, nil
}

// convertLimitOffset converting limit and offset to part of query
func (db *Database) convertLimitOffset(limit, offset int) string {
	result := ""
	if limit != -1 {
		result += fmt.Sprintf(" LIMIT %d", limit)
	}
	if offset != -1 {
		result += fmt.Sprintf(" OFFSET %d", offset)
	}

	return result
}
