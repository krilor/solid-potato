# Thoughts and decisions

* Will run postgres since its boring and I know it.
* No point in adding a migration tool at this point. Will go for idempotent sql
  migrations and a just target.
* Use case seems very batch oriented. Old-school ELT should be good.
* Batch loading is considered to be out-of-scope, so I'll just do a simple COPY
  and merge.
