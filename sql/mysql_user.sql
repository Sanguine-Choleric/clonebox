CREATE USER 'web'@'localhost';
GRANT SELECT, INSERT, UPDATE, DELETE ON snippetbox.* TO 'web'@'%';
ALTER USER 'web'@'localhost' IDENTIFIED BY 'pass';