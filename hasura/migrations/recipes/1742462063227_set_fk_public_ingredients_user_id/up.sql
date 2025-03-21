alter table "public"."ingredients"
  add constraint "ingredients_user_id_fkey"
  foreign key ("user_id")
  references "public"."users"
  ("id") on update cascade on delete cascade;
