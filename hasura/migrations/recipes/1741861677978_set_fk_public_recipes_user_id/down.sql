alter table "public"."recipes" drop constraint "recipes_user_id_fkey",
  add constraint "recipes_user_id_fkey"
  foreign key ("user_id")
  references "public"."users"
  ("id") on update no action on delete cascade;
