CREATE TABLE "administrator" (
    "administrator_id" serial,
    "username" text,"password" text,
    "date_created" timestamptz,
    "date_modified" timestamptz,
    PRIMARY KEY ("administrator_id")
);
