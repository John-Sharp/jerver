CREATE ROLE jerver WITH LOGIN;

GRANT ALL ON DATABASE jerver TO jerver;

CREATE TABLE users (
   Uuid uuid NOT NULL PRIMARY KEY, 
   FirstName text,
   SecondName text,
   Username text,
   HashedPwd bytea);


CREATE TABLE threads (
   Uuid uuid NOT NULL PRIMARY KEY, 
   Title text);

CREATE TABLE messages (
   Uuid uuid NOT NULL PRIMARY KEY,
   ThreadId uuid NOT NULL,
   AuthorId uuid NOT NULL,
   Content text,
   FOREIGN KEY(ThreadId) REFERENCES threads(Uuid),
   FOREIGN KEY(AuthorId) REFERENCES users(Uuid));

GRANT SELECT, INSERT, UPDATE, DELETE
ON ALL TABLES IN SCHEMA public 
TO jerver;
