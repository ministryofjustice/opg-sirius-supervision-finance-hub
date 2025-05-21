-- +goose Up
CREATE ROLE api;

SET SEARCH_PATH TO public;

CREATE TABLE persons
(
    id            INTEGER NOT NULL
        PRIMARY KEY,
    salutation                                                   varchar(255) default NULL::character varying,
    firstname     VARCHAR(255) DEFAULT NULL,
    surname       VARCHAR(255) DEFAULT NULL,
    caserecnumber VARCHAR(255) DEFAULT NULL,
    feepayer_id   INTEGER      DEFAULT NULL
        CONSTRAINT fk_a25cc7d3aff282de
            REFERENCES persons,
    deputytype    VARCHAR(255) DEFAULT NULL,
    deputynumber                                                 integer,
    correspondencebywelsh                                        boolean      default false not null,
    specialcorrespondencerequirements_largeprint                 boolean      default false not null,
    organisationname                                             varchar(255) default NULL::character varying,
    email                                                        varchar(255) default NULL::character varying,
    type                                                         varchar(255)               ,
    clientstatus                                                 varchar(255) default NULL::character varying


);

ALTER TABLE public.persons
    OWNER TO api;

CREATE TABLE public.cases
(
    id          INTEGER NOT NULL
        PRIMARY KEY,
    client_id   INTEGER
        CONSTRAINT fk_1c1b038b19eb6921
            REFERENCES public.persons,
    orderstatus VARCHAR(255) DEFAULT NULL
);

CREATE INDEX cases_orderstatus_index ON public.cases (orderstatus);

CREATE INDEX idx_1c1b038b19eb6921 ON public.cases (client_id);

CREATE TABLE public.assignees
(
    id INTEGER NOT NULL PRIMARY KEY,
    name VARCHAR(255) DEFAULT NULL,
    surname VARCHAR(255) DEFAULT NULL
);

create table addresses
(
    id                  integer not null
        primary key,
    person_id           integer
        constraint fk_6fca7516217bbb47
            references public.persons
            on delete cascade,
    address_lines       json,
    town                varchar(255) default NULL::character varying,
    county              varchar(255) default NULL::character varying,
    postcode            varchar(255) default NULL::character varying,
    isairmailrequired   boolean
);

alter table addresses
    owner to api;

create index idx_6fca7516217bbb47
    on addresses (person_id);

create index idx_address_postcode
    on addresses (postcode);

create table warnings
(
    id           integer                   not null
        primary key,
    added_by     integer
        constraint fk_6949e612699b6baf
            references assignees
            on delete cascade,
    closed_by    integer
        constraint fk_6949e61288f6e01
            references public.assignees
            on delete cascade,
    warningtype  varchar(255) default NULL::character varying,
    warningtext  text,
    dateadded    timestamp(0),
    dateclosed   timestamp(0),
    systemstatus boolean      default true not null
);

alter table warnings
    owner to api;

create index idx_6949e612699b6baf
    on warnings (added_by);

create index idx_6949e61288f6e01
    on warnings (closed_by);



create table person_warning
(
    person_id  integer not null
        constraint fk_62d02f4f217bbb47
            references public.persons
            on delete cascade,
    warning_id integer not null
        constraint fk_62d02f4fbff38603
            references public.warnings
            on delete cascade,
    primary key (person_id, warning_id)
);

alter table person_warning
    owner to api;

create index idx_62d02f4f217bbb47
    on person_warning (person_id);

create index idx_62d02f4fbff38603
    on person_warning (warning_id);


CREATE SEQUENCE public.persons_id_seq;

ALTER SEQUENCE public.persons_id_seq OWNER TO api;

CREATE SEQUENCE public.cases_id_seq;

ALTER SEQUENCE public.cases_id_seq OWNER TO api;

CREATE SEQUENCE public.assignees_id_seq;

ALTER SEQUENCE public.assignees_id_seq OWNER TO api;

CREATE SEQUENCE public.addresses_id_seq;

ALTER SEQUENCE public.addresses_id_seq OWNER TO api;

CREATE SCHEMA supervision_finance;
GRANT ALL ON SCHEMA supervision_finance TO api;
SET SEARCH_PATH TO supervision_finance;

create sequence billing_period_id_seq;

alter sequence billing_period_id_seq owner to api;

create sequence counter_id_seq;

