CREATE TABLE user_details (
    username VARCHAR(64) PRIMARY KEY NOT NULL UNIQUE,
    name text NOT NULL
);

CREATE TABLE `google_creds` (
    google_id VARCHAR(24) PRIMARY KEY NOT NULL UNIQUE,
    username VARCHAR(64)NOT NULL UNIQUE, 
    email TEXT NOT NULL UNIQUE,
    picture TEXT,
    FOREIGN KEY (username) REFERENCES user_details (username)
    ON DELETE CASCADE ON UPDATE NO ACTION
);
