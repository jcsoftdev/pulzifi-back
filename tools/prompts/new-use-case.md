# Prompt: New Use Case

## When to Use

When adding a new use case to an existing module.

## Prompt

```
Create a new use case `{use_case_name}` in module `{module_name}`.

Follow the project's vertical slicing pattern:

1. Create `modules/{module_name}/application/{use_case_name}/` with:
   - handler.go — orchestration logic (call repos, apply business rules, return response)
   - request.go — input DTO with validation
   - response.go — output DTO
   - handler_test.go — unit test using in-memory repository

2. Package naming: directory `{use_case_name}` → package name without underscores
   (e.g., directory `create_check` → `package createcheck`)

3. The handler constructor should accept repository interfaces from domain/repositories/

4. Wire the handler into infrastructure/http/module.go with the appropriate HTTP method and route

5. Add tenant-awareness if the use case reads/writes tenant-scoped data

Business logic:
- {describe the business rules}

Input:
- {describe the expected request fields}

Output:
- {describe the expected response fields}
```
