package ppostgres

import (
	"common/domain"
	"common/utils"
	"common/utils/cerrs"
	"database/sql"
	"fmt"
	"net/http"
)

// --------------------------------
// INFRASTRUCTURE
// --------------------------------
// PostgresRepository
// --------------------------------

type PostgresRepository[E domain.IEntity, M domain.IModel, Q any] struct {
	Queries Q
	Conn    *sql.DB
}

func (m *PostgresRepository[E, M, Q]) View(data []E) {

	for _, e := range data {

		fmt.Println(string(domain.ToJSON(e)))
		fmt.Println("-------------------------------------------------")
	}

}

func (r *PostgresRepository[E, M, Q]) rowsToEntity(columns []string, rows *sql.Rows) utils.Result[E] {
	values := make([]interface{}, len(columns))
	pointers := make([]interface{}, len(columns))
	for i := range values {
		pointers[i] = &values[i]
	}

	// Escanear la fila.
	if err := rows.Scan(pointers...); err != nil {
		return utils.Result[E]{Err: cerrs.NewCustomError(http.StatusInternalServerError, err.Error(), "postgres.matching_low.scan")}
	}

	// Construir el mapa para la fila.
	rowMap := make(map[string]interface{})
	for i, colName := range columns {
		// Si el valor es []byte, se convierte a string.
		if b, ok := values[i].([]byte); ok {
			rowMap[colName] = string(b)
		} else {
			rowMap[colName] = values[i]
		}
	}

	entity, err := domain.FromJSON[E](rowMap)
	if err != nil {
		return utils.Result[E]{Err: cerrs.NewCustomError(http.StatusInternalServerError, err.Error(), "postgres.matching_low.from_json")}
	}

	return utils.Result[E]{Data: entity}
}
