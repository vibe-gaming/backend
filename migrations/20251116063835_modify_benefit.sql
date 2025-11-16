-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';
-- +goose StatementEnd

alter table benefit
    modify valid_from datetime null comment 'Дата начала акции';

alter table benefit
    modify valid_to datetime null comment 'Дата окончания акции';



-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
-- +goose StatementEnd

alter table benefit
    modify valid_from datetime not null comment 'Дата начала акции';

alter table benefit
    modify valid_to datetime not null comment 'Дата окончания акции';