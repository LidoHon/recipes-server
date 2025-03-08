alter table "public"."payments"
  add constraint "payments_buyer_id_fkey"
  foreign key ("buyer_id")
  references "public"."users"
  ("id") on update no action on delete cascade;
