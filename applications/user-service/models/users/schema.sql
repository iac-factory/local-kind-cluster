--
-- User
--

CREATE TABLE "User"
(
    "id"           bigserial
        CONSTRAINT "user-id-primary-key" primary key,

    "name"         varchar(255)             default NULL::character varying,
    "display-name" text                                  null,

    "email"        varchar(255)                          not null
        CONSTRAINT "user-email-validation-constraint" CHECK ("User"."email" ~* '^[A-Za-z0-9._+%-]+@[A-Za-z0-9.-]+[.][A-Za-z]+$')
        CONSTRAINT "user-email-unique-constraint" unique,

    "avatar"       text                     default NULL,

    "marketing"    boolean                  default true not null,

    "creation"     timestamp with time zone default now(),
    "modification" timestamp with time zone,
    "deletion"     timestamp with time zone
);

CREATE INDEX IF NOT EXISTS "user-deletion-index" on "User" (deletion);
CREATE INDEX IF NOT EXISTS "user-email-index" on "User" (email);
