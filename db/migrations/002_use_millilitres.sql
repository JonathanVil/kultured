ALTER TABLE batches RENAME COLUMN tea_volume_l TO tea_volume_ml;
UPDATE batches SET tea_volume_ml = tea_volume_ml * 1000;
