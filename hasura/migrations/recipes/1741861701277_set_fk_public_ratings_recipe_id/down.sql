alter table "public"."ratings" drop constraint "ratings_recipe_id_fkey",
  add constraint "ratings_recipe_id_fkey"
  foreign key ("recipe_id")
  references "public"."recipes"
  ("id") on update restrict on delete cascade;
