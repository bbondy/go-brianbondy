# sql

## Basics

Select all:  
`SELECT * FROM table;`

Select columns:  
`SELECT id, name FROM users;`

Filter rows:  
`SELECT * FROM users WHERE active = true;`

Limit and order:  
`SELECT * FROM users ORDER BY created_at DESC LIMIT 10;`

Distinct:  
`SELECT DISTINCT status FROM orders;`

## Aggregation

Count/sum/avg:  
`SELECT COUNT(*), SUM(total), AVG(total) FROM orders;`

Group by:  
`SELECT status, COUNT(*) FROM orders GROUP BY status;`

Having:  
`SELECT status, COUNT(*) FROM orders GROUP BY status HAVING COUNT(*) > 10;`

## Joins

Inner join:  
`SELECT * FROM a JOIN b ON a.id = b.a_id;`

Left join:  
`SELECT * FROM a LEFT JOIN b ON a.id = b.a_id;`

Full join:  
`SELECT * FROM a FULL OUTER JOIN b ON a.id = b.a_id;`

## Subqueries and CTEs

Subquery:  
`SELECT * FROM users WHERE id IN (SELECT user_id FROM orders);`

CTE:  
```sql
WITH recent AS (
  SELECT * FROM orders WHERE created_at > NOW() - INTERVAL '7 days'
)
SELECT * FROM recent;
```

Recursive CTE:  
```sql
WITH RECURSIVE tree AS (
  SELECT id, parent_id, name FROM nodes WHERE parent_id IS NULL
  UNION ALL
  SELECT n.id, n.parent_id, n.name
  FROM nodes n
  JOIN tree t ON n.parent_id = t.id
)
SELECT * FROM tree;
```

## Window functions

Row number:  
`SELECT *, ROW_NUMBER() OVER (PARTITION BY user_id ORDER BY created_at) FROM orders;`

Running total:  
`SELECT *, SUM(total) OVER (PARTITION BY user_id ORDER BY created_at) FROM orders;`

Lag/lead:  
`SELECT *, LAG(total) OVER (ORDER BY created_at) FROM orders;`

## Insert/update/delete

Insert:  
`INSERT INTO users (name, email) VALUES ('A', 'a@example.com');`

Insert from select:  
`INSERT INTO archive SELECT * FROM users WHERE deleted = true;`

Update:  
`UPDATE users SET active = false WHERE last_login < NOW() - INTERVAL '1 year';`

Delete:  
`DELETE FROM users WHERE id = 123;`

## Upsert

Postgres upsert:  
```sql
INSERT INTO users (id, email)
VALUES (1, 'a@example.com')
ON CONFLICT (id) DO UPDATE SET email = EXCLUDED.email;
```

## Transactions

Transaction block:  
```sql
BEGIN;
UPDATE accounts SET balance = balance - 100 WHERE id = 1;
UPDATE accounts SET balance = balance + 100 WHERE id = 2;
COMMIT;
```

Rollback:  
`ROLLBACK;`

## Indexes

Create index:  
`CREATE INDEX idx_users_email ON users(email);`

Composite index:  
`CREATE INDEX idx_orders_user_created ON orders(user_id, created_at);`

Unique index:  
`CREATE UNIQUE INDEX idx_users_username ON users(username);`

## Constraints

Primary key:  
`id INT PRIMARY KEY`

Foreign key:  
`user_id INT REFERENCES users(id)`

Check constraint:  
`CHECK (total >= 0)`

## Schema

Create table:  
```sql
CREATE TABLE users (
  id SERIAL PRIMARY KEY,
  email TEXT NOT NULL UNIQUE,
  created_at TIMESTAMP DEFAULT NOW()
);
```

Alter table:  
`ALTER TABLE users ADD COLUMN last_login TIMESTAMP;`

Drop table:  
`DROP TABLE users;`

## Views and materialized views

View:  
`CREATE VIEW active_users AS SELECT * FROM users WHERE active = true;`

Materialized view:  
`CREATE MATERIALIZED VIEW daily_totals AS SELECT date_trunc('day', created_at), SUM(total) FROM orders GROUP BY 1;`

Refresh MV:  
`REFRESH MATERIALIZED VIEW daily_totals;`

## JSON (Postgres)

JSON field:  
`SELECT data->>'name' FROM events;`

JSONB containment:  
`SELECT * FROM events WHERE data @> '{"type":"click"}';`

## Text search (Postgres)

TS vector:  
`SELECT to_tsvector('english', body) @@ plainto_tsquery('english', 'search');`

## Explain and analyze

Query plan:  
`EXPLAIN SELECT * FROM users WHERE email = 'a@example.com';`

Analyze:  
`EXPLAIN ANALYZE SELECT * FROM users WHERE email = 'a@example.com';`

## Permissions

Grant select:  
`GRANT SELECT ON users TO analyst;`

Revoke:  
`REVOKE ALL ON users FROM analyst;`

## Migrations

Add column with default (safe pattern):  
```sql
ALTER TABLE users ADD COLUMN status TEXT;
UPDATE users SET status = 'active' WHERE status IS NULL;
ALTER TABLE users ALTER COLUMN status SET NOT NULL;
```
