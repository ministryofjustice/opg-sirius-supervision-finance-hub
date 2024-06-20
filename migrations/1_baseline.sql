-- +goose Up
CREATE ROLE api;

CREATE TABLE public.assignees
(
    id          INTEGER      NOT NULL
        PRIMARY KEY,
    name        VARCHAR(255) NOT NULL,
    type        VARCHAR(255) NOT NULL,
    email       VARCHAR(255) DEFAULT NULL::CHARACTER VARYING,
    surname     VARCHAR(255) DEFAULT NULL::CHARACTER VARYING,
    roles       JSON,
    suspended   BOOLEAN      DEFAULT FALSE,
    phonenumber VARCHAR(255) DEFAULT NULL::CHARACTER VARYING,
    deleted     TIMESTAMP(0),
    teamtype    VARCHAR(255) DEFAULT NULL::CHARACTER VARYING,
    updateddate TIMESTAMP(0) DEFAULT NULL::TIMESTAMP WITHOUT TIME ZONE,
    permanent   BOOLEAN      DEFAULT FALSE
);

ALTER TABLE public.assignees
    OWNER TO api;

CREATE SCHEMA supervision_finance;
GRANT ALL ON SCHEMA supervision_finance TO api;

CREATE SEQUENCE supervision_finance.billing_period_id_seq;

ALTER SEQUENCE supervision_finance.billing_period_id_seq OWNER TO api;

CREATE SEQUENCE supervision_finance.counter_id_seq;

ALTER SEQUENCE supervision_finance.counter_id_seq OWNER TO api;

CREATE SEQUENCE supervision_finance.fee_reduction_id_seq;

ALTER SEQUENCE supervision_finance.fee_reduction_id_seq OWNER TO api;

CREATE SEQUENCE supervision_finance.finance_client_id_seq;

ALTER SEQUENCE supervision_finance.finance_client_id_seq OWNER TO api;

CREATE SEQUENCE supervision_finance.invoice_email_status_id_seq;

ALTER SEQUENCE supervision_finance.invoice_email_status_id_seq OWNER TO api;

CREATE SEQUENCE supervision_finance.invoice_fee_range_id_seq;

ALTER SEQUENCE supervision_finance.invoice_fee_range_id_seq OWNER TO api;

CREATE SEQUENCE supervision_finance.invoice_id_seq;

ALTER SEQUENCE supervision_finance.invoice_id_seq OWNER TO api;

CREATE SEQUENCE supervision_finance.ledger_allocation_id_seq;

ALTER SEQUENCE supervision_finance.ledger_allocation_id_seq OWNER TO api;

CREATE SEQUENCE supervision_finance.ledger_id_seq;

ALTER SEQUENCE supervision_finance.ledger_id_seq OWNER TO api;

CREATE SEQUENCE supervision_finance.property_id_seq;

ALTER SEQUENCE supervision_finance.property_id_seq OWNER TO api;

CREATE SEQUENCE supervision_finance.rate_id_seq;

ALTER SEQUENCE supervision_finance.rate_id_seq OWNER TO api;

CREATE SEQUENCE supervision_finance.report_id_seq;

ALTER SEQUENCE supervision_finance.report_id_seq OWNER TO api;

CREATE TABLE supervision_finance.counter
(
    id      INTEGER     NOT NULL
        PRIMARY KEY,
    key     VARCHAR(50) NOT NULL,
    counter INTEGER     NOT NULL
);

ALTER TABLE supervision_finance.counter
    OWNER TO api;

CREATE INDEX idx_counter_key
    ON supervision_finance.counter (key);

CREATE UNIQUE INDEX uniq_26df0c148a90aba9
    ON supervision_finance.counter (key);

CREATE TABLE supervision_finance.finance_client
(
    id             INTEGER      NOT NULL
        PRIMARY KEY,
    client_id      INTEGER      NOT NULL,
    sop_number     TEXT         NOT NULL,
    payment_method VARCHAR(255) NOT NULL,
    batchnumber    INTEGER
);

COMMENT ON COLUMN supervision_finance.finance_client.payment_method IS '(DC2Type:refdata)';

ALTER TABLE supervision_finance.finance_client
    OWNER TO api;

CREATE TABLE supervision_finance.billing_period
(
    id                INTEGER NOT NULL
        PRIMARY KEY,
    finance_client_id INTEGER
        CONSTRAINT fk_f586876342ac816b
            REFERENCES supervision_finance.finance_client,
    order_id          INTEGER,
    start_date        DATE    NOT NULL,
    end_date          DATE
);

