package ppostgres

import (
	"common/domain/criteria"
	"common/utils"
	"common/utils/cerrs"
	"fmt"
	"net/http"
	"strings"
)

// MatchingLow realiza una consulta aplicando filtros definidos en criteria y devuelve las entidades resultantes.
//
// Parámetros:
//   - cr: Objeto de tipo criteria.Criteria que contiene uno o varios filtros a aplicar en la consulta.
//     Cada filtro debe tener un campo (Field), un operador (Operator) y un valor (Value).
//   - model: Puntero a una instancia del modelo que implementa el método TableName() para determinar el nombre de la tabla.
//
// Retorno:
//   - utils.Result[[]E]: Un resultado que contiene un slice de entidades del tipo E en el campo Data en caso de éxito,
//     o un error (en el campo Err) si ocurre algún fallo durante la consulta o el mapeo de las filas a entidades.
//
// Proceso interno:
//  1. Se obtiene el nombre de la tabla a partir del método TableName() del modelo.
//  2. Se construye la consulta base "SELECT * FROM <tabla>".
//  3. Se recorren los filtros obtenidos desde criteria.Criteria y se construye la cláusula WHERE utilizando
//     marcadores numerados para cada filtro.
//  4. Se ejecuta la consulta utilizando la conexión r.Conn y se obtienen las filas resultantes.
//  5. Se extraen los nombres de las columnas para mapear cada fila a una entidad mediante el método rowsToEntity.
//  6. Se acumulan las entidades en un slice que se retorna dentro de utils.Result[[]E].
//
// Ejemplo de uso:
//
// func (r *UserPostgresRepository) Matching(cr criteria.Criteria) utils.Result[[]entities.User] {
//
//		model := &UserModel{}
//
//		return r.MatchingLow(cr, model)
//	}
func (r *PostgresRepository[E, M, Q]) MatchingLow(cr criteria.Criteria, table_name string, offset int, limit int) utils.Result[[]E] {

	tableName := table_name

	// Construir la consulta base.
	queryStr := "SELECT * FROM " + tableName

	var whereClauses []string
	var args []interface{}

	// Recorrer los filtros para construir la cláusula WHERE usando marcadores numerados.
	for i, f := range cr.Filters.Get() {
		placeholder := fmt.Sprintf("$%d", i+1)
		whereClauses = append(whereClauses, fmt.Sprintf("%s %s %s", f.Field, f.Operator, placeholder))
		args = append(args, f.Value)
	}

	if len(whereClauses) > 0 {
		queryStr += " WHERE " + strings.Join(whereClauses, " AND ")
	}

	// Ejecutar la consulta usando la conexión Conn.
	rows, err := r.Conn.Query(queryStr, args...)
	if err != nil {
		return utils.Result[[]E]{Err: cerrs.NewCustomError(http.StatusInternalServerError, err.Error(), "postgres.matching_low.get_rows")}
	}
	defer rows.Close()

	// Obtener los nombres de las columnas para mapear cada fila.
	columns, err := rows.Columns()
	if err != nil {
		return utils.Result[[]E]{Err: cerrs.NewCustomError(http.StatusInternalServerError, err.Error(), "postgres.matching_low.get_columns")}
	}

	var entities []E

	// Aquí mapeamos cada fila a un map[string]interface{}
	for rows.Next() {

		entity := r.rowsToEntity(columns, rows)
		if entity.Err != nil {
			return utils.Result[[]E]{Err: entity.Err}
		}

		entities = append(entities, entity.Data)
	}

	if err := rows.Err(); err != nil {
		return utils.Result[[]E]{Err: cerrs.NewCustomError(http.StatusInternalServerError, err.Error(), "postgres.matching_low.get_rows_err")}
	}

	return utils.Result[[]E]{Data: entities}
}
