CREATE TABLE "user" (
    "user_id" serial,
    "name" text,"email" text,
    "date_created" timestamptz,
    "date_modified" timestamptz,
    PRIMARY KEY ("user_id")
);
