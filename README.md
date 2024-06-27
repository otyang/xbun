# xbun: Database Library for Go

**xbun** is a Go library that simplifies interaction with relational databases. It provides a clean and concise API for performing common operations like creating, reading, updating, and deleting data. It leverages the powerful Bun library ([https://github.com/uptrace/bun](https://github.com/uptrace/bun)) underneath, offering a familiar interface with additional helper functions.

### Features

* Supports PostgreSQL and SQLite databases.
* Provides functions for:
    * Creating new records (with optional duplicate handling).
    * Deleting records by primary key or custom criteria.
    * Performing upserts (insert or update on conflict).
    * Updating records by primary key or custom criteria.
    * Finding single records by primary key or custom criteria.
    * Finding multiple records with pagination support.
    * Executing database transactions.
* Configurable query logging for debugging purposes.
* Automatic connection pooling management.

### Installation

**xbun** requires Go 1.18 or later. To install:

```bash
go get -u github.com/your-username/xbun
```

Replace `your-username` with your actual Go module path.

### Usage

**1. Connect to your database:**

```go
package main

import (
    "context"
    "fmt"

    "github.com/your-username/xbun"
)

func main() {
    db, err := xbun.NewDBConn("pg", "postgres://user:password@host:port/database", 10, true)
    if err != nil {
        panic(err)
    }
    defer db.Close()

    // Use the db connection throughout your application
}
```

**2. Define your database models:**

```go
type User struct {
    ID       int64  `bun:",pk"`
    Username string `bun:"unique"`
    Email    string
}
```

**3. Perform CRUD operations:**

```go
func CreateUser(ctx context.Context, db *xbun.DB, user *User) error {
    return xbun.Create(ctx, db, false, user) // Insert without ignoring duplicates
}

func FindUserByID(ctx context.Context, db *xbun.DB, id int64) (User, error) {
    var user User
    return xbun.FindOne(ctx, db, &user, func(q *bun.SelectQuery) *bun.SelectQuery {
        return q.Where("id = ?", id)
    })
}

func UpdateUserEmail(ctx context.Context, db *xbun.DB, id int64, newEmail string) error {
    return xbun.Update(ctx, db, func(q *bun.UpdateQuery) *bun.UpdateQuery {
        return q.Set("email = ?", newEmail).Where("id = ?", id)
    })
}

func DeleteUser(ctx context.Context, db *xbun.DB, id int64) (int64, error) {
    return xbun.Delete(ctx, db, &User{ID: id}, nil) // Delete by primary key
}
```

**4. Transactions:**

```go
func TransferFunds(ctx context.Context, db *xbun.DB, fromID, toID int64, amount float64) error {
    return xbun.Transaction(ctx, db, func(ctx context.Context, tx bun.Tx) error {
        // Perform operations within the transaction
        return nil
    })
}
```

**Please refer to the full source code for detailed documentation of all functions and available options.**
