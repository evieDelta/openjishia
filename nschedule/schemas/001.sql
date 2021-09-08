CREATE SCHEMA nscheduler;

CREATE TABLE nscheduler.events
(
    "id" bigserial NOT NULL,
    "action" text NOT NULL,
    "details" jsonb NOT NULL,
    "defer_count" integer NOT NULL,
    "time" timestamp without time zone NOT NULL,
    PRIMARY KEY (id)
);

CREATE INDEX index_time_get
    ON nscheduler.events USING btree
    ("action" ASC NULLS LAST, "time" ASC NULLS LAST)
    INCLUDE("id", "details", "defer_count")
;