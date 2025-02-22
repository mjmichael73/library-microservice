-- Seed genres
INSERT INTO userservice.users (user_id, first_name, last_name, email, is_admin, password)
VALUES
    (uuid_generate_v4(), 'Mojtaba', 'Michael', 'mojimich2015@gmail.com', true, '$2a$10$5bYWuC276Uog3W4b/dzdOudDJSZ3B.NulQziDrnBVTWNL51JH3gtC'),
    (uuid_generate_v4(), 'John', 'Doe', 'johndoe@example.com', false, '$2a$10$5bYWuC276Uog3W4b/dzdOudDJSZ3B.NulQziDrnBVTWNL51JH3gtC');