alter sequence counter_id_seq owner to api;

create sequence fee_reduction_id_seq;

alter sequence fee_reduction_id_seq owner to api;

create sequence finance_client_id_seq;

alter sequence finance_client_id_seq owner to api;

create sequence invoice_email_status_id_seq;

alter sequence invoice_email_status_id_seq owner to api;

create sequence invoice_fee_range_id_seq;

alter sequence invoice_fee_range_id_seq owner to api;

create sequence invoice_id_seq;

alter sequence invoice_id_seq owner to api;

create sequence ledger_allocation_id_seq;

alter sequence ledger_allocation_id_seq owner to api;

create sequence ledger_id_seq;

alter sequence ledger_id_seq owner to api;

create sequence property_id_seq;

alter sequence property_id_seq owner to api;

create sequence rate_id_seq;

alter sequence rate_id_seq owner to api;

create sequence report_id_seq;

alter sequence report_id_seq owner to api;

create table counter
(
    id      integer     not null
        primary key,
    key     varchar(50) not null,
    counter integer     not null
);

alter table counter
    owner to api;

create index idx_counter_key
    on counter (key);

create unique index uniq_26df0c148a90aba9
    on counter (key);

create table finance_client
(
    id             integer      not null
        primary key,
    client_id      integer      not null,
    sop_number     text         not null,
    payment_method varchar(255) not null,
    batchnumber    integer
);

comment on column finance_client.payment_method is '(DC2Type:refdata)';

alter table finance_client
    owner to api;

create table billing_period
(
    id                integer not null
        primary key,
    finance_client_id integer
        constraint fk_f586876342ac816b
            references finance_client,
    order_id          integer,
    start_date        date    not null,
    end_date          date
);

alter table billing_period
    owner to api;

create index idx_c64d624c7a3c530d
    on billing_period (finance_client_id);

create table fee_reduction
(
    id                integer                    not null
        primary key,
    finance_client_id integer
        constraint fk_6ab78de42ac816b
            references finance_client,
    type              varchar(255)               not null,
    evidencetype      varchar(255) default NULL::character varying,
    startdate         date                       not null,
    enddate           date                       not null,
    notes             text                       not null,
    deleted           boolean      default false not null,
    datereceived      date
);

comment on column fee_reduction.type is '(DC2Type:refdata)';

comment on column fee_reduction.evidencetype is '(DC2Type:refdata)';

alter table fee_reduction
    owner to api;

create index idx_690054cf7a3c530d
    on fee_reduction (finance_client_id);

create index idx_finance_client_batch_number
    on finance_client (batchnumber);

create table invoice
(
    id                integer     not null
        primary key,
    person_id         integer,
    finance_client_id integer
        constraint fk_7df7fbe042ac816b
            references finance_client
            on delete cascade,
    feetype           text        not null,
    reference         varchar(50) not null,
    startdate         date        not null,
    enddate           date        not null,
    amount            integer     not null,
    supervisionlevel  varchar(255) default NULL::character varying,
    confirmeddate     date,
    batchnumber       integer,
    raiseddate        date,
    source            varchar(20)  default NULL::character varying,
    scheduledfn14date date,
    cacheddebtamount  integer
);

comment on column invoice.amount is '(DC2Type:money)';

comment on column invoice.supervisionlevel is '(DC2Type:refdata)';

comment on column invoice.cacheddebtamount is '(DC2Type:money)';

alter table invoice
    owner to api;

create index idx_77988f287a3c530d
    on invoice (finance_client_id);

create index idx_invoice_batch_number
    on invoice (batchnumber);

create unique index uniq_77988f28aea34913
    on invoice (reference);

create table invoice_email_status
(
    id          integer      not null
        primary key,
    invoice_id  integer
        constraint fk_64081dd12989f1fd
            references invoice
            on delete cascade,
    status      varchar(255) not null,
    templateid  varchar(255) not null,
    createddate date
);

comment on column invoice_email_status.status is '(DC2Type:refdata)';

comment on column invoice_email_status.templateid is '(DC2Type:refdata)';

alter table invoice_email_status
    owner to api;

create index idx_d0ae32bc2989f1fd
    on invoice_email_status (invoice_id);

