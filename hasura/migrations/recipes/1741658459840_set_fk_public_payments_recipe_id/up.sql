alter table "public"."payments" drop constraint "payments_recipe_id_fkey",
  add constraint "payments_recipe_id_fkey"
  foreign key ("recipe_id")
  references "public"."recipes"
  ("id") on update no action on delete cascade;
