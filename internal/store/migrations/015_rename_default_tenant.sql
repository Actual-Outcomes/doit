-- +goose Up
UPDATE tenant SET name = 'Actual Outcomes', slug = 'actual-outcomes' WHERE slug = 'default';

-- +goose Down
UPDATE tenant SET name = 'Default', slug = 'default' WHERE slug = 'actual-outcomes';
