alter table "public"."sold_recipes"
  add constraint "sold_recipes_seller_id_fkey"
  foreign key ("seller_id")
  references "public"."users"
  ("id") on update no action on delete cascade;
