alter table "public"."sold_recipes" drop constraint "sold_recipes_seller_id_fkey",
  add constraint "sold_recipes_seller_id_fkey"
  foreign key ("seller_id")
  references "public"."users"
  ("id") on update cascade on delete cascade;
