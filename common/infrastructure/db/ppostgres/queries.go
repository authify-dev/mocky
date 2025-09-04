package ppostgres

import "database/sql"

type IQueries[T any] interface {
	WithTx(tx *sql.Tx) T
}
