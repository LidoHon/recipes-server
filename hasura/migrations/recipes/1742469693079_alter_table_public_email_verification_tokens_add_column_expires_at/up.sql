alter table "public"."email_verification_tokens" add column "expires_at" timestamp
 null default (now() + '24:00:00'::interval);
