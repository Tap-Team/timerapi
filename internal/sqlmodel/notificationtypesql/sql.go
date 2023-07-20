package notificationtypesql

/*
create table if not exists notifications_types (
    id smallint not null primary key,
    type varchar(30) not null,

    constraint notifications_type_unique unique (type)
);
*/

const Table = "notification_types"

type notification_types_column string

func (c notification_types_column) String() string {
	return string(c)
}

func (c notification_types_column) Table() string {
	return Table
}

const (
	ID   notification_types_column = "id"
	Type notification_types_column = "type"
)

const (
	ConstraintUniqueType = "notifications_type_unique"
)
