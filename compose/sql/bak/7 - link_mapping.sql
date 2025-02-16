CREATE TABLE link_mapping
(
    id            INTEGER      NOT NULL PRIMARY KEY AUTO_INCREMENT,
    original_link VARCHAR(100) NOT NULL UNIQUE,
    short_link    VARCHAR(100) NOT NULL UNIQUE
);