create table invoice_fee_range
(
    id               integer      not null
        primary key,
    invoice_id       integer
        constraint fk_36446bf82989f1fd
            references invoice
            on delete cascade,
    supervisionlevel varchar(255) not null,
    fromdate         date         not null,
    todate           date         not null,
    amount           integer      not null
);

comment on column invoice_fee_range.supervisionlevel is '(DC2Type:refdata)';

comment on column invoice_fee_range.amount is '(DC2Type:money)';

alter table invoice_fee_range
    owner to api;

create index idx_5dd85a2d2989f1fd
    on invoice_fee_range (invoice_id);

create table ledger
(
    id                integer                                      not null
        primary key,
    reference         varchar(50)                                  not null,
    datetime          timestamp(0)                                 not null,
    method            varchar(255)                                 not null,
    amount            integer                                      not null,
    notes             text,
    type              varchar(255)                                 not null,
    status            varchar(255) default NULL::character varying not null,
    finance_client_id integer
        constraint fk_ea14203c42ac816b
            references finance_client
            on delete cascade,
    parent_id         integer
        constraint fk_ea14203c727aca70
            references ledger
            on delete cascade,
    fee_reduction_id  integer
        constraint fk_ea14203c47b45492
            references fee_reduction
            on delete cascade,
    confirmeddate     date,
    bankdate          date,
    batchnumber       integer,
    bankaccount       varchar(255) default NULL::character varying,
    source            varchar(20)  default NULL::character varying,
    line              integer
);

comment on column ledger.amount is '(DC2Type:money)';

comment on column ledger.type is '(DC2Type:refdata)';

comment on column ledger.status is '(DC2Type:refdata)';

comment on column ledger.bankaccount is '(DC2Type:refdata)';

alter table ledger
    owner to api;

create index idx_85cecfb26abf21a3
    on ledger (fee_reduction_id);

create index idx_85cecfb2727aca70
    on ledger (parent_id);

create index idx_85cecfb27a3c530d
    on ledger (finance_client_id);

create index idx_ledger_batch_number
    on ledger (batchnumber);

create unique index uniq_85cecfb2aea34913
    on ledger (reference);

create table ledger_allocation
(
    id            integer      not null
        primary key,
    ledger_id     integer
        constraint fk_b11e238deb264cb8
            references ledger
            on delete cascade,
    invoice_id    integer
        constraint fk_b11e238d2989f1fd
            references invoice
            on delete cascade,
    datetime      timestamp(0) not null,
    amount        integer      not null,
    status        varchar(255) not null,
    reference     varchar(25) default NULL::character varying,
    notes         text,
    allocateddate date,
    batchnumber   integer,
    source        varchar(20) default NULL::character varying
);

comment on column ledger_allocation.amount is '(DC2Type:money)';

comment on column ledger_allocation.status is '(DC2Type:refdata)';

alter table ledger_allocation
    owner to api;

create index idx_da8212582989f1fd
    on ledger_allocation (invoice_id);

create index idx_da821258a7b913dd
    on ledger_allocation (ledger_id);

create index idx_ledger_allocation_batch_number
    on ledger_allocation (batchnumber);

create unique index uniq_da821258aea34913
    on ledger_allocation (reference);

create table property
(
    id    integer      not null
        primary key,
    key   varchar(100) not null,
    value varchar(255) not null
);

alter table property
    owner to api;

create unique index uniq_cf11cc358a90aba9
    on property (key);

create table rate
(
    id        integer     not null
        primary key,
    type      varchar(50) not null,
    startdate date,
    enddate   date,
    amount    integer     not null
);

comment on column rate.amount is '(DC2Type:money)';

alter table rate
    owner to api;

create table report
(
    id                    integer      not null
        primary key,
    batchnumber           integer      not null,
    type                  varchar(255) not null,
    datetime              timestamp(0) not null,
    count                 integer      not null,
    invoicedate           timestamp(0),
    totalamount           integer,
    firstinvoicereference varchar(50) default NULL::character varying,
    lastinvoicereference  varchar(50) default NULL::character varying,
    createdbyuser_id      integer
);

comment on column report.type is '(DC2Type:refdata)';

comment on column report.totalamount is '(DC2Type:money)';

alter table report
    owner to api;

create index idx_819a1c8ae1f44b34
    on report (createdbyuser_id);

create unique index uniq_819a1c8a36967d99
    on report (batchnumber);

-- +goose Down

-- Baseline migration - no down