ALTER TABLE supervision_finance.billing_period
    OWNER TO api;

CREATE INDEX idx_c64d624c7a3c530d
    ON supervision_finance.billing_period (finance_client_id);

CREATE TABLE supervision_finance.fee_reduction
(
    id                INTEGER                    NOT NULL
        PRIMARY KEY,
    finance_client_id INTEGER
        CONSTRAINT fk_6ab78de42ac816b
            REFERENCES supervision_finance.finance_client,
    type              VARCHAR(255)               NOT NULL,
    evidencetype      VARCHAR(255) DEFAULT NULL::CHARACTER VARYING,
    startdate         DATE                       NOT NULL,
    enddate           DATE                       NOT NULL,
    notes             TEXT                       NOT NULL,
    deleted           BOOLEAN      DEFAULT FALSE NOT NULL,
    datereceived      DATE
);

COMMENT ON COLUMN supervision_finance.fee_reduction.type IS '(DC2Type:refdata)';

COMMENT ON COLUMN supervision_finance.fee_reduction.evidencetype IS '(DC2Type:refdata)';

ALTER TABLE supervision_finance.fee_reduction
    OWNER TO api;

CREATE INDEX idx_690054cf7a3c530d
    ON supervision_finance.fee_reduction (finance_client_id);

CREATE INDEX idx_finance_client_batch_number
    ON supervision_finance.finance_client (batchnumber);

CREATE TABLE supervision_finance.invoice
(
    id                INTEGER     NOT NULL
        PRIMARY KEY,
    person_id         INTEGER,
    finance_client_id INTEGER
        CONSTRAINT fk_7df7fbe042ac816b
            REFERENCES supervision_finance.finance_client
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

COMMENT ON COLUMN supervision_finance.invoice.amount IS '(DC2Type:money)';

COMMENT ON COLUMN supervision_finance.invoice.supervisionlevel IS '(DC2Type:refdata)';

COMMENT ON COLUMN supervision_finance.invoice.cacheddebtamount IS '(DC2Type:money)';

ALTER TABLE supervision_finance.invoice
    OWNER TO api;

CREATE INDEX idx_77988f287a3c530d
    ON supervision_finance.invoice (finance_client_id);

CREATE INDEX idx_invoice_batch_number
    ON supervision_finance.invoice (batchnumber);

CREATE UNIQUE INDEX uniq_77988f28aea34913
    ON supervision_finance.invoice (reference);

CREATE TABLE supervision_finance.invoice_email_status
(
    id          INTEGER      NOT NULL
        PRIMARY KEY,
    invoice_id  INTEGER
        CONSTRAINT fk_64081dd12989f1fd
            REFERENCES supervision_finance.invoice
            ON DELETE CASCADE,
    status      VARCHAR(255) NOT NULL,
    templateid  VARCHAR(255) NOT NULL,
    createddate DATE
);

COMMENT ON COLUMN supervision_finance.invoice_email_status.status IS '(DC2Type:refdata)';

COMMENT ON COLUMN supervision_finance.invoice_email_status.templateid IS '(DC2Type:refdata)';

ALTER TABLE supervision_finance.invoice_email_status
    OWNER TO api;

CREATE INDEX idx_d0ae32bc2989f1fd
    ON supervision_finance.invoice_email_status (invoice_id);

CREATE TABLE supervision_finance.invoice_fee_range
(
    id               INTEGER      NOT NULL
        PRIMARY KEY,
    invoice_id       INTEGER
        CONSTRAINT fk_36446bf82989f1fd
            REFERENCES supervision_finance.invoice
            ON DELETE CASCADE,
    supervisionlevel VARCHAR(255) NOT NULL,
    fromdate         DATE         NOT NULL,
    todate           DATE         NOT NULL,
    amount           INTEGER      NOT NULL
);

COMMENT ON COLUMN supervision_finance.invoice_fee_range.supervisionlevel IS '(DC2Type:refdata)';

COMMENT ON COLUMN supervision_finance.invoice_fee_range.amount IS '(DC2Type:money)';

ALTER TABLE supervision_finance.invoice_fee_range
    OWNER TO api;

CREATE INDEX idx_5dd85a2d2989f1fd
    ON supervision_finance.invoice_fee_range (invoice_id);

CREATE TABLE supervision_finance.ledger
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
            REFERENCES supervision_finance.finance_client
            ON DELETE CASCADE,
    parent_id         INTEGER
        CONSTRAINT fk_ea14203c727aca70
            REFERENCES supervision_finance.ledger
            ON DELETE CASCADE,
    fee_reduction_id  INTEGER
        CONSTRAINT fk_ea14203c47b45492
            REFERENCES supervision_finance.fee_reduction
            ON DELETE CASCADE,
    confirmeddate     DATE,
    bankdate          DATE,
    batchnumber       INTEGER,
    bankaccount       VARCHAR(255) DEFAULT NULL::CHARACTER VARYING,
    source            VARCHAR(20)  DEFAULT NULL::CHARACTER VARYING,
    line              INTEGER
);

