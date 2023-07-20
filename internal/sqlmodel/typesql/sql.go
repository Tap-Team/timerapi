package typesql

/*
create table if not exists types (
    id smallserial primary key,
    type varchar(30) not null,

    constraint types_unique unique (type)
);
*/

const Table = "types"

type types_column string

func (c types_column) Table() string {
	return Table
}

func (c types_column) String() string {
	return string(c)
}

const (
	ID   types_column = "id"
	Type types_column = "type"
)

const (
	TypesUnique = "types_unique"
)
