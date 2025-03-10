alter table "public"."like" drop constraint "like_recipe_id_fkey",
  add constraint "like_recipe_id_fkey"
  foreign key ("recipe_id")
  references "public"."recipes"
  ("id") on update no action on delete no action;
