alter table "public"."like" drop constraint "like_user_id_fkey",
  add constraint "like_user_id_fkey"
  foreign key ("user_id")
  references "public"."users"
  ("id") on update no action on delete no action;
