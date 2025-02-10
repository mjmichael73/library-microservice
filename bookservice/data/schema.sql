CREATE SCHEMA IF NOT EXISTS bookservice;
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE bookservice.genres (
    genre_id UUID PRIMARY KEY,
    title VARCHAR NOT NULL,
    description VARCHAR,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE bookservice.authors (
    author_id UUID PRIMARY KEY,
    name VARCHAR NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE bookservice.books (
    book_id UUID PRIMARY KEY,
    title VARCHAR NOT NULL,
    summary VARCHAR,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE bookservice.books_authors (
    book_id UUID REFERENCES bookservice.books(book_id) ON DELETE CASCADE,
    author_id UUID REFERENCES bookservice.authors(author_id) ON DELETE CASCADE,
    PRIMARY KEY (book_id, author_id)
);

CREATE TABLE bookservice.books_genres (
    book_id UUID REFERENCES bookservice.books(book_id) ON DELETE CASCADE,
    genre_id UUID REFERENCES bookservice.genres(genre_id) ON DELETE CASCADE,
    PRIMARY KEY (book_id, genre_id)
);
