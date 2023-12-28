CREATE DATABASE payroll;

\connect payroll;

CREATE TYPE jobgroup AS ENUM ('A', 'B');

CREATE TABLE IF NOT EXISTS jobgroup_rate (
    job_group jobgroup PRIMARY KEY,
    rate FLOAT NOT NULL
);

CREATE TABLE IF NOT EXISTS processed_files (
    id INTEGER UNIQUE NOT NULL
);

CREATE TABLE IF NOT EXISTS worklog (
    id BIGSERIAL PRIMARY KEY,
    employee_id INTEGER NOT NULL,
    log_date TIMESTAMP NOT NULL,
    log_hours FLOAT DEFAULT 0.0,
    job_group jobgroup NOT NULL,
    updated_ts TIMESTAMP WITH TIME ZONE NOT NULL
);

INSERT INTO jobgroup_rate(job_group, rate) VALUES ('A', 20), ('B', 30);