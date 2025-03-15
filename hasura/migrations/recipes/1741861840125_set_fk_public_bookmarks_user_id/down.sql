alter table "public"."bookmarks" drop constraint "bookmarks_user_id_fkey",
  add constraint "bookmarks_user_id_fkey"
  foreign key ("user_id")
  references "public"."users"
  ("id") on update no action on delete cascade;
