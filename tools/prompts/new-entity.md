# Prompt: New Entity

## When to Use

When adding a new business entity to a module's domain layer.

## Prompt

```
Create a new entity `{EntityName}` in module `{module_name}`.

Follow the project's DDD conventions:

1. Create the entity struct in `modules/{module_name}/domain/entities/{entity_name}.go`
   - Use `uuid.UUID` for ID fields (github.com/google/uuid)
   - Use `time.Time` for timestamps
   - No external framework imports in domain layer
   - Add constructor function `New{EntityName}(...)` that generates UUID and sets CreatedAt

2. Create the repository interface in `modules/{module_name}/domain/repositories/{entity_name}_repository.go`
   - Define CRUD methods: Create, GetByID, List, Update, Delete
   - All methods take `context.Context` as first parameter
   - Return domain errors for not-found cases

3. Create the PostgreSQL implementation in `modules/{module_name}/infrastructure/persistence/{entity_name}_postgres.go`
   - Constructor: `NewPostgres{EntityName}Repository(db *sql.DB, tenant string)`
   - All queries prefixed with `middleware.GetSetSearchPathSQL(tenant)`
   - Use parameterized queries ($1, $2...) — never string concatenation

4. Create an in-memory implementation in `modules/{module_name}/infrastructure/persistence/{entity_name}_memory.go`
   - For unit testing without database dependency

Entity fields:
- {list the fields with types}

Table name: {table_name} (in tenant schema)
```
