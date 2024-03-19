CREATE TABLE IF NOT EXISTS users (
                                     "id" INT NOT NULL,
                                     "name" TEXT NOT NULL,
                                     "email" TEXT NOT NULL,
                                     "roles" json NOT NULL,

                                     PRIMARY KEY ("id")
);