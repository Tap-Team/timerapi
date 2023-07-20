package colorsql

/*
create table if not exists colors (
    id smallserial primary key,
    color varchar(30) not null,

    constraint colors_unique unique (color)
);
*/

const Table = "colors"

type colors_column string

func (c colors_column) Table() string {
	return Table
}

func (c colors_column) String() string {
	return string(c)
}

const (
	ID    colors_column = "id"
	Color colors_column = "color"
)

const (
	ColorsUnique = "colors_unique"
)
