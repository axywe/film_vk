DROP TABLE IF EXISTS actor_movie;
DROP TABLE IF EXISTS actors;
DROP TABLE IF EXISTS movies;
DROP TABLE IF EXISTS users;

CREATE TABLE IF NOT EXISTS actors (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    gender VARCHAR(50),
    birthdate DATE NOT NULL
);

CREATE TABLE IF NOT EXISTS movies (
    id SERIAL PRIMARY KEY,
    title VARCHAR(150) NOT NULL,
    description TEXT,
    release_date DATE NOT NULL,
    rating DECIMAL(3, 1) CHECK (rating >= 0 AND rating <= 10)
);

CREATE TABLE IF NOT EXISTS actor_movie (
    actor_id INT NOT NULL,
    movie_id INT NOT NULL,
    PRIMARY KEY (actor_id, movie_id),
    FOREIGN KEY (actor_id) REFERENCES actors (id) ON DELETE CASCADE,
    FOREIGN KEY (movie_id) REFERENCES movies (id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    username VARCHAR(50) NOT NULL,
    password VARCHAR(255) NOT NULL,
    role int NOT NULL
);


-- Test data
-- Login: admin, Password: admin
INSERT INTO users (username, password, role) VALUES ('admin', '$2a$10$NjIPpHePTDy5hJs/JmX90uWxWT5jOqrw0OyrBg88lmiQvlHQHbAXu', 1); 
-- Login: user, Password: user
INSERT INTO users (username, password, role) VALUES ('user', '$2a$10$ajvqHTuI3ixFdkI2WUJrF.KPPp2etsdgtj/jccMH0yek7W8JZK3P6', 2);