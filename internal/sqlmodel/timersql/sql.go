package timersql

/*
create table if not exists timers (
    id uuid not null,
    creator bigint not null,
    start_time timestamp(0) not null,
    end_time timestamp(0) not null,
    name varchar(60) not null,
    description text,
    with_music boolean not null,
    color_id smallint not null,
    type_id smallint not null,
    status_id smallint not null,
    duration bigint not null,
    utc smallint not null default 0,

    constraint fk_timers__types foreign key (type_id) references types(id),

    constraint fk_timers__colors foreign key (color_id) references colors(id),

    constraint fk_timers__timer_status foreign key (status_id) references timer_status(id),

    constraint timers_key primary key (id)
);
*/

const Table = "timers"

type timer_column string

func (t timer_column) String() string {
	return string(t)
}

func (t timer_column) Table() string {
	return Table
}

const (
	ID          timer_column = "id"
	Creator     timer_column = "creator"
	EndTime     timer_column = "end_time"
	Name        timer_column = "name"
	Description timer_column = "description"
	WithMusic   timer_column = "with_music"
	ColorId     timer_column = "color_id"
	TypeId      timer_column = "type_id"
	UTC         timer_column = "utc"
	Duration    timer_column = "duration"
	IsDeleted   timer_column = "is_deleted"
)

const (
	FK_Colors  = "fk_timers__colors"
	FK_Types   = "fk_timers__types"
	FK_Status  = "fk_timers__timer_status"
	PrimaryKey = "timers_key"
)
