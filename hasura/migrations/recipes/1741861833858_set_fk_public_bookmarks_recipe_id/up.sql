alter table "public"."bookmarks" drop constraint "bookmarks_recipe_id_fkey",
  add constraint "bookmarks_recipe_id_fkey"
  foreign key ("recipe_id")
  references "public"."recipes"
  ("id") on update cascade on delete cascade;
