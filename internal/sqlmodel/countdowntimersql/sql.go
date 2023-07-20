package countdowntimersql

/*
create table if not exists countdown_timers (
    timer_id uuid not null,
    pause_time timestamp(0),
    is_paused boolean,

    constraint fk_countdown_timers__timers foreign key (timer_id) references timers(id),

    constraint countdown_timers_key primary key (timer_id)
);
*/

const Table = "countdown_timers"

type ctrl_column string

func (c ctrl_column) String() string {
	return string(c)
}

func (c ctrl_column) Table() string {
	return Table
}

const (
	TimerId   ctrl_column = "timer_id"
	PauseTime ctrl_column = "pause_time"
	IsPaused  ctrl_column = "is_paused"
)

const (
	FK_Timers  = "fk_countdown_timers__timers"
	PrimaryKey = "countdown_timers_key"
)
