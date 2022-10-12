package call

// Query sets up how a SQL query will be called
type Query struct {
	// Statement that will be queried.
	// eg: SELECT * FROM products
	Statement string
	// Params that can be passed to the SQL query
	Params []any
}
