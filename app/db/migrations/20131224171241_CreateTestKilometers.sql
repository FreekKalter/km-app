-- +goose Up
CREATE TABLE test_kilometers (
    Id serial NOT NULL CONSTRAINT pmT PRIMARY KEY,
    Date date CONSTRAINT unique_dateT UNIQUE,
    Begin int,
    Eerste int,
    Laatste int,
    Terug int,
    Comment varchar(200)
);

-- +goose Down
DROP TABLE kilometers;
