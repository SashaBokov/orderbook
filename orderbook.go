package orderbook

import (
	"orderbook/repository/postgres"
)

// Order is representation of P2P order
type Order struct {
	Id        string  `json:"id" db:"id"`
	MakerId   string  `json:"maker_id" db:"maker_id"`
	TokenBid  string  `json:"token_bid" db:"token_bid"`
	TokenAsk  string  `json:"token_ask" db:"token_ask"`
	Rate      float64 `json:"rate" json:"rate"`
	MaxVolume float64 `json:"max_volume" db:"max_volume"`
	MinVolume float64 `json:"min_volume" db:"min_volume"`
}

type OrderBook interface {
	// AddNewPair adding new pair to orderbook
	AddNewPair(tokenBid, tokenAsk string) error
	// AddOrder adding new order to orderbook
	AddOrder(order Order) error

	// GetOrderById getting order from orderbook
	GetOrderById(orderId string) (Order, error)
	// GetOrderWithMaxRate getting order from orderbook with max rate
	GetOrderWithMaxRate(tokenBid, tokenAsk string) (Order, error)
	// GetOrderWithMinRate getting order from orderbook with min rate
	GetOrderWithMinRate(tokenBid, tokenAsk string) (Order, error)
	// GetOrderWithMaxVolume getting order from orderbook with max volume
	GetOrderWithMaxVolume(tokenBid, tokenAsk string) (Order, error)
	// GetOrderWithMinVolume getting order from orderbook with min volume
	GetOrderWithMinVolume(tokenBid, tokenAsk string) (Order, error)

	// For using lists without limit and/or offset, set limit -1 and/or offset -1

	// ListOrdersByPair getting orders from orderbook by pair
	ListOrdersByPair(tokenBid, tokenAsk string, limit, offset int) ([]Order, error)
	// ListOrdersByMakerId getting order from orderbook
	ListOrdersByMakerId(makerId string, limit, offset int) (Order, error)
	//	ListMaxRateOrders getting orders from orderbook with max rate
	ListMaxRateOrders(tokenBid, tokenAsk string, limit, offset int) ([]Order, error)
	//	ListMinRateOrders getting orders from orderbook with min rate
	ListMinRateOrders(tokenBid, tokenAsk string, limit, offset int) ([]Order, error)
	//	ListMaxVolumeOrders getting orders from orderbook with max volume
	ListMaxVolumeOrders(tokenBid, tokenAsk string, limit, offset int) ([]Order, error)
	//	ListMinVolumeOrders getting orders from orderbook with min volume
	ListMinVolumeOrders(tokenBid, tokenAsk string, limit, offset int) ([]Order, error)

	// RemovePair removing pair from orderbook
	RemovePair(tokenBid, tokenAsk string) error
	// RemoveOrder removing order from orderbook
	RemoveOrder(orderId string) error
}

// NewOrderBookPostgres OrderBookPostgres constructor returns OrderBook postgres implementation
func NewOrderBookPostgres(databaseURL string) (OrderBook, error) {
	return postgres.New(databaseURL)
}
