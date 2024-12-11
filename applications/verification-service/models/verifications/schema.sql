CREATE TABLE "Verification"
(
    "id"           bigserial CONSTRAINT "verification-id-primary-key" primary key,
    "email"        varchar(255) not null CONSTRAINT "verification-email-validation-constraint" CHECK ("Verification"."email" ~* '^[A-Za-z0-9._+%-]+@[A-Za-z0-9.-]+[.][A-Za-z]+$') CONSTRAINT "verification-email-unique-constraint" unique,
    "code"         varchar(32)  not null CONSTRAINT "verification-code-unique-constraint" unique,
    "verified"     bool NOT NULL default false,
    "creation"     timestamp with time zone default now(),
    "modification" timestamp with time zone,
    "deletion"     timestamp with time zone
);

CREATE INDEX IF NOT EXISTS "verification-code-index" on "Verification" (code);
CREATE INDEX IF NOT EXISTS "verification-verified-index" on "Verification" (verified);
CREATE INDEX IF NOT EXISTS "verification-email-index" on "Verification" (email);
CREATE INDEX IF NOT EXISTS "verification-deletion-index" on "Verification" (deletion);
