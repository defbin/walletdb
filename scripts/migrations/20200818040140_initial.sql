-- migrate:up
create type currency as enum ('BTC', 'ETH');

create table wallets (
    id          serial primary key,
    balance     decimal             not null,
    currency    currency            not null
);

create table transactions (
    id          serial primary key,
    sender      integer references wallets(id),
    receiver    integer references wallets(id),
    amount      decimal                         not null,
    fee_amount  decimal                         not null,
    created_at  timestamptz default now()       not null
);

-- migrate:down
drop table transactions;
drop table wallets;
drop type currency;
