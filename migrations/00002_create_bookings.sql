-- +goose Up
create table if not exists bookings (
    id text primary key,
    event_type_id text not null references event_types (id) on delete cascade,
    guest_name text not null,
    guest_email text not null,
    starts_at timestamptz not null,
    ends_at timestamptz not null,
    created_at timestamptz not null default now(),
    constraint bookings_time_range_check check (starts_at < ends_at)
);

create index if not exists bookings_time_range_idx on bookings (starts_at, ends_at);

-- +goose Down
drop table if exists bookings;
