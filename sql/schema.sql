drop table if exists tb_memo;

create table tb_memo
(
    id          integer primary key,
    memoId      integer not null,
    content     text,
    author      text,
    website     text,
    publishTime timestamp,
    tags        text,
    email       text,
    userId      integer,
    created     timestamp,
    updated     timestamp,
    resources   text,
    avatarUrl   text
);

CREATE UNIQUE INDEX `t_memo_uni1` ON `tb_memo` (`memoId`, `userId`, `website`);

drop table if exists tb_square_whitelist;
create table tb_square_whitelist
(
    id    integer primary key,
    token text
);
CREATE UNIQUE INDEX `square_whitelist_uni1` ON `tb_square_whitelist` (`token`);
