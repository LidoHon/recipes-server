CREATE TABLE "public"."comments" ("id" serial NOT NULL, "recipe_id" integer NOT NULL, "user_id" integer NOT NULL, "comment" varchar NOT NULL, "created_at" timestamptz NOT NULL DEFAULT now(), "updated_at" timestamptz NOT NULL DEFAULT now(), PRIMARY KEY ("id") , FOREIGN KEY ("recipe_id") REFERENCES "public"."recipes"("id") ON UPDATE no action ON DELETE no action);
CREATE OR REPLACE FUNCTION "public"."set_current_timestamp_updated_at"()
RETURNS TRIGGER AS $$
DECLARE
  _new record;
BEGIN
  _new := NEW;
  _new."updated_at" = NOW();
  RETURN _new;
END;
$$ LANGUAGE plpgsql;
CREATE TRIGGER "set_public_comments_updated_at"
BEFORE UPDATE ON "public"."comments"
FOR EACH ROW
EXECUTE PROCEDURE "public"."set_current_timestamp_updated_at"();
COMMENT ON TRIGGER "set_public_comments_updated_at" ON "public"."comments"
IS 'trigger to set value of column "updated_at" to current timestamp on row update';
