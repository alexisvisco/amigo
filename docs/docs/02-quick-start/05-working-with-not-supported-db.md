# Working with not supported database

If you want to use a database that is not supported by amigo, you can. 

If the DSN is not recognized you will be using the base interface which is `base.Schema` (it only implement methods to manipulate the versions table) but you have access to the `*sql.DB`. 