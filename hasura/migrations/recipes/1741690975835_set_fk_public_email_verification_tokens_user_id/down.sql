alter table "public"."email_verification_tokens" drop constraint "email_verification_tokens_user_id_fkey",
  add constraint "email_verification_tokens_user_id_fkey"
  foreign key ("user_id")
  references "public"."users"
  ("id") on update no action on delete no action;
