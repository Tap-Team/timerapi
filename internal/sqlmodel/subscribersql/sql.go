package subscribersql

/*
create table if not exists timer_subcribers (
    timer_id uuid not null,
    user_id bigint not null,

    constraint fk_timer_subcribers__timers foreign key (timer_id) references timers(id),

    constraint timer_subcribers_key primary key (timer_id, user_id)
);
*/

const Table = "timer_subcribers"

type subcriber_column string

func (s subcriber_column) String() string {
	return string(s)
}

func (s subcriber_column) Table() string {
	return Table
}

const (
	TimerId subcriber_column = "timer_id"
	UserId  subcriber_column = "user_id"
)

const (
	FK_Timers  = "fk_timer_subcribers__timers"
	PrimaryKey = "timer_subcribers_key"
)
