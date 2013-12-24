-- +goose Up
CREATE TABLE test_kilometers (
    Id serial NOT NULL CONSTRAINT pm PRIMARY KEY,
    Date date CONSTRAINT unique_date UNIQUE,
    Begin int,
    Eerste int,
    Laatste int,
    Terug int,
    Comment varchar(200)
);

-- +goose Down
DROP TABLE kilometers;
