alter table "public"."ratings" drop constraint "ratings_user_id_fkey",
  add constraint "ratings_user_id_fkey"
  foreign key ("user_id")
  references "public"."users"
  ("id") on update cascade on delete cascade;
