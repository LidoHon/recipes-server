alter table "public"."like" add column "created_at" timestamptz
 null default now();
