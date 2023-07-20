BEGIN;

create table if not exists colors (
    id smallserial primary key,
    color varchar(30) not null,

    constraint colors_unique unique (color)
);

create table if not exists types (
    id smallserial primary key,
    type varchar(30) not null,

    constraint types_unique unique (type)
);

create table if not exists timers (
    id uuid not null,
    creator bigint not null,
    end_time timestamp(0) not null,
    name varchar(60),
    description text,
    with_music boolean not null,
    color_id smallint not null,
    type_id smallint not null,
    duration bigint not null,
    utc smallint not null default 0,
    is_deleted boolean not null default false,

    constraint fk_timers__types foreign key (type_id) references types(id),

    constraint fk_timers__colors foreign key (color_id) references colors(id),

    constraint timers_key primary key (id)
);

create table if not exists timer_subcribers (
    timer_id uuid not null,
    user_id bigint not null,

    constraint fk_timer_subcribers__timers foreign key (timer_id) references timers(id) on delete cascade,

    constraint timer_subcribers_key primary key (timer_id, user_id)
);


create table if not exists countdown_timers (
    timer_id uuid not null,
    pause_time timestamp(0) default null,
    is_paused boolean default false,

    constraint fk_countdown_timers__timers foreign key (timer_id) references timers(id) on delete cascade,
    
    constraint countdown_timers_key primary key (timer_id)
);


create table if not exists notification_types (
    id smallserial not null primary key,
    type varchar(30) not null,

    constraint notifications_type_unique unique (type)
);

create table if not exists notifications (

    notification_type_id smallint not null,
    timer_id uuid not null,
    user_id bigint not null,

    constraint fk_notifications__notifications_types foreign key (notification_type_id) references notification_types(id),

    constraint fk_notifications__timers foreign key (timer_id) references timers(id) on delete cascade,

    constraint notifications_key primary key(timer_id,user_id)

);


COMMIT;


BEGIN;

INSERT INTO notification_types (type) VALUES ('notification_delete'), ('notification_expired');

INSERT INTO colors (color) VALUES ('DEFAULT'),('RED'),('GREEN'),('BLUE'),('PURPLE'),('YELLOW');

INSERT INTO types (type) VALUES ('COUNTDOWN'), ('DATE');


COMMIT;