COMMENT ON COLUMN supervision_finance.ledger.amount IS '(DC2Type:money)';

COMMENT ON COLUMN supervision_finance.ledger.type IS '(DC2Type:refdata)';

COMMENT ON COLUMN supervision_finance.ledger.status IS '(DC2Type:refdata)';

COMMENT ON COLUMN supervision_finance.ledger.bankaccount IS '(DC2Type:refdata)';

ALTER TABLE supervision_finance.ledger
    OWNER TO api;

CREATE INDEX idx_85cecfb26abf21a3
    ON supervision_finance.ledger (fee_reduction_id);

CREATE INDEX idx_85cecfb2727aca70
    ON supervision_finance.ledger (parent_id);

CREATE INDEX idx_85cecfb27a3c530d
    ON supervision_finance.ledger (finance_client_id);

CREATE INDEX idx_ledger_batch_number
    ON supervision_finance.ledger (batchnumber);

CREATE UNIQUE INDEX uniq_85cecfb2aea34913
    ON supervision_finance.ledger (reference);

CREATE TABLE supervision_finance.ledger_allocation
(
    id            INTEGER      NOT NULL
        PRIMARY KEY,
    ledger_id     INTEGER
        CONSTRAINT fk_b11e238deb264cb8
            REFERENCES supervision_finance.ledger
            ON DELETE CASCADE,
    invoice_id    INTEGER
        CONSTRAINT fk_b11e238d2989f1fd
            REFERENCES supervision_finance.invoice
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

COMMENT ON COLUMN supervision_finance.ledger_allocation.amount IS '(DC2Type:money)';

COMMENT ON COLUMN supervision_finance.ledger_allocation.status IS '(DC2Type:refdata)';

ALTER TABLE supervision_finance.ledger_allocation
    OWNER TO api;

CREATE INDEX idx_da8212582989f1fd
    ON supervision_finance.ledger_allocation (invoice_id);

CREATE INDEX idx_da821258a7b913dd
    ON supervision_finance.ledger_allocation (ledger_id);

CREATE INDEX idx_ledger_allocation_batch_number
    ON supervision_finance.ledger_allocation (batchnumber);

CREATE UNIQUE INDEX uniq_da821258aea34913
    ON supervision_finance.ledger_allocation (reference);

CREATE TABLE supervision_finance.property
(
    id    INTEGER      NOT NULL
        PRIMARY KEY,
    key   VARCHAR(100) NOT NULL,
    value VARCHAR(255) NOT NULL
);

ALTER TABLE supervision_finance.property
    OWNER TO api;

CREATE UNIQUE INDEX uniq_cf11cc358a90aba9
    ON supervision_finance.property (key);

CREATE TABLE supervision_finance.rate
(
    id        INTEGER     NOT NULL
        PRIMARY KEY,
    type      VARCHAR(50) NOT NULL,
    startdate DATE,
    enddate   DATE,
    amount    INTEGER     NOT NULL
);

COMMENT ON COLUMN supervision_finance.rate.amount IS '(DC2Type:money)';

ALTER TABLE supervision_finance.rate
    OWNER TO api;

CREATE TABLE supervision_finance.report
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

COMMENT ON COLUMN supervision_finance.report.type IS '(DC2Type:refdata)';

COMMENT ON COLUMN supervision_finance.report.totalamount IS '(DC2Type:money)';

ALTER TABLE supervision_finance.report
    OWNER TO api;

CREATE INDEX idx_819a1c8ae1f44b34
    ON supervision_finance.report (createdbyuser_id);

CREATE UNIQUE INDEX uniq_819a1c8a36967d99
    ON supervision_finance.report (batchnumber);

-- +goose Down

-- Baseline migration - no down