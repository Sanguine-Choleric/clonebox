CREATE TABLE snippets
(
    id      INTEGER      NOT NULL PRIMARY KEY AUTO_INCREMENT,
    title   VARCHAR(100) NOT NULL,
    content TEXT         NOT NULL,
    created DATETIME     NOT NULL,
    expires DATETIME     NOT NULL
);
CREATE INDEX idx_snippets_created ON snippets (created);
CREATE TABLE users
(
    id              INTEGER      NOT NULL PRIMARY KEY AUTO_INCREMENT,
    name            VARCHAR(255) NOT NULL,
    email           VARCHAR(255) NOT NULL,
    hashed_password CHAR(60)     NOT NULL,
    created         DATETIME     NOT NULL
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
    id            INTEGER      NOT NULL PRIMARY KEY AUTO_INCREMENT,
    original_link VARCHAR(100) NOT NULL UNIQUE,
    short_link    VARCHAR(100) NOT NULL UNIQUE
);

INSERT INTO link_mapping (original_link, short_link)
VALUES ('https://existent.com',
        '123456');