#!/bin/bash
# Creates all 5 databases for the SQLi lab
set -e

PGPASSWORD=sqlipass psql -v ON_ERROR_STOP=1 --username "sqliuser" --dbname "postgres" <<-EOSQL
    CREATE DATABASE app1_db;
    CREATE DATABASE app2_db;
    CREATE DATABASE app3_db;
    CREATE DATABASE app4_db;
    CREATE DATABASE app5_db;
EOSQL

# ─── App1: ShopFlow ────────────────────────────────────────────────────────────
PGPASSWORD=sqlipass psql -v ON_ERROR_STOP=1 --username "sqliuser" --dbname "app1_db" <<-'EOSQL'
CREATE TABLE categories (
    id   SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL
);

CREATE TABLE products (
    id          SERIAL PRIMARY KEY,
    name        VARCHAR(200) NOT NULL,
    description TEXT,
    price       NUMERIC(10,2) NOT NULL,
    category    VARCHAR(100),
    released    INTEGER DEFAULT 1,
    stock       INTEGER DEFAULT 100
);

CREATE TABLE users (
    id       SERIAL PRIMARY KEY,
    username VARCHAR(100) UNIQUE NOT NULL,
    password VARCHAR(255) NOT NULL,
    email    VARCHAR(200),
    role     VARCHAR(50) DEFAULT 'customer'
);

CREATE TABLE promo_codes (
    id       SERIAL PRIMARY KEY,
    code     VARCHAR(50) UNIQUE NOT NULL,
    discount INTEGER,
    active   BOOLEAN DEFAULT true
);

INSERT INTO categories VALUES (1,'Electronics'),(2,'Clothing'),(3,'Books'),(4,'Sports');

INSERT INTO products(name,description,price,category,released,stock) VALUES
  ('Laptop Pro 15','High-performance laptop for professionals',1299.99,'Electronics',1,50),
  ('Wireless Headphones','Active noise-cancelling over-ear headphones',249.99,'Electronics',1,120),
  ('Smart Watch','Fitness tracking smartwatch with GPS',399.99,'Electronics',1,80),
  ('Running Shoes','Lightweight carbon-plate racing shoes',129.99,'Sports',1,200),
  ('Yoga Mat','6mm non-slip natural rubber mat',49.99,'Sports',1,300),
  ('Best Sellers Bundle','Collection of top 10 novels',79.99,'Books',1,150),
  ('Summer Dress','Lightweight linen summer dress',59.99,'Clothing',1,100),
  ('SECRET_PROTOTYPE_X9','[CLASSIFIED] Pre-release device — internal only',9999.99,'Electronics',0,1),
  ('INTERNAL_VOUCHER_MASTER','All-access internal voucher',0.00,'Internal',0,1);

INSERT INTO users(username,password,email,role) VALUES
  ('admin','Sup3r$ecret!2024','admin@shopflow.local','admin'),
  ('alice','alice_pw_2024','alice@example.com','customer'),
  ('bob','b0bR0cks!','bob@example.com','customer');

INSERT INTO promo_codes(code,discount,active) VALUES
  ('SUMMER20',20,true),
  ('WELCOME10',10,true),
  ('BLACKFRIDAY50',50,true),
  ('ADMIN_BACKDOOR_9999',100,false);
EOSQL

# ─── App2: LibroBase ───────────────────────────────────────────────────────────
PGPASSWORD=sqlipass psql -v ON_ERROR_STOP=1 --username "sqliuser" --dbname "app2_db" <<-'EOSQL'
CREATE TABLE books (
    id        SERIAL PRIMARY KEY,
    title     VARCHAR(300) NOT NULL,
    author    VARCHAR(200),
    genre     VARCHAR(100),
    isbn      VARCHAR(20),
    year      INTEGER,
    available BOOLEAN DEFAULT true
);

CREATE TABLE members (
    id         SERIAL PRIMARY KEY,
    name       VARCHAR(200) NOT NULL,
    email      VARCHAR(200),
    membership VARCHAR(50) DEFAULT 'standard',
    pin        VARCHAR(10)
);

CREATE TABLE staff (
    id       SERIAL PRIMARY KEY,
    username VARCHAR(100) UNIQUE NOT NULL,
    password VARCHAR(255) NOT NULL,
    role     VARCHAR(50)
);

CREATE TABLE secret_data (
    id    SERIAL PRIMARY KEY,
    label VARCHAR(100),
    value TEXT
);

INSERT INTO books(title,author,genre,isbn,year) VALUES
  ('Clean Code','Robert C. Martin','Technology','978-0132350884',2008),
  ('The Pragmatic Programmer','Andrew Hunt','Technology','978-0135957059',2019),
  ('Dune','Frank Herbert','Sci-Fi','978-0441013593',1965),
  ('1984','George Orwell','Dystopia','978-0451524935',1949),
  ('Sapiens','Yuval Noah Harari','History','978-0062316097',2011),
  ('The Great Gatsby','F. Scott Fitzgerald','Fiction','978-0743273565',1925),
  ('Deep Work','Cal Newport','Self-Help','978-1455586691',2016);

