CREATE SCHEMA IF NOT EXISTS loanservice;

CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE
    loanservice.borrows (
        genre_id UUID PRIMARY KEY,
        user_id UUID,
        book_id UUID,
        from_date TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
        to_date TIMESTAMP,
        created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
        updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
    );