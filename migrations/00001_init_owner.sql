-- +goose Up
create table if not exists owner_profiles (
    id text primary key,
    display_name text not null,
    timezone text not null
);

create table if not exists availability_windows (
    owner_id text not null references owner_profiles (id) on delete cascade,
    day_of_week text not null,
    start_time time not null,
    end_time time not null,
    primary key (owner_id, day_of_week),
    constraint availability_windows_day_of_week_check check (
        day_of_week in ('monday', 'tuesday', 'wednesday', 'thursday', 'friday', 'saturday', 'sunday')
    ),
    constraint availability_windows_time_range_check check (start_time < end_time)
);

create table if not exists event_types (
    id text primary key,
    title text not null,
    description text not null default '',
    duration_minutes integer not null,
    created_at timestamptz not null default now(),
    constraint event_types_duration_minutes_check check (duration_minutes > 0)
);

insert into owner_profiles (id, display_name, timezone)
select 'owner-default', 'Owner', 'Europe/Moscow'
where not exists (
    select 1
    from owner_profiles
);

insert into availability_windows (owner_id, day_of_week, start_time, end_time)
select owner.id, seeded.day_of_week, seeded.start_time::time, seeded.end_time::time
from (
    values
        ('monday', '09:00', '17:00'),
        ('tuesday', '09:00', '17:00'),
        ('wednesday', '09:00', '17:00'),
        ('thursday', '09:00', '17:00'),
        ('friday', '09:00', '17:00')
) as seeded(day_of_week, start_time, end_time)
cross join (
    select id
    from owner_profiles
    order by id
    limit 1
) as owner
where not exists (
    select 1
    from availability_windows
);

insert into event_types (id, title, description, duration_minutes)
select seeded.id, seeded.title, seeded.description, seeded.duration_minutes
from (
    values
        ('event-type-intro', 'Знакомство', '', 15),
        ('event-type-lesson', 'Урок', '', 30)
) as seeded(id, title, description, duration_minutes)
where not exists (
    select 1
    from event_types
);

-- +goose Down
drop table if exists availability_windows;
drop table if exists event_types;
drop table if exists owner_profiles;
