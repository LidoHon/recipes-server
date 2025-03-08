alter table "public"."payments"
  add constraint "payments_sold_recipe_id_fkey"
  foreign key ("sold_recipe_id")
  references "public"."sold_recipes"
  ("id") on update no action on delete cascade;
