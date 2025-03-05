alter table "public"."bookmarks" alter column "recipe_id" set default nextval('bookmarks_recipe_id_seq'::regclass);
