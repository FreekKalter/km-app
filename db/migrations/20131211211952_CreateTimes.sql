-- +goose Up
CREATE TABLE times (
    Id serial NOT NULL,
    Date        date,
    CheckIn     integer,
    CheckOut    integer,
    PRIMARY KEY(Id)
);

-- +goose Down
DROP TABLE times;
