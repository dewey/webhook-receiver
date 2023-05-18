-- +goose Up
-- +goose StatementBegin
CREATE TABLE cache
(
    key                  text NOT NULL,
    notification_service text NOT NULL,
    date                 text NOT NULL
);

CREATE UNIQUE INDEX cache_key_notification_service_uindex
    ON cache (key, notification_service);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE cache;
DROP INDEX cache_key_notification_service_uindex;
-- +goose StatementEnd
