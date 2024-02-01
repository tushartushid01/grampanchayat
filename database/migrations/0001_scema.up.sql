create type status_type as enum('new', 'processing', 'completed');


create type gender_type as enum ('male', 'female', 'other') ;


CREATE TABLE IF NOT EXISTS address (
                                       id serial primary key,
                                       address text,
                                       created_at timestamp with time zone default now(),
                                       updated_at timestamp with time zone,
                                       archived_at timestamp with time zone

);
-- create type role_type as enum('dm','sdm', 'secretary', 'labour_department', 'insurance_department');

CREATE TABLE IF NOT EXISTS roles(
                                    id SERIAL PRIMARY KEY ,
                                    role TEXT ,
                                    created_at TIMESTAMP WITH TIME ZONE DEFAULT now() NOT NULL ,
                                    updated_at TIMESTAMP WITH TIME ZONE ,
                                    archived_at TIMESTAMP WITH TIME ZONE
);

CREATE TABLE IF NOT EXISTS tehsil(
                                     id SERIAL PRIMARY KEY ,
                                     name TEXT NOT NULL ,
                                     created_at TIMESTAMP WITH TIME ZONE DEFAULT now() NOT NULL ,
                                     updated_at TIMESTAMP WITH TIME ZONE  ,
                                     archived_at TIMESTAMP WITH TIME ZONE
);


CREATE TABLE IF NOT EXISTS tehsil_address(
                                             id serial primary key,
                                             tehsil INTEGER references tehsil(id) not null ,
                                             address_id INTEGER references address(id) not null,
                                             created_at timestamp with time zone default now(),
                                             updated_at timestamp with time zone,
                                             archived_at timestamp with time zone
);


CREATE TABLE IF NOT EXISTS gram_panchayat(
                                             id SERIAL PRIMARY KEY ,
                                             name TEXT NOT NULL ,
                                             tehsil_id INTEGER REFERENCES tehsil(id) not null ,
                                             created_at TIMESTAMP WITH TIME ZONE DEFAULT now() NOT NULL ,
                                             updated_at TIMESTAMP WITH TIME ZONE  ,
                                             archived_at TIMESTAMP WITH TIME ZONE
);



CREATE TABLE IF NOT EXISTS users(
                                    id SERIAL PRIMARY KEY ,
                                    name TEXT NOT NULL ,
                                    email TEXT UNIQUE CHECK (email <>'') NOT NULL,
                                    phone_no TEXT UNIQUE NOT NULL ,
                                    date_of_birth date not null ,
                                    gender   gender_type,
                                    roles_id INTEGER REFERENCES roles(id),
                                    created_at TIMESTAMP WITH TIME ZONE DEFAULT now() NOT NULL ,
                                    updated_at TIMESTAMP WITH TIME ZONE ,
                                    archived_at TIMESTAMP WITH TIME ZONE
);

CREATE TABLE IF NOT EXISTS death_details(
                                            id SERIAL PRIMARY KEY ,
                                            name TEXT NOT NULL ,
                                            phone_no TEXT ,
                                            age      INTEGER NOT NULL ,
                                            date_of_birth DATE NOT NULL,
                                            gender   gender_type NOT NULL ,
                                            aadhar_number TEXT ,
                                            status status_type NOT NULL ,
                                            created_by INTEGER REFERENCES users(id),
                                            date_of_death TIMESTAMP WITH TIME ZONE,
                                            ended_on TIMESTAMP WITH TIME ZONE,
                                            created_at TIMESTAMP WITH TIME ZONE DEFAULT now() NOT NULL ,
                                            updated_at TIMESTAMP WITH TIME ZONE ,
                                            archived_at TIMESTAMP WITH TIME ZONE
);


CREATE TABLE IF NOT EXISTS death_details_address(
                                                    id serial primary key,
                                                    death_detail_id INTEGER references death_details(id) not null ,
                                                    address_id INTEGER references address(id) not null,
                                                    created_at timestamp with time zone default now(),
                                                    updated_at timestamp with time zone,
                                                    archived_at timestamp with time zone
);


CREATE TABLE IF NOT EXISTS user_tehsil(
                                          id serial primary key,
                                          tehsil_id INTEGER references tehsil(id) not null ,
                                          user_id INTEGER references users(id) not null,
                                          created_at timestamp with time zone default now(),
                                          updated_at timestamp with time zone,
                                          archived_at timestamp with time zone
);


CREATE TABLE IF NOT EXISTS user_gram_panchayat(
                                                  id serial primary key,
                                                  gram_panchayat_id INTEGER references gram_panchayat(id) not null ,
                                                  user_id INTEGER references users(id) not null,
                                                  created_at timestamp with time zone default now(),
                                                  updated_at timestamp with time zone,
                                                  archived_at timestamp with time zone
);



-- create type task_type as enum('victim_compensation', 'death_certificate', 'insurance_policy', 'family_registration');

CREATE TABLE IF NOT EXISTS task(
                                   id SERIAL PRIMARY KEY ,
                                   name TEXT,
                                   death_id INTEGER REFERENCES death_details(id),
                                   user_id INTEGER REFERENCES users(id),
                                   status  status_type,
                                   is_rejected bool,
                                   reason TEXT,
                                   start_date TIMESTAMP WITH TIME ZONE ,
                                   completed_date TIMESTAMP WITH TIME ZONE,
                                   created_at TIMESTAMP WITH TIME ZONE DEFAULT now() NOT NULL ,
                                   updated_at TIMESTAMP WITH TIME ZONE ,
                                   archived_at TIMESTAMP WITH TIME ZONE
);


CREATE TABLE IF NOT EXISTS sessions(
                                       id uuid primary key default gen_random_uuid() not null ,
                                       user_id INTEGER REFERENCES users(id) NOT NULL ,
                                       created_at TIMESTAMP WITH TIME ZONE DEFAULT now() NOT NULL ,
                                       updated_at TIMESTAMP WITH TIME ZONE  ,
                                       expires_at TIMESTAMP WITH TIME ZONE
);


CREATE TABLE IF NOT EXISTS roles_task_relation(
                                                  id SERIAL PRIMARY KEY ,
                                                  role_id integer references roles(id),
                                                  task text ,
                                                  created_at TIMESTAMP WITH TIME ZONE DEFAULT now() NOT NULL ,
                                                  updated_at TIMESTAMP WITH TIME ZONE ,
                                                  archived_at TIMESTAMP WITH TIME ZONE
);


CREATE TABLE IF NOT EXISTS otp(
                                  id SERIAL PRIMARY KEY ,
                                  user_id INTEGER REFERENCES users(id),
                                  otp INTEGER NOT NULL ,
                                  expiring_time TIMESTAMP WITH TIME ZONE,
                                  created_at TIMESTAMP WITH TIME ZONE DEFAULT now() NOT NULL ,
                                  archived_at TIMESTAMP WITH TIME ZONE
);

