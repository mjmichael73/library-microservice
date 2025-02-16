CREATE SCHEMA IF NOT EXISTS loanservice;

CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE
    loanservice.borrows (
        borrow_id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
        user_id UUID,
        book_id UUID,
        from_date TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
        to_date TIMESTAMP,
        returned_date TIMESTAMP,
        status VARCHAR(20) DEFAULT 'active',
        remarks TEXT,
        created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
        updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
    );

CREATE INDEX idx_user_id ON loanservice.borrows(user_id);
CREATE INDEX idx_book_id ON loanservice.borrows(book_id);
CREATE INDEX idx_from_date ON loanservice.borrows(from_date);
CREATE INDEX idx_to_date ON loanservice.borrows(to_date);
CREATE INDEX idx_status ON loanservice.borrows(status);