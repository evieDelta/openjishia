create schema highlights;

create table highlights.guilds (
    guild_id    text    primary key,
    enabled     boolean not null    default false
);

create table highlights.highlights (
    user_id     text    not null,
    guild_id    text    not null,
    enabled     boolean not null    default true,
    words       text[]  not null    default array[]::text[],
    blocks      text[]  not null    default array[]::text[],

    primary key (user_id, guild_id)
);
