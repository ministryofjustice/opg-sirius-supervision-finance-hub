-- +goose Up
CREATE ROLE api;

CREATE TABLE public.persons
(
    id            INTEGER NOT NULL
        PRIMARY KEY,
    firstname     VARCHAR(255) DEFAULT NULL,
    surname       VARCHAR(255) DEFAULT NULL,
    caserecnumber VARCHAR(255) DEFAULT NULL,
    feepayer_id   INTEGER      DEFAULT NULL
        CONSTRAINT fk_a25cc7d3aff282de
            REFERENCES public.persons,
    deputytype    VARCHAR(255) DEFAULT NULL
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

CREATE SEQUENCE persons_id_seq;

ALTER SEQUENCE persons_id_seq OWNER TO api;

CREATE SEQUENCE cases_id_seq;

ALTER SEQUENCE cases_id_seq OWNER TO api;

CREATE SCHEMA supervision_finance;
GRANT ALL ON SCHEMA supervision_finance TO api;
SET SEARCH_PATH TO supervision_finance;

CREATE SEQUENCE billing_period_id_seq;

ALTER SEQUENCE billing_period_id_seq OWNER TO api;

CREATE SEQUENCE counter_id_seq;

ALTER SEQUENCE counter_id_seq OWNER TO api;

CREATE SEQUENCE fee_reduction_id_seq;

ALTER SEQUENCE fee_reduction_id_seq OWNER TO api;

CREATE SEQUENCE finance_client_id_seq;

ALTER SEQUENCE finance_client_id_seq OWNER TO api;

CREATE SEQUENCE invoice_email_status_id_seq;

ALTER SEQUENCE invoice_email_status_id_seq OWNER TO api;

CREATE SEQUENCE invoice_fee_range_id_seq;

ALTER SEQUENCE invoice_fee_range_id_seq OWNER TO api;

CREATE SEQUENCE invoice_id_seq;

ALTER SEQUENCE invoice_id_seq OWNER TO api;

CREATE SEQUENCE ledger_allocation_id_seq;

ALTER SEQUENCE ledger_allocation_id_seq OWNER TO api;

CREATE SEQUENCE ledger_id_seq;

ALTER SEQUENCE ledger_id_seq OWNER TO api;

CREATE SEQUENCE property_id_seq;

ALTER SEQUENCE property_id_seq OWNER TO api;

CREATE SEQUENCE rate_id_seq;

ALTER SEQUENCE rate_id_seq OWNER TO api;

CREATE SEQUENCE report_id_seq;

ALTER SEQUENCE report_id_seq OWNER TO api;

CREATE TABLE counter
(
    id      INTEGER     NOT NULL
        PRIMARY KEY,
    key     VARCHAR(50) NOT NULL,
    counter INTEGER     NOT NULL
);

ALTER TABLE counter
    OWNER TO api;

CREATE INDEX idx_counter_key
    ON counter (key);

CREATE UNIQUE INDEX uniq_26df0c148a90aba9
    ON counter (key);

CREATE TABLE finance_client
(
    id             INTEGER      NOT NULL
        PRIMARY KEY,
    client_id      INTEGER      NOT NULL,
    sop_number     TEXT         NOT NULL,
    payment_method VARCHAR(255) NOT NULL,
    batchnumber    INTEGER
);

COMMENT ON COLUMN finance_client.payment_method IS '(DC2Type:refdata)';

ALTER TABLE finance_client
    OWNER TO api;

CREATE TABLE billing_period
(
    id                INTEGER NOT NULL
        PRIMARY KEY,
    finance_client_id INTEGER
        CONSTRAINT fk_f586876342ac816b
            REFERENCES finance_client,
    order_id          INTEGER,
    start_date        DATE    NOT NULL,
    end_date          DATE
);

ALTER TABLE billing_period
    OWNER TO api;

CREATE INDEX idx_c64d624c7a3c530d
    ON billing_period (finance_client_id);

CREATE TABLE fee_reduction
(
    id                INTEGER                    NOT NULL
        PRIMARY KEY,
    finance_client_id INTEGER
        CONSTRAINT fk_6ab78de42ac816b
            REFERENCES finance_client,
    type              VARCHAR(255)               NOT NULL,
    evidencetype      VARCHAR(255) DEFAULT NULL::CHARACTER VARYING,
    startdate         DATE                       NOT NULL,
    enddate           DATE                       NOT NULL,
    notes             TEXT                       NOT NULL,
    deleted           BOOLEAN      DEFAULT FALSE NOT NULL,
    datereceived      DATE
);

COMMENT ON COLUMN fee_reduction.type IS '(DC2Type:refdata)';

COMMENT ON COLUMN fee_reduction.evidencetype IS '(DC2Type:refdata)';

ALTER TABLE fee_reduction
    OWNER TO api;

CREATE INDEX idx_690054cf7a3c530d
    ON fee_reduction (finance_client_id);

CREATE INDEX idx_finance_client_batch_number
    ON finance_client (batchnumber);

CREATE TABLE invoice
(
    id                INTEGER     NOT NULL
        PRIMARY KEY,
    person_id         INTEGER,
    finance_client_id INTEGER
        CONSTRAINT fk_7df7fbe042ac816b
            REFERENCES finance_client
            ON DELETE CASCADE,
    feetype           TEXT        NOT NULL,
    reference         VARCHAR(50) NOT NULL,
    startdate         DATE        NOT NULL,
    enddate           DATE        NOT NULL,
    amount            INTEGER     NOT NULL,
    supervisionlevel  VARCHAR(255) DEFAULT NULL::CHARACTER VARYING,
    confirmeddate     DATE,
    batchnumber       INTEGER,
    raiseddate        DATE,
    source            VARCHAR(20)  DEFAULT NULL::CHARACTER VARYING,
    scheduledfn14date DATE,
    cacheddebtamount  INTEGER
);

COMMENT ON COLUMN invoice.amount IS '(DC2Type:money)';

COMMENT ON COLUMN invoice.supervisionlevel IS '(DC2Type:refdata)';

COMMENT ON COLUMN invoice.cacheddebtamount IS '(DC2Type:money)';

ALTER TABLE invoice
    OWNER TO api;

CREATE INDEX idx_77988f287a3c530d
    ON invoice (finance_client_id);

CREATE INDEX idx_invoice_batch_number
    ON invoice (batchnumber);

CREATE UNIQUE INDEX uniq_77988f28aea34913
    ON invoice (reference);

CREATE TABLE invoice_email_status
(
    id          INTEGER      NOT NULL
        PRIMARY KEY,
    invoice_id  INTEGER
        CONSTRAINT fk_64081dd12989f1fd
            REFERENCES invoice
            ON DELETE CASCADE,
    status      VARCHAR(255) NOT NULL,
    templateid  VARCHAR(255) NOT NULL,
    createddate DATE
);

COMMENT ON COLUMN invoice_email_status.status IS '(DC2Type:refdata)';

COMMENT ON COLUMN invoice_email_status.templateid IS '(DC2Type:refdata)';

ALTER TABLE invoice_email_status
    OWNER TO api;

CREATE INDEX idx_d0ae32bc2989f1fd
    ON invoice_email_status (invoice_id);

CREATE TABLE invoice_fee_range
(
    id               INTEGER      NOT NULL
        PRIMARY KEY,
    invoice_id       INTEGER
        CONSTRAINT fk_36446bf82989f1fd
            REFERENCES invoice
            ON DELETE CASCADE,
    supervisionlevel VARCHAR(255) NOT NULL,
    fromdate         DATE         NOT NULL,
    todate           DATE         NOT NULL,
    amount           INTEGER      NOT NULL
);

COMMENT ON COLUMN invoice_fee_range.supervisionlevel IS '(DC2Type:refdata)';

COMMENT ON COLUMN invoice_fee_range.amount IS '(DC2Type:money)';

ALTER TABLE invoice_fee_range
    OWNER TO api;

CREATE INDEX idx_5dd85a2d2989f1fd
    ON invoice_fee_range (invoice_id);

CREATE TABLE ledger
(
    id                INTEGER                                      NOT NULL
        PRIMARY KEY,
    reference         VARCHAR(50)                                  NOT NULL,
    datetime          TIMESTAMP(0)                                 NOT NULL,
    method            VARCHAR(255)                                 NOT NULL,
    amount            INTEGER                                      NOT NULL,
    notes             TEXT,
    type              VARCHAR(255)                                 NOT NULL,
    status            VARCHAR(255) DEFAULT NULL::CHARACTER VARYING NOT NULL,
    finance_client_id INTEGER
        CONSTRAINT fk_ea14203c42ac816b
            REFERENCES finance_client
            ON DELETE CASCADE,
    parent_id         INTEGER
        CONSTRAINT fk_ea14203c727aca70
            REFERENCES ledger
            ON DELETE CASCADE,
    fee_reduction_id  INTEGER
        CONSTRAINT fk_ea14203c47b45492
            REFERENCES fee_reduction
            ON DELETE CASCADE,
    confirmeddate     DATE,
    bankdate          DATE,
    batchnumber       INTEGER,
    bankaccount       VARCHAR(255) DEFAULT NULL::CHARACTER VARYING,
    source            VARCHAR(20)  DEFAULT NULL::CHARACTER VARYING,
    line              INTEGER
);

COMMENT ON COLUMN ledger.amount IS '(DC2Type:money)';

COMMENT ON COLUMN ledger.type IS '(DC2Type:refdata)';

COMMENT ON COLUMN ledger.status IS '(DC2Type:refdata)';

COMMENT ON COLUMN ledger.bankaccount IS '(DC2Type:refdata)';

ALTER TABLE ledger
    OWNER TO api;

CREATE INDEX idx_85cecfb26abf21a3
    ON ledger (fee_reduction_id);

CREATE INDEX idx_85cecfb2727aca70
    ON ledger (parent_id);

CREATE INDEX idx_85cecfb27a3c530d
    ON ledger (finance_client_id);

CREATE INDEX idx_ledger_batch_number
    ON ledger (batchnumber);

CREATE UNIQUE INDEX uniq_85cecfb2aea34913
    ON ledger (reference);

CREATE TABLE ledger_allocation
(
    id            INTEGER      NOT NULL
        PRIMARY KEY,
    ledger_id     INTEGER
        CONSTRAINT fk_b11e238deb264cb8
            REFERENCES ledger
            ON DELETE CASCADE,
    invoice_id    INTEGER
        CONSTRAINT fk_b11e238d2989f1fd
            REFERENCES invoice
            ON DELETE CASCADE,
    datetime      TIMESTAMP(0) NOT NULL,
    amount        INTEGER      NOT NULL,
    status        VARCHAR(255) NOT NULL,
    reference     VARCHAR(25) DEFAULT NULL::CHARACTER VARYING,
    notes         TEXT,
    allocateddate DATE,
    batchnumber   INTEGER,
    source        VARCHAR(20) DEFAULT NULL::CHARACTER VARYING
);

COMMENT ON COLUMN ledger_allocation.amount IS '(DC2Type:money)';

COMMENT ON COLUMN ledger_allocation.status IS '(DC2Type:refdata)';

ALTER TABLE ledger_allocation
    OWNER TO api;

CREATE INDEX idx_da8212582989f1fd
    ON ledger_allocation (invoice_id);

CREATE INDEX idx_da821258a7b913dd
    ON ledger_allocation (ledger_id);

CREATE INDEX idx_ledger_allocation_batch_number
    ON ledger_allocation (batchnumber);

CREATE UNIQUE INDEX uniq_da821258aea34913
    ON ledger_allocation (reference);

CREATE TABLE property
(
    id    INTEGER      NOT NULL
        PRIMARY KEY,
    key   VARCHAR(100) NOT NULL,
    value VARCHAR(255) NOT NULL
);

ALTER TABLE property
    OWNER TO api;

CREATE UNIQUE INDEX uniq_cf11cc358a90aba9
    ON property (key);

CREATE TABLE rate
(
    id        INTEGER     NOT NULL
        PRIMARY KEY,
    type      VARCHAR(50) NOT NULL,
    startdate DATE,
    enddate   DATE,
    amount    INTEGER     NOT NULL
);

COMMENT ON COLUMN rate.amount IS '(DC2Type:money)';

ALTER TABLE rate
    OWNER TO api;

CREATE TABLE report
(
    id                    INTEGER      NOT NULL
        PRIMARY KEY,
    batchnumber           INTEGER      NOT NULL,
    type                  VARCHAR(255) NOT NULL,
    datetime              TIMESTAMP(0) NOT NULL,
    count                 INTEGER      NOT NULL,
    invoicedate           TIMESTAMP(0),
    totalamount           INTEGER,
    firstinvoicereference VARCHAR(50) DEFAULT NULL::CHARACTER VARYING,
    lastinvoicereference  VARCHAR(50) DEFAULT NULL::CHARACTER VARYING,
    createdbyuser_id      INTEGER
);

COMMENT ON COLUMN report.type IS '(DC2Type:refdata)';

COMMENT ON COLUMN report.totalamount IS '(DC2Type:money)';

ALTER TABLE report
    OWNER TO api;

CREATE INDEX idx_819a1c8ae1f44b34
    ON report (createdbyuser_id);

CREATE UNIQUE INDEX uniq_819a1c8a36967d99
    ON report (batchnumber);

-- +goose Down

-- Baseline migration - no down