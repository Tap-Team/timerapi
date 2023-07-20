package sqlutils

import "github.com/jackc/pgx/v5"

func ScanList[T any](rows pgx.Rows, scan func(pgx.Row, *T) error) ([]*T, error) {
	list := make([]*T, 0)
	for rows.Next() {
		v := new(T)
		err := scan(rows, v)
		if err != nil {
			return nil, err
		}
		list = append(list, v)
	}
	return list, nil
}
