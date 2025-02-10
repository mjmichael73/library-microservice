-- Seed genres
INSERT INTO bookservice.genres (genre_id, title, description)
VALUES
    (uuid_generate_v4(), 'Fiction', 'Fictional books including novels and stories'),
    (uuid_generate_v4(), 'Science Fiction', 'Books based on futuristic concepts and advanced technology'),
    (uuid_generate_v4(), 'Fantasy', 'Books involving magical elements and mythical creatures'),
    (uuid_generate_v4(), 'Non-Fiction', 'Books based on factual information and real events');

-- Seed authors
INSERT INTO bookservice.authors (author_id, name)
VALUES
    (uuid_generate_v4(), 'George Orwell'),
    (uuid_generate_v4(), 'Isaac Asimov'),
    (uuid_generate_v4(), 'J.K. Rowling'),
    (uuid_generate_v4(), 'Malcolm Gladwell');

-- Seed books
INSERT INTO bookservice.books (book_id, title, summary)
VALUES
    (uuid_generate_v4(), '1984', 'A dystopian novel set in a totalitarian society ruled by Big Brother.'),
    (uuid_generate_v4(), 'Foundation', 'A science fiction novel about the collapse and rebirth of a galactic empire.'),
    (uuid_generate_v4(), 'Harry Potter and the Sorcerer''s Stone', 'A fantasy novel about a young wizard''s adventures at Hogwarts.'),
    (uuid_generate_v4(), 'Outliers', 'A non-fiction book that explores the factors contributing to high levels of success.');

-- Seed books_authors (associating books with authors)
INSERT INTO bookservice.books_authors (book_id, author_id)
VALUES
    ((SELECT book_id FROM bookservice.books WHERE title = '1984'), (SELECT author_id FROM bookservice.authors WHERE name = 'George Orwell')),
    ((SELECT book_id FROM bookservice.books WHERE title = 'Foundation'), (SELECT author_id FROM bookservice.authors WHERE name = 'Isaac Asimov')),
    ((SELECT book_id FROM bookservice.books WHERE title = 'Harry Potter and the Sorcerer''s Stone'), (SELECT author_id FROM bookservice.authors WHERE name = 'J.K. Rowling')),
    ((SELECT book_id FROM bookservice.books WHERE title = 'Outliers'), (SELECT author_id FROM bookservice.authors WHERE name = 'Malcolm Gladwell'));

-- Seed books_genres (associating books with genres)
INSERT INTO bookservice.books_genres (book_id, genre_id)
VALUES
    ((SELECT book_id FROM bookservice.books WHERE title = '1984'), (SELECT genre_id FROM bookservice.genres WHERE title = 'Fiction')),
    ((SELECT book_id FROM bookservice.books WHERE title = 'Foundation'), (SELECT genre_id FROM bookservice.genres WHERE title = 'Science Fiction')),
    ((SELECT book_id FROM bookservice.books WHERE title = 'Harry Potter and the Sorcerer''s Stone'), (SELECT genre_id FROM bookservice.genres WHERE title = 'Fantasy')),
    ((SELECT book_id FROM bookservice.books WHERE title = 'Outliers'), (SELECT genre_id FROM bookservice.genres WHERE title = 'Non-Fiction'));
