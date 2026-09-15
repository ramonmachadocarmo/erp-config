INSERT INTO settings (key, value) VALUES ('default_warehouse_id', '')
ON CONFLICT (key) DO NOTHING;
