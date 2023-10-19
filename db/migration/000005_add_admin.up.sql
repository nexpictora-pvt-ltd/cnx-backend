CREATE TABLE "admins" (
  "admin_id" serial UNIQUE PRIMARY KEY NOT NULL,
  "name" varchar NOT NULL,
  "email" varchar UNIQUE NOT NULL,
  "phone" varchar NOT NULL,
  "address" varchar NOT NULL,
  "hashed_password" varchar NOT NULL,
  "created_at" timestamptz NOT NULL DEFAULT (now()),
  "password_changed_at" timestamptz NOT NULL DEFAULT '0001-01-01 00:00:00Z'
);

-- Set the starting value of the admin_id sequence to 2300
SELECT setval('admins_admin_id_seq', 2300);

-- Adding modified_by column to the orders table
ALTER TABLE IF EXISTS "orders"
ADD COLUMN "modified_by" serial NOT NULL;

ALTER TABLE "orders" ADD FOREIGN KEY ("modified_by") REFERENCES "admins" ("admin_id");

