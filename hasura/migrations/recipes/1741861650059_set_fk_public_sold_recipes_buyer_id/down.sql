alter table "public"."sold_recipes" drop constraint "sold_recipes_buyer_id_fkey",
  add constraint "sold_recipes_buyer_id_fkey"
  foreign key ("buyer_id")
  references "public"."users"
  ("id") on update no action on delete no action;
