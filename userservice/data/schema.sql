CREATE SCHEMA IF NOT EXISTS userservice;

CREATE TABLE
    userservice.users (
        user_id UUID PRIMARY KEY,
        first_name VARCHAR,
        last_name VARCHAR,
        email VARCHAR UNIQUE NOT NULL,
        is_admin BOOLEAN DEFAULT FALSE,
        password VARCHAR
    );