alter table "public"."ratings" alter column "recipe_id" set default nextval('ratings_recipe_id_seq'::regclass);
