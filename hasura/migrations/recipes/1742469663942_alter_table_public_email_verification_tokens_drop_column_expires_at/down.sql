alter table "public"."email_verification_tokens" alter column "expires_at" set default (now() + '24:00:00'::interval);
alter table "public"."email_verification_tokens" alter column "expires_at" drop not null;
alter table "public"."email_verification_tokens" add column "expires_at" timetz;
