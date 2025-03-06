CREATE TABLE "public"."recipe_images" ("id" serial NOT NULL, "recipe_id" integer NOT NULL, "is_featured" boolean NOT NULL, "created_at" timestamptz NOT NULL DEFAULT now(), "updated_at" timestamptz NOT NULL DEFAULT now(), PRIMARY KEY ("id") , FOREIGN KEY ("recipe_id") REFERENCES "public"."recipes"("id") ON UPDATE cascade ON DELETE cascade);COMMENT ON TABLE "public"."recipe_images" IS E'this table stores multiple images for recipes';
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
CREATE TRIGGER "set_public_recipe_images_updated_at"
BEFORE UPDATE ON "public"."recipe_images"
FOR EACH ROW
EXECUTE PROCEDURE "public"."set_current_timestamp_updated_at"();
COMMENT ON TRIGGER "set_public_recipe_images_updated_at" ON "public"."recipe_images"
IS 'trigger to set value of column "updated_at" to current timestamp on row update';
