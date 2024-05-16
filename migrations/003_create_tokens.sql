CREATE TABLE google_tokens (
                               id SERIAL PRIMARY KEY,
                               user_id INT NOT NULL,
                               access_token TEXT NOT NULL,
                               refresh_token TEXT NOT NULL,
                               expires_at TIMESTAMP NOT NULL,
                               CONSTRAINT fk_user FOREIGN KEY (user_id) REFERENCES users(id),
                               UNIQUE (user_id)
);

CREATE TABLE tokens (
                        id SERIAL PRIMARY KEY,
                        user_id INT NOT NULL,
                        access_token TEXT NOT NULL,
                        refresh_token TEXT NOT NULL,
                        expires_at TIMESTAMP NOT NULL,
                        CONSTRAINT fk_user FOREIGN KEY (user_id) REFERENCES users(id),
                        UNIQUE (user_id)
);