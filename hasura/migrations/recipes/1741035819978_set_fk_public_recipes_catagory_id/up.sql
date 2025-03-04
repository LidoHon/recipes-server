alter table "public"."recipes"
  add constraint "recipes_catagory_id_fkey"
  foreign key ("catagory_id")
  references "public"."catagories"
  ("id") on update no action on delete no action;