INSERT INTO members(name,email,membership,pin) VALUES
  ('Alice Ferreira','alice@libra.local','premium','7842'),
  ('Carlos Mendes','carlos@example.com','standard','1193'),
  ('Diana Costa','diana@example.com','premium','5501'),
  ('Admin Account','admin@libra.local','staff','0000');

INSERT INTO staff(username,password,role) VALUES
  ('librarian','l1br@r14n2024','staff'),
  ('admin','Adm1n_S3cret!','admin');

INSERT INTO secret_data(label,value) VALUES
  ('db_master_pass','pg_sup3r_s3cr3t_2024'),
  ('api_key','sk-libra-f3a9d812c045b678'),
  ('backup_schedule','daily 03:00 UTC');
EOSQL

# ─── App3: NewsHub ─────────────────────────────────────────────────────────────
PGPASSWORD=sqlipass psql -v ON_ERROR_STOP=1 --username "sqliuser" --dbname "app3_db" <<-'EOSQL'
CREATE TABLE articles (
    id        SERIAL PRIMARY KEY,
    title     VARCHAR(500) NOT NULL,
    content   TEXT,
    category  VARCHAR(100),
    author    VARCHAR(200),
    published BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE TABLE users (
    id           SERIAL PRIMARY KEY,
    username     VARCHAR(100) UNIQUE NOT NULL,
    password     VARCHAR(255) NOT NULL,
    email        VARCHAR(200),
    session_token VARCHAR(64)
);

CREATE TABLE tags (
    id   SERIAL PRIMARY KEY,
    name VARCHAR(100) UNIQUE NOT NULL
);

CREATE TABLE article_tags (
    article_id INTEGER REFERENCES articles(id),
    tag_id     INTEGER REFERENCES tags(id),
    PRIMARY KEY (article_id, tag_id)
);

CREATE TABLE comments (
    id         SERIAL PRIMARY KEY,
    article_id INTEGER,
    author     VARCHAR(200),
    content    TEXT,
    approved   BOOLEAN DEFAULT false,
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE TABLE admin_notes (
    id      SERIAL PRIMARY KEY,
    note    TEXT,
    created_at TIMESTAMP DEFAULT NOW()
);

INSERT INTO articles(title,content,category,author) VALUES
  ('Local Elections: What You Need to Know','Detailed breakdown of upcoming municipal elections...','Politics','Maria Silva'),
  ('Tech Giants Report Record Profits','Apple, Google and Microsoft all posted record Q2...','Technology','João Ramos'),
  ('Climate Summit: Key Agreements Reached','World leaders agreed on new carbon reduction targets...','Environment','Ana Oliveira'),
  ('Startup Scene Booms in São Paulo','Over 400 new startups registered in Q1 this year...','Business','Carlos Lima'),
  ('New Study Links Sleep to Productivity','Researchers at USP found that 8h sleep improves...','Health','Dr. Fátima Santos');

INSERT INTO users(username,password,email,session_token) VALUES
  ('editor','Ed1t0r_2024!','editor@newshub.local','tok_a1b2c3d4e5f6'),
  ('admin','N3wsHub_Adm1n!','admin@newshub.local','tok_x9y8z7w6v5u4'),
  ('reporter1','Rep0rter!23','reporter@newshub.local','tok_m3n4o5p6q7r8');

INSERT INTO tags(name) VALUES ('breaking'),('politics'),('tech'),('environment'),('health'),('business');

INSERT INTO article_tags VALUES (1,2),(2,3),(3,4),(4,6),(5,5);

INSERT INTO admin_notes(note) VALUES
  ('DB backup creds: pg_pass=Backupx9#2024'),
  ('Admin panel at /internal/admin - key: nh_adm_7f3a9'),
  ('Reporter API token: nr_api_d812c045b678f3a9');
EOSQL

# ─── App4: StaffPortal ─────────────────────────────────────────────────────────
PGPASSWORD=sqlipass psql -v ON_ERROR_STOP=1 --username "sqliuser" --dbname "app4_db" <<-'EOSQL'
CREATE TABLE departments (
    id      SERIAL PRIMARY KEY,
    name    VARCHAR(100) NOT NULL,
    manager VARCHAR(200),
    budget  NUMERIC(12,2)
);

CREATE TABLE employees (
    id         SERIAL PRIMARY KEY,
    name       VARCHAR(200) NOT NULL,
    department VARCHAR(100),
    title      VARCHAR(200),
    email      VARCHAR(200),
    salary     NUMERIC(10,2),
    hire_date  DATE
);

CREATE TABLE credentials (
    id          SERIAL PRIMARY KEY,
    username    VARCHAR(100) UNIQUE NOT NULL,
    password    VARCHAR(255) NOT NULL,
    employee_id INTEGER REFERENCES employees(id),
    is_admin    BOOLEAN DEFAULT false
);

CREATE TABLE oob_log (
    id      SERIAL PRIMARY KEY,
    payload TEXT,
    source  VARCHAR(100),
    ts      TIMESTAMP DEFAULT NOW()
);

CREATE OR REPLACE FUNCTION oob_exfil(data TEXT, src TEXT DEFAULT 'unknown') RETURNS TEXT AS $$
BEGIN
    INSERT INTO oob_log(payload, source) VALUES(data, src);
    RETURN data;
END;
$$ LANGUAGE plpgsql;

INSERT INTO departments(name,manager,budget) VALUES
  ('Engineering','Lucas Martins',850000.00),
  ('Finance','Paula Souza',320000.00),
  ('Marketing','Roberto Alves',450000.00),
  ('HR','Fernanda Lima',280000.00);

INSERT INTO employees(name,department,title,email,salary,hire_date) VALUES
  ('Lucas Martins','Engineering','CTO','lucas@staffportal.local',18500.00,'2019-03-15'),
  ('Paula Souza','Finance','CFO','paula@staffportal.local',17200.00,'2020-01-10'),
  ('Ana Rocha','Engineering','Senior Developer','ana@staffportal.local',12000.00,'2021-06-01'),
  ('Carlos Brito','Marketing','Marketing Lead','carlos@staffportal.local',9800.00,'2022-02-14'),
  ('Juliana Pires','HR','HR Manager','juliana@staffportal.local',8500.00,'2021-11-20'),
  ('Diego Nascimento','Engineering','DevOps Engineer','diego@staffportal.local',11000.00,'2022-08-08');

INSERT INTO credentials(username,password,employee_id,is_admin) VALUES
  ('admin','Sp0rt@l_Adm1n#2024',1,true),
  ('lucas.martins','L!CAS_cto_9x2024',1,true),
  ('paula.souza','P@ul4_Fin2024!',2,false),
  ('ana.rocha','An@_D3v_secur3!',3,false);
EOSQL

# ─── App5: StockTrack ──────────────────────────────────────────────────────────
PGPASSWORD=sqlipass psql -v ON_ERROR_STOP=1 --username "sqliuser" --dbname "app5_db" <<-'EOSQL'
CREATE TABLE inventory (
    id       SERIAL PRIMARY KEY,
    sku      VARCHAR(50) UNIQUE NOT NULL,
    name     VARCHAR(200) NOT NULL,
    quantity INTEGER DEFAULT 0,
    location VARCHAR(100),
    price    NUMERIC(10,2)
);

CREATE TABLE suppliers (
    id      SERIAL PRIMARY KEY,
    name    VARCHAR(200) NOT NULL,
    contact VARCHAR(200),
    api_key VARCHAR(64)
);

CREATE TABLE orders (
    id         SERIAL PRIMARY KEY,
    sku        VARCHAR(50),
    quantity   INTEGER,
    status     VARCHAR(50) DEFAULT 'pending',
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE TABLE oob_log (
    id      SERIAL PRIMARY KEY,
    payload TEXT,
    source  VARCHAR(100),
    ts      TIMESTAMP DEFAULT NOW()
);

CREATE OR REPLACE FUNCTION oob_exfil(data TEXT, src TEXT DEFAULT 'unknown') RETURNS TEXT AS $$
BEGIN
    INSERT INTO oob_log(payload, source) VALUES(data, src);
    RETURN data;
END;
$$ LANGUAGE plpgsql;

INSERT INTO inventory(sku,name,quantity,location,price) VALUES
  ('SKU-001','Industrial Pump A2',150,'Warehouse A - Row 3',2499.00),
  ('SKU-002','Conveyor Belt Module',80,'Warehouse B - Row 1',8750.00),
  ('SKU-003','Electric Motor 5kW',200,'Warehouse A - Row 7',1350.00),
  ('SKU-004','Safety Valve SV-300',500,'Warehouse C - Row 2',320.00),
  ('SKU-005','Pressure Sensor PS-100',1200,'Warehouse B - Row 5',89.00),
  ('SKU-006','Control Panel CP-X',45,'Warehouse A - Row 1',15200.00);

INSERT INTO suppliers(name,contact,api_key) VALUES
  ('TechParts Ltda','techparts@supplier.local','sup_api_a1b2c3d4e5f6g7h8'),
  ('Industrial Corp','corp@industrial.local','sup_api_x9y8z7w6v5u4t3'),
  ('SafeEquip Brasil','safe@equip.local','sup_api_q1w2e3r4t5y6u7');

INSERT INTO orders(sku,quantity,status) VALUES
  ('SKU-001',10,'delivered'),
  ('SKU-003',25,'shipped'),
  ('SKU-005',100,'pending');
EOSQL

echo "All databases initialized successfully."
