-- migration user
CREATE USER potato WITH PASSWORD 'potato_pwd';
GRANT CREATE ON DATABASE potato TO potato;

-- app users
-- privs managed by migration scripts

-- api
CREATE USER reader WITH PASSWORD 'reader_pwd';
GRANT CONNECT ON DATABASE potato TO reader;

-- loader
CREATE USER loader WITH PASSWORD 'loader_pwd';
GRANT CONNECT ON DATABASE potato TO loader;


-- initial/gold schema
CREATE SCHEMA potato AUTHORIZATION potato;
