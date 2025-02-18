package gorm

import (
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"gorm.io/gorm/schema"
	"reflect"
)

func scanMapList(initialized bool, rows Rows, db *DB, values []any, columns []string, dest any) []map[string]any {
	list, ok := dest.(*[]map[string]any)
	if !ok {
		if v, ok := dest.([]map[string]any); ok {
			list = &v
		}
	}
	columnTypes, _ := rows.ColumnTypes()
	for initialized || rows.Next() {
		prepareValues(values, db, columnTypes, columns)

		initialized = false
		db.RowsAffected++
		db.AddError(rows.Scan(values...))

		mapValue := map[string]any{}
		_scanIntoMap(mapValue, values, columns, db)

		*list = append(*list, mapValue)
	}
	return *list
}

func _scanIntoMap(mapValue map[string]interface{}, values []interface{}, columns []string, db *DB) {
	defer func() {
		if r := recover(); r != nil {
			panic(r)
		}
	}()
	for idx, column := range columns {
		value := values[idx]
		field, ok := db.Statement.Schema.FieldsByDBName[column]
		if ok {
			column = field.Name
			if field.DataType == schema.Object || field.DataType == schema.Array {
				val := getJsonText(value)
				var data any
				if field.DataType == schema.Object {
					data = make(map[string]any)
				} else {
					data = make([]any, 0)
				}
				if len(val) > 0 {
					err := json.Unmarshal([]byte(val), &data)
					if err != nil {
						panic(err)
					}
				}
				mapValue[column] = data
				continue
			}
		}
		if reflectValue := reflect.Indirect(reflect.Indirect(reflect.ValueOf(values[idx]))); reflectValue.IsValid() {
			mapValue[column] = reflectValue.Interface()
			if valuer, ok := mapValue[column].(driver.Valuer); ok {
				mapValue[column], _ = valuer.Value()
			} else if b, ok := mapValue[column].(sql.RawBytes); ok {
				mapValue[column] = string(b)
			}
		} else {
			mapValue[column] = nil
		}
	}
}

func getJsonText(value any) string {
	res := ""
	if s, ok := value.(**string); ok {
		if s != nil && *s != nil {
			res = **s
		}

	} else if s, ok := value.(*string); ok {
		if s != nil {
			res = *s
		}
	} else if s, ok := value.(string); ok {
		res = s
	}
	return res
}

func scanMap(initialized bool, rows Rows, db *DB, values []any, columns []string, dest any) error {
	if initialized || rows.Next() {
		columnTypes, _ := rows.ColumnTypes()
		prepareValues(values, db, columnTypes, columns)

		db.RowsAffected++
		db.AddError(rows.Scan(values...))

		mapValue, ok := dest.(map[string]interface{})
		if !ok {
			if v, ok := dest.(*map[string]interface{}); ok {
				if *v == nil {
					*v = map[string]interface{}{}
				}
				mapValue = *v
			}
		}
		if mapValue == nil {
			mapValue = map[string]interface{}{}
		}
		scanIntoMap(mapValue, values, columns)
	}
	return nil
}
