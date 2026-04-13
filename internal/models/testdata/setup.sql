CREATE TABLE snippets
(
    id      SERIAL          NOT NULL PRIMARY KEY,
    title   VARCHAR(100)    NOT NULL,
    content TEXT            NOT NULL,
    created TIMESTAMP       NOT NULL,
    expires TIMESTAMP       NOT NULL
);
CREATE INDEX idx_snippets_created ON snippets (created);

CREATE TABLE users
(
    id              SERIAL          NOT NULL PRIMARY KEY,
    name            VARCHAR(255)    NOT NULL,
    email           VARCHAR(255)    NOT NULL,
    hashed_password CHAR(60)        NOT NULL,
    created         TIMESTAMP       NOT NULL
);
ALTER TABLE users
    ADD CONSTRAINT users_uc_email UNIQUE (email);

INSERT INTO users (name, email, hashed_password, created)
VALUES ('Alice Jones',
        'alice@example.com',
        '$2a$12$RB1A2rdzmXNzt5uPOU9hWOgojGbYCWlV3jJj6474KhdRM947r/gwC',
        '2022-01-01 10:00:00');

CREATE TABLE link_mapping
(
    id            SERIAL          NOT NULL PRIMARY KEY,
    original_link VARCHAR(100)    NOT NULL UNIQUE,
    short_link    VARCHAR(100)    NOT NULL UNIQUE
);

INSERT INTO link_mapping (original_link, short_link)
VALUES ('https://existent.com', '123456');

CREATE TABLE files
(
    id           SERIAL              NOT NULL PRIMARY KEY,
    file_name    VARCHAR(100)        NOT NULL,
    file_uuid    VARCHAR(100) UNIQUE NOT NULL,
    file_size    INTEGER             NOT NULL,
    checksum     VARCHAR(100)        NOT NULL,
    storage_path VARCHAR(100)        NOT NULL,
    upload_date  TIMESTAMP           NOT NULL
);
CREATE INDEX idx_files_file_name ON files (file_name);

INSERT INTO files (file_name, file_uuid, file_size, checksum, storage_path, upload_date)
VALUES ('test_file.pdf',
        '123456',
        100,
        'abcdef',
        '/clonebox/upload',
        '2025-01-01 10:00:00');