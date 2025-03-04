alter table "public"."steps" drop constraint "steps_recipe_id_fkey",
  add constraint "steps_recipe_id_fkey"
  foreign key ("recipe_id")
  references "public"."recipes"
  ("id") on update no action on delete cascade;
