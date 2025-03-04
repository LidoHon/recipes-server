alter table "public"."recipes" alter column "avarage_rating" set default 0;
alter table "public"."recipes" alter column "avarage_rating" drop not null;
alter table "public"."recipes" add column "avarage_rating" numeric;
