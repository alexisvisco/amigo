# Others operations

They are functions that are not in the constructive, destructive, or informative categories.

- [Exec(query string, args ...interface{})](https://pkg.go.dev/github.com/alexisvisco/mig/pkg/schema/pg#Schema.Exec)

If you want to reverse a query in a `change` function you should use the Reversible method.


```go
s.Reversible(schema.Directions{
    Up: func() {
        s.Exec("INSERT INTO public.mig_schema_versions (id) VALUES ('1')")
    },
    
    Down: func() { 
        s.Exec("DELETE FROM public.mig_schema_versions WHERE id = '1'")
    },
})
```
