CREATE TABLE IF NOT EXISTS users (
                                     "id" INT NOT NULL,
                                     "name" TEXT NOT NULL,
                                     "email" TEXT NOT NULL,
                                     "roles" json NOT NULL,

                                     PRIMARY KEY ("id")
);

INSERT INTO users (id, name, email, roles) VALUES (65, 'Gopher', 'hello@gopher.com', '{"Finance Reporting":"Finance Reporting","Manager":"Manager"}');
