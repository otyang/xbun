# xbun  

A collection of utility functions to streamline common database operations using the Bun: [[https://bun.uptrace.dev/](https://bun.uptrace.dev/)] SQL query builder for Go.

## Features

* **Simplified CRUD:** 
   - `Create` functions with optional duplicate handling
   - `CreateBulk` for inserting multiple records efficiently
   - Easy retrieval with `SelectOneByPK`, `SelectOneWhere`, and `SelectManyWhere`
   - Flexible updating with `UpdateOneByPK`, `UpdateManyByPK`, and `UpdateOneWhere`
   - Convenient deletion using `DeleteByPK` and `DeleteWhere`

* **Upsert:** Effortlessly handle insert or update actions in a single operation.

* **Transactions:** Use `Transaction` for atomic database updates.

 
## Quick Example

```go
package main

import (
  "context"
  "github.com/uptrace/bun"
  "github.com/otyang/xbun" 
)

type User struct {
  bun.BaseModel `bun:"table:users"` // Assuming you use Bun's BaseModel
  ID            int64             `bun:"id,pk,autoincrement"`
  Name          string            `bun:"name"`
}

func main() {
  // ... Connect to your database using Bun.

  ctx := context.Background()

  // Create a new user
  user := &User{Name: "John Doe"}
  err := xbun.Create(ctx, db, user, true) // Ignore duplicates
  if err != nil {
    // ... Handle error
  }

  // Update an existing user
  user.Name = "Jane Doe"
  err = xbun.UpdateOneByPK(ctx, db, user)
  if err != nil {
    // ... Handle error 
  }
}
```

For detailed descriptions and usage notes of each function, please refer unit test or package documentation on go pkg website.

## Contributing

We welcome contributions! To submit changes or report issues:

1. Fork the repository.
2. Create a branch for your changes.
3. Open a pull request with a clear explanation.

## License

This project is licensed under the MIT License - see the LICENSE file for details.
 