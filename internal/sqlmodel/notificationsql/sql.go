package notificationsql

/*
create table if not exists notifications (

    notification_type_id smallint not null,
    timer_id uuid not null,
    user_id bigint not null,

    constraint fk_notifications__notifications_types foreign key (notification_type_id) references notifications_types(id),

    constraint fk_notifications__timers foreign key (timer_id) references timers(id) on delete cascade,

    constraint notifications_key primary key(timer_id,user_id)

);
*/

const Table = "notifications"

type notification_column string

func (n notification_column) String() string {
	return string(n)
}

func (n notification_column) Table() string {
	return Table
}

const (
	TimerId            notification_column = "timer_id"
	UserId             notification_column = "user_id"
	NotificationTypeId notification_column = "notification_type_id"
)

const (
	FK_NotificationsType = "fk_notifications__notifications_types"
	FK_Timers            = "fk_notifications__timers"
	PrimaryKey           = "notifications_key"
)
