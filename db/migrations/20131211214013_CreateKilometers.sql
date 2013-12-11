-- +goose Up
CREATE TABLE kilometers (
    Id serial NOT NULL,
    Date date,
    Begin int,
    Eerste int,
    Laatste int,
    Terug int,
    Comment varchar(200),
    PRIMARY KEY(Id)
);

-- +goose Down
DROP TABLE kilometers;
