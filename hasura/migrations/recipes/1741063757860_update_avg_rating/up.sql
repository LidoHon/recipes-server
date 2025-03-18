CREATE OR REPLACE FUNCTION update_average_rating()
RETURNS TRIGGER AS $$
BEGIN
    -- calculate the average rating for the recipe
    UPDATE recipe
    SET average_rating = COALESCE((
        SELECT AVG(rating)::NUMERIC(3, 2)
        FROM ratings
        WHERE recipe_id = NEW.recipe_id
    ), 0.00)
    WHERE id = NEW.recipe_id;

    RETURN NULL;
END;
$$ LANGUAGE plpgsql;
