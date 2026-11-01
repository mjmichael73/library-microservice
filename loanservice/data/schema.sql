CREATE SCHEMA IF NOT EXISTS loanservice;

CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- R-024: no FK constraints on user_id/book_id.
-- Cross-database FKs are impossible: users live in userservice's database and
-- books in bookservice's. Ownership is enforced at the application boundary in
-- loanservice/internal/server/borrow.go (user existence/auth via userservice,
-- book existence via bookservice) before a borrow is persisted.
CREATE TABLE
    loanservice.borrows (
        borrow_id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
        user_id UUID NOT NULL,
        book_id UUID NOT NULL,
        from_date TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
        to_date TIMESTAMP,
        returned_date TIMESTAMP,
        status VARCHAR(20) DEFAULT 'active' NOT NULL,
        remarks TEXT,
        created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
        updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
        CONSTRAINT chk_borrows_status CHECK (status IN ('active', 'returned', 'overdue'))
    );

CREATE INDEX idx_user_id ON loanservice.borrows(user_id);
CREATE INDEX idx_book_id ON loanservice.borrows(book_id);
CREATE INDEX idx_from_date ON loanservice.borrows(from_date);
CREATE INDEX idx_to_date ON loanservice.borrows(to_date);
CREATE INDEX idx_status ON loanservice.borrows